> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# Source to Deployment

This tutorial will show you how to clone a Github source repository, build/push container image from the source and finally deploy it. In this example we are going to use [kaniko](https://github.com/GoogleContainerTools/kaniko) for building container image. 

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

## Create Secret

In order to push image into container registry, you have to configure credential-initializer. First, create a secret with docker credentials and proper annotations. Also, you need to specify this secret in workflow's service-account in order to configure credential-initializer.

```console
$ kubectl apply -f ./docs/examples/deploy/secret.yaml
secret/docker-credential created
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: docker-credential
  annotations:
    credential.kube.ci/docker-0: https://index.docker.io/v1/
type: kubernetes.io/basic-auth
stringData:
  username: ...
  password: ...
```

## Configure RBAC

You need to specify a service-account in `spec.serviceAccount` to ensure RBAC for the workflow. This service-account along with operator's service-account must have `list` and `watch` permissions for the resources specified in `spec.triggers`.

Now, create a service-account for the workflow and specify previously created secret.

```yaml
# service-account for workflow
apiVersion: v1
kind: ServiceAccount
metadata:
  name: wf-sa
  namespace: default
secrets:
- name: docker-credential
```

Then, create a cluster-role with ConfigMap `list` and `watch` permissions. Also, you need to set necessary RBAC permissions to deploy your application (Deployment `get` and `create` permissions in this example). Now, bind it with service-accounts of both workflow and operator.

```console
$ kubectl apply -f ./docs/examples/deploy/rbac.yaml
serviceaccount/wf-sa created
clusterrole.rbac.authorization.k8s.io/wf-role created
rolebinding.rbac.authorization.k8s.io/wf-role-binding created
clusterrolebinding.rbac.authorization.k8s.io/operator-role-binding created
```

## Create Workflow

```console
$ kubectl apply -f ./docs/examples/deploy/workflow.yaml
workflow.engine.kube.ci/sample-workflow created
```

```yaml
apiVersion: engine.kube.ci/v1alpha1
kind: Workflow
metadata:
  name: sample-workflow
  namespace: default
spec:
  triggers:
  - apiVersion: v1
    kind: ConfigMap
    resource: configmaps
    namespace: default
    name: sample-config
    onCreateOrUpdate: true
    onDelete: false
  serviceAccount: wf-sa
  executionOrder: Serial
  allowManualTrigger: true
  steps:
  - name: clone
    image: alpine/git
    commands:
    - sh
    args:
    - -c
    - git clone https://github.com/kube-ci/kubeci-gpig .
  - name: build-and-push
    image: gcr.io/kaniko-project/executor
    args:
    - --dockerfile=/kubeci/workspace/Dockerfile
    - --context=/kubeci/workspace
    - --destination=index.docker.io/kubeci/kubeci-gpig:kaniko
  - name: deploy
    image: appscode/kubectl:1.12
    commands:
    - sh
    args:
    - -c
    - kubectl apply -f deploy.yaml
```

Here, step `clone` clones the source repository from Github into current working directory i.e. `/kubeci/workspace`. This repository contains following `Dockerfile` in it's root directory.

```
# build stage
FROM golang:alpine AS build-env
ADD . /go/src/github.com/kube-ci/kubeci-gpig
WORKDIR /go/src/github.com/kube-ci/kubeci-gpig
RUN CGO_ENABLED=0 go build -o goapp && chmod +x goapp

# final stage
FROM alpine
COPY --from=build-env /go/src/github.com/kube-ci/kubeci-gpig/goapp /usr/bin/kubeci-gpig
ENTRYPOINT ["kubeci-gpig"]
```

Step `build-and-push` runs a `kaniko` image which builds container from the specified Dockerfile and pushes it into the specified registry.

Step `deploy` runs a `kubectl` image which deploys your application using following `deploy.yaml` file contained in the source repository.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubeci-gpig-deployment
  labels:
    app: kubeci-gpig
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubeci-gpig
  template:
    metadata:
      labels:
        app: kubeci-gpig
    spec:
      containers:
      - name: kubeci-gpig
        image: kubeci/kubeci-gpig:kaniko
        ports:
        - containerPort: 9090
```

## Trigger Workflow

Now trigger the workflow by creating a `Trigger` custom-resource which contains a complete ConfigMap resource inside `.request` section.

```console
$ kubectl apply -f ./docs/examples/deploy/trigger.yaml
trigger.extensions.kube.ci/sample-trigger created
```

Whenever a workflow is triggered, a workplan is created and respective pods are scheduled.

```console
$ kubectl get workplan -l workflow=sample-workflow
NAME                    CREATED AT
sample-workflow-zzqkx   5s
```

```console
$ kubectl get pods -l workplan=sample-workflow-zzqkx
NAME                      READY   STATUS     RESTARTS   AGE
sample-workflow-zzqkx-0   0/1     Init:2/4   0          47s
```

## Check Logs

You can check logs of any step to verify if all the operations has been successfully completed or not. For example, run following command to check logs of `build-and-push` step:

```console
$ kubectl get --raw '/apis/extensions.kube.ci/v1alpha1/namespaces/default/workplanlogs/sample-workflow-zzqkx?step=build-and-push'
```

## Check Deployment

```console
$ kubectl get deployment -l app=kubeci-gpig
NAME                     DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
kubeci-gpig-deployment   1         1         1            1           11s

$ kubectl get pod -l app=kubeci-gpig
NAME                                      READY   STATUS    RESTARTS   AGE
kubeci-gpig-deployment-74c75c9b79-w89h8   1/1     Running   0          43s

$ kubectl logs kubeci-gpig-deployment-74c75c9b79-w89h8
Starting server
```

## Cleanup

```console
$ kubectl delete -f ./docs/examples/deploy
serviceaccount "wf-sa" deleted
clusterrole.rbac.authorization.k8s.io "wf-role" deleted
rolebinding.rbac.authorization.k8s.io "wf-role-binding" deleted
clusterrolebinding.rbac.authorization.k8s.io "operator-role-binding" deleted
secret "docker-credential" deleted
workflow.engine.kube.ci "sample-workflow" deleted
Error from server (NotFound): error when deleting "docs/examples/deploy/trigger.yaml": the server could not find the requested resource

$ kubectl delete deployment -l app=kubeci-gpig
deployment.extensions "kubeci-gpig-deployment" deleted
```
