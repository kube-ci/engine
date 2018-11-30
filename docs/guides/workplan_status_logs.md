> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# Workplan Status and Logs

This tutorial will show you how to use interactive `workplan-logs` CLI to collect logs of different steps and `workplan-viewer` web-ui to view workplan-status and logs.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

## Configure RBAC

You need to specify a service-account in `spec.serviceAccount` to ensure RBAC for the workflow. This service-account along with operator's service-account must have `list` and `watch` permissions for the resources specified in `spec.triggers`.

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
  allowManualTrigger: true
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
trigger.extensions.kube.ci/sample-trigger created
```

Whenever a workflow is triggered, a workplan is created and respective pods are scheduled.

```console
$ kubectl get workplan -l workflow=sample-workflow
NAME                    CREATED AT
sample-workflow-v8skm   13s
```

```console
$ kubectl get pods -l workplan=sample-workflow-v8skm
NAME                      READY   STATUS      RESTARTS   AGE
sample-workflow-v8skm-0   0/1     Completed   0          29s
```

## Workplan Logs CLI

You need to provide namespace, workplan and step using flags. For example:

```console
$ kubeci-engine workplan-logs --namespace default --workplan sample-workflow-v8skm --step step-echo
hello world
```

Alternatively, we can choose them interactively:

```console
$ kubeci-engine workplan-logs
? Choose a Namespace: default
? Choose a Workflow: sample-workflow
? Choose a Workplan: sample-workflow-v8skm
? Choose a running/terminated step: step-echo
hello world
```

## Workplan Viewer

The web-ui is deployed as a sidecar container during installation process. You can expose it with port forwarding. The status page will refresh after every 30 seconds. For `Running` and `Terminated` steps, there are links for accessing logs.

- Status URL: `http://127.0.0.1:9090/namespaces/{namespace}/workplans/{workplan-name}`
- Logs URL: `http://127.0.0.1:9090/namespaces/{namespace}/workplans/{workplan-name}/steps/{step-name}`

First, get the operator pod:

```console
$ kubectl get pods --all-namespaces | grep kubeci-engine
kube-system   kubeci-engine-5944797bb5-qxn5n                 3/3     Running     0          65m
```

Now, forward `9090` port of operator pod:

```console
$ kubectl port-forward -n kube-system kubeci-engine-5944797bb5-qxn5n 9090:9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Go to following URL to get current of status workplan `sample-workflow-v8skm`:

`http://127.0.0.1:9090/namespaces/default/workplans/sample-workflow-v8skm`

Go to following URL to get logs of step `step-echo`:

`http://127.0.0.1:9090/namespaces/default/workplans/sample-workflow-v8skm/steps/step-echo`
