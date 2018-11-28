> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# JSON Path Data

Your steps might need some information about the resource for which the workflow was triggered. This tutorial will show you how to populate explicit environment variables from json-path data of the triggering resource. Using `envFromPath` you can specify a set of key value pairs indicating which json-path data to be mapped to which environment variable.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

## Configure RBAC

First, create a service-account for the workflow. Then, create a cluster-role with ConfigMap `get`, `list` and `watch` permissions. Now, bind the cluster-role with service-accounts of both workflow and operator.

```console
$ kubectl apply -f ./docs/examples/json-path/rbac.yaml
serviceaccount/wf-sa created
clusterrole.rbac.authorization.k8s.io/wf-role created
rolebinding.rbac.authorization.k8s.io/wf-role-binding created
clusterrolebinding.rbac.authorization.k8s.io/operator-role-binding created
```

## Create Workflow

```console
$ kubectl apply -f ./docs/examples/json-path/workflow.yaml
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
    envFromPath:
      ENV_ONE: '{$.data.example\.property\.1}'
      ENV_TWO: '{$.data.example\.property\.2}'
  serviceAccount: wf-sa
  allowManualTrigger: true
  steps:
  - name: step-one
    image: alpine
    commands:
    - sh
    args:
    - -c
    - echo ENV_ONE=$ENV_ONE; echo ENV_TWO=$ENV_TWO
```

## Trigger Workflow

Now trigger the workflow by creating a `Trigger` custom-resource which contains a complete ConfigMap resource inside `.request` section.

```console
$ kubectl apply -f ./docs/examples/json-path/trigger.yaml
trigger.extensions.kube.ci/sample-trigger created
```

Whenever a workflow is triggered, a workplan is created and respective pods are scheduled.

```console
$ kubectl get workplan -l workflow=sample-workflow
NAME                    CREATED AT
sample-workflow-8nfcr   5s
```

```console
$ kubectl get pods -l workplan=sample-workflow-8nfcr
NAME                      READY   STATUS      RESTARTS   AGE
sample-workflow-8nfcr-0   0/1     Completed   0          25s
```

Also, resolved environment variables from json-path data are set to `spec.envVar` of the workplan.

```yaml
$ kubectl get workplan sample-workflow-8nfcr -o yaml
apiVersion: engine.kube.ci/v1alpha1
kind: Workplan
metadata:
  creationTimestamp: 2018-11-28T05:16:56Z
  generateName: sample-workflow-
  generation: 1
  labels:
    workflow: sample-workflow
  name: sample-workflow-8nfcr
  namespace: default
  ownerReferences:
  - apiVersion: engine.kube.ci/v1alpha1
    blockOwnerDeletion: true
    kind: Workflow
    name: sample-workflow
    uid: ce8c454c-f2cc-11e8-a969-0800270eb1c1
  resourceVersion: "5114"
  selfLink: /apis/engine.kube.ci/v1alpha1/namespaces/default/workplans/sample-workflow-8nfcr
  uid: d2eaefb0-f2cc-11e8-a969-0800270eb1c1
spec:
  envVar:
  - name: ENV_TWO
    value: world
  - name: ENV_ONE
    value: hello
  tasks:
  - ParallelSteps:
    - args:
      - -c
      - echo deleting files/folders; ls /kubeci; rm -rf /kubeci/home/*; rm -rf /kubeci/workspace/*
      commands:
      - sh
      image: alpine
      name: cleanup-step
    SerialSteps:
    - args:
      - -c
      - echo ENV_ONE=$ENV_ONE; echo ENV_TWO=$ENV_TWO
      commands:
      - sh
      image: alpine
      name: step-one
  triggeredFor:
    objectReference:
      apiVersion: v1
      kind: ConfigMap
      name: sample-config
      namespace: default
    resourceGeneration: 0$9874914804914738715
  workflow: sample-workflow
```

## Check Logs

The `step-one` prints the values of explicit environment variables populated from json-path data of the triggering resource (which is `sample-config` configmap in this example).

```console
$ kubectl get --raw '/apis/extensions.kube.ci/v1alpha1/namespaces/default/workplanlogs/sample-workflow-8nfcr?step=step-one'
ENV_ONE=hello
ENV_TWO=world
```

Here, we can see that, `HOME`, `NAMESPACE` and `WORKPLAN` environment variables are available in both containers.

## Cleanup

```console
$ kubectl delete -f docs/examples/json-path/
serviceaccount "wf-sa" deleted
clusterrole.rbac.authorization.k8s.io "wf-role" deleted
rolebinding.rbac.authorization.k8s.io "wf-role-binding" deleted
clusterrolebinding.rbac.authorization.k8s.io "operator-role-binding" deleted
workflow.engine.kube.ci "sample-workflow" deleted
Error from server (NotFound): error when deleting "docs/examples/json-path/trigger.yaml": the server could not find the requested resource
```
