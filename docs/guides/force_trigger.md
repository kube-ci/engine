> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# Force Trigger

This tutorial will show you how to trigger a Workflow manually. Here, we will use the same workflow used in previous [serial-execution](serial_execution.md) example, but trigger it without creating any ConfigMap. Note that, for force trigger we have to set `allowForceTrigger` to true.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

## Configure RBAC

First, create a service-account for the workflow. Then, create a cluster-role with ConfigMap `list` and `watch` permissions. Now, bind it with service-accounts of both workflow and operator.

```console
$ kubectl apply -f ./docs/examples/force-trigger/rbac.yaml
serviceaccount/wf-sa created
clusterrole.rbac.authorization.k8s.io/wf-role created
rolebinding.rbac.authorization.k8s.io/wf-role-binding created
clusterrolebinding.rbac.authorization.k8s.io/operator-role-binding created
```

## Create Workflow

```console
$ kubectl apply -f ./docs/examples/force-trigger/workflow.yaml
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
  allowForceTrigger: true
  steps:
  - name: step-echo
    image: alpine
    commands:
    - echo
    args:
    - hello world
  - name: step-wait
    image: alpine
    commands:
    - sleep
    args:
    - 10s
```

## Trigger Workflow

Now trigger the workflow by creating a `Trigger` custom-resource which contains a complete ConfigMap resource inside `.request` section.

```console
$ kubectl apply -f ./docs/examples/force-trigger/trigger.yaml
trigger.extension.kube.ci/sample-trigger created
```

```yaml
apiVersion: extension.kube.ci/v1alpha1
kind: Trigger
metadata:
  name: sample-trigger
  namespace: default
workflows:
- sample-workflow
request:
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: sample-config
    namespace: default
  data:
    example.property.1: hello
    example.property.2: world
```

Whenever a workflow is triggered, a workplan is created and respective pods are scheduled.

```console
$ kubectl get workplan -l workflow=sample-workflow
NAME                    CREATED AT
sample-workflow-wnmw2   13s
```

```console
$ kubectl get pods -l workplan=sample-workflow-wnmw2
NAME                      READY   STATUS      RESTARTS   AGE
sample-workflow-wnmw2-0   0/1     Completed   0          29s
```

## Cleanup

```console
$ kubectl delete -f docs/examples/force-trigger/
serviceaccount "wf-sa" deleted
clusterrole.rbac.authorization.k8s.io "wf-role" deleted
rolebinding.rbac.authorization.k8s.io "wf-role-binding" deleted
clusterrolebinding.rbac.authorization.k8s.io "operator-role-binding" deleted
workflow.engine.kube.ci "sample-workflow" deleted
Error from server (NotFound): error when deleting "docs/examples/force-trigger/trigger.yaml": the server could not find the requested resource
```
