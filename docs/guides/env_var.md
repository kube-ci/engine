> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# Explicit Environment Variables

This tutorial will show you how to specify explicit environment variables and populate them from source(configmaps/secrets) inside all step-containers.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

## Configure RBAC

First, create a service-account for the workflow. Then, create a cluster-role with ConfigMap `list` and `watch` permissions. Since we will populate environment variables from secret, we also need secret get permission. Now, bind the cluster-role with service-accounts of both workflow and operator.

```console
$ kubectl apply -f ./docs/examples/env-var/rbac.yaml
serviceaccount/wf-sa created
clusterrole.rbac.authorization.k8s.io/wf-role created
rolebinding.rbac.authorization.k8s.io/wf-role-binding created
clusterrolebinding.rbac.authorization.k8s.io/operator-role-binding created
```

## Create Secret

Create secret which will be used as source for populating environment variables.

```console
$ kubectl apply -f ./docs/examples/env-var/secret.yaml
secret/sample-secret created
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sample-secret
  namespace: default
type: Opaque
data:
  KEY_ONE: c2VjcmV0LWRhdGEtb25l # echo -n 'secret-data-one' | base64
  KEY_TWO: c2VjcmV0LWRhdGEtdHdv # echo -n 'secret-data-two' | base64
```

## Create Workflow

```console
$ kubectl apply -f ./docs/examples/env-var/workflow.yaml
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
  allowForceTrigger: true
  steps:
  - name: step-one
    image: alpine
    commands:
    - sh
    args:
    - -c
    - echo ENV_ONE=$ENV_ONE; echo ENV_TWO=$ENV_TWO; echo KEY_ONE=$KEY_ONE; echo KEY_TWO=$KEY_TWO
  envVar:
  - name: ENV_ONE
    value: one
  - name: ENV_TWO
    valueFrom:
      secretKeyRef:
        name: sample-secret
        key: KEY_TWO
  envFrom:
  - secretRef:
      name: sample-secret
```

## Trigger Workflow

Now trigger the workflow by creating a `Trigger` custom-resource which contains a complete ConfigMap resource inside `.request` section.

```console
$ kubectl apply -f ./docs/examples/env-var/trigger.yaml
trigger.extension.kube.ci/sample-trigger created
```

Whenever a workflow is triggered, a workplan is created and respective pods are scheduled.

```console
$ kubectl get workplan -l workflow=sample-workflow
NAME                    CREATED AT
sample-workflow-blkh9   5s
```

```console
$ kubectl get pods -l workplan=sample-workflow-blkh9
NAME                      READY   STATUS      RESTARTS   AGE
sample-workflow-blkh9-0   0/1     Completed   0          25s
```

## Check Logs

The `step-one` prints the values of explicit environment variables.

```console
$ kubectl get --raw '/apis/extension.kube.ci/v1alpha1/namespaces/default/workplanlogs/sample-workflow-blkh9?step=step-one'
ENV_ONE=one
ENV_TWO=secret-data-two
KEY_ONE=secret-data-one
KEY_TWO=secret-data-two
```

Here, we can see that, `HOME`, `NAMESPACE` and `WORKPLAN` environment variables are available in both containers.

## Cleanup

```console
$ kubectl delete -f docs/examples/env-var/
serviceaccount "wf-sa" deleted
clusterrole.rbac.authorization.k8s.io "wf-role" deleted
rolebinding.rbac.authorization.k8s.io "wf-role-binding" deleted
clusterrolebinding.rbac.authorization.k8s.io "operator-role-binding" deleted
secret "sample-secret" deleted
workflow.engine.kube.ci "sample-workflow" deleted
Error from server (NotFound): error when deleting "docs/examples/env-var/trigger.yaml": the server could not find the requested resource
```
