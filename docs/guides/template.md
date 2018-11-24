> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# Workflow Template

This tutorial will show you how to create a workflow-template and invoke it from a workflow. You can invoke same template from multiple workflows in the same namespace.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

## Configure RBAC

First, create a service-account for the workflow. Then, create a cluster-role with ConfigMap `list` and `watch` permissions. Now, bind it with service-accounts of both workflow and operator.

```console
$ kubectl apply -f ./docs/examples/template/rbac.yaml
serviceaccount/wf-sa created
clusterrole.rbac.authorization.k8s.io/wf-role created
rolebinding.rbac.authorization.k8s.io/wf-role-binding created
clusterrolebinding.rbac.authorization.k8s.io/operator-role-binding created
```

## Create Template and Workflow

```console
$ kubectl apply -f ./docs/examples/template/template.yaml
workflowtemplate.engine.kube.ci/sample-template created
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
  template:
    name: sample-template
    arguments:
      image: alpine
      cmd: echo hello world
```

```console
$ kubectl apply -f ./docs/examples/template/workflow.yaml
workflow.engine.kube.ci/sample-workflow created
```

```yaml
apiVersion: engine.kube.ci/v1alpha1
kind: WorkflowTemplate
metadata:
  name: sample-template
  namespace: default
spec:
  steps:
  - name: step-echo
    image: ${image=busybox}
    commands:
    - ${shell=sh}
    args:
    - -c
    - ${cmd}
```

After resolving template steps will look like below:

```yaml
steps:
- name: step-echo
  image: alpine      # substituted value
  commands:
  - sh               # default value
  args:
  - -c
  - echo hello world # substituted value
```

## Trigger Workflow

Now trigger the workflow by creating a `Trigger` custom-resource which contains a complete ConfigMap resource inside `.request` section.

```console
$ kubectl apply -f ./docs/examples/template/trigger.yaml
trigger.extension.kube.ci/sample-trigger created
```

Whenever a workflow is triggered, a workplan is created and respective pods are scheduled.

```console
$ kubectl get workplan -l workflow=sample-workflow
NAME                    CREATED AT
sample-workflow-smld7   14s
```

```console
$ kubectl get pods -l workplan=sample-workflow-smld7
NAME                      READY   STATUS      RESTARTS   AGE
sample-workflow-smld7-0   0/1     Completed   0          48s
```

## Cleanup

```console
$ kubectl delete -f docs/examples/template/
serviceaccount "wf-sa" deleted
clusterrole.rbac.authorization.k8s.io "wf-role" deleted
rolebinding.rbac.authorization.k8s.io "wf-role-binding" deleted
clusterrolebinding.rbac.authorization.k8s.io "operator-role-binding" deleted
workflowtemplate.engine.kube.ci "sample-template" deleted
workflow.engine.kube.ci "sample-workflow" deleted
Error from server (NotFound): error when deleting "docs/examples/template/trigger.yaml": the server could not find the requested resource
```
