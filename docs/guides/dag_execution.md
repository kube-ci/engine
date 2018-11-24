> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# DAG Dependency

This tutorial will show you how to use KubeCI engine to configure a Workflow consisting of a set of tasks with DAG dependency.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

## Configure RBAC

First, create a service-account for the workflow. Then, create a cluster-role with ConfigMap `list` and `watch` permissions. Now, bind it with service-accounts of both workflow and operator.

```console
$ kubectl apply -f ./docs/examples/dag-execution/rbac.yaml
serviceaccount/wf-sa created
clusterrole.rbac.authorization.k8s.io/wf-role created
rolebinding.rbac.authorization.k8s.io/wf-role-binding created
clusterrolebinding.rbac.authorization.k8s.io/operator-role-binding created
```

## Create Workflow

```console
$ kubectl apply -f ./docs/examples/dag-execution/workflow.yaml
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
  executionOrder: DAG
  allowForceTrigger: true
  steps:
  - name: step-01
    image: alpine
    commands:
    - sh
    args:
    - -c
    - date; sleep 5s; date
  - name: step-02
    image: alpine
    commands:
    - sh
    args:
    - -c
    - date; sleep 5s; date
    dependency:
    - step-01
  - name: step-03
    image: alpine
    commands:
    - sh
    args:
    - -c
    - date; sleep 5s; date
    dependency:
    - step-01
  - name: step-04
    image: alpine
    commands:
    - sh
    args:
    - -c
    - date; sleep 5s; date
    dependency:
    - step-03
  - name: step-05
    image: alpine
    commands:
    - sh
    args:
    - -c
    - date; sleep 5s; date
    dependency:
    - step-04
  - name: step-06
    image: alpine
    commands:
    - sh
    args:
    - -c
    - date; sleep 5s; date
    dependency:
    - step-04
# dependency: 01 | 02 03 | 04 | 05 06
```

## Trigger Workflow

Now trigger the workflow by creating a `Trigger` custom-resource which contains a complete ConfigMap resource inside `.request` section.

```console
$ kubectl apply -f ./docs/examples/dag-execution/trigger.yaml
trigger.extension.kube.ci/sample-trigger created
```

Whenever a workflow is triggered, a workplan is created and respective pods are scheduled.

```console
$ kubectl get workplan -l workflow=sample-workflow
NAME                    CREATED AT
sample-workflow-7kndd   8s
```

```console
$ kubectl get pods -l workplan=sample-workflow-7kndd
NAME                      READY   STATUS      RESTARTS   AGE
sample-workflow-7kndd-0   0/2     Completed   0          32s
sample-workflow-7kndd-1   0/2     Completed   0          19s
sample-workflow-7kndd-2   0/1     Completed   0          7s
```

## Get Resolved Dependency

Process: `workflow.spec.steps` ---> `resolve-dependency` ---> `layers` ---> `workplan.spec.tasks` ---> `pods`

To know the resolved dependency check the `.spec.tasks` and `.status.stepTree` sections of the workplan.

```yaml
$ kubectl get workplan sample-workflow-7kndd -o yaml
apiVersion: engine.kube.ci/v1alpha1
kind: Workplan
metadata:
  creationTimestamp: 2018-11-08T06:16:49Z
  generateName: sample-workflow-
  generation: 1
  labels:
    workflow: sample-workflow
  name: sample-workflow-7kndd
  namespace: default
  ownerReferences:
  - apiVersion: engine.kube.ci/v1alpha1
    blockOwnerDeletion: true
    kind: Workflow
    name: sample-workflow
    uid: d982e107-e31d-11e8-a7e0-080027868e9e
  resourceVersion: "8081"
  selfLink: /apis/engine.kube.ci/v1alpha1/namespaces/default/workplans/sample-workflow-7kndd
  uid: e06eb1cb-e31d-11e8-a7e0-080027868e9e
spec:
  tasks:
  - ParallelSteps:
    - args:
      - -c
      - date; sleep 5s; date
      commands:
      - sh
      dependency:
      - step-01
      image: alpine
      name: step-02
    - args:
      - -c
      - date; sleep 5s; date
      commands:
      - sh
      dependency:
      - step-01
      image: alpine
      name: step-03
    SerialSteps:
    - args:
      - -c
      - date; sleep 5s; date
      commands:
      - sh
      image: alpine
      name: step-01
  - ParallelSteps:
    - args:
      - -c
      - date; sleep 5s; date
      commands:
      - sh
      dependency:
      - step-04
      image: alpine
      name: step-05
    - args:
      - -c
      - date; sleep 5s; date
      commands:
      - sh
      dependency:
      - step-04
      image: alpine
      name: step-06
    SerialSteps:
    - args:
      - -c
      - date; sleep 5s; date
      commands:
      - sh
      dependency:
      - step-03
      image: alpine
      name: step-04
  - ParallelSteps:
    - args:
      - -c
      - echo deleting files/folders; ls /kubeci; rm -rf /kubeci/home/*; rm -rf /kubeci/workspace/*
      commands:
      - sh
      image: alpine
      name: cleanup-step
    SerialSteps: []
  triggeredFor:
    objectReference:
      apiVersion: v1
      kind: ConfigMap
      name: sample-config
      namespace: default
    resourceGeneration: 0$9874914804914738715
  workflow: sample-workflow
status:
  phase: Succeeded
  reason: All tasks completed successfully
  stepTree:
  - - ContainerState:
        terminated:
          containerID: docker://1dfb81655a610f16d445f6249206871682293f5726bd725c36594ef155c7f6c1
          exitCode: 0
          finishedAt: 2018-11-08T06:17:11Z
          reason: Completed
          startedAt: 2018-11-08T06:17:06Z
      Name: step-01
      Namespace: default
      PodName: sample-workflow-7kndd-0
      Status: Terminated
  - - ContainerState:
        terminated:
          containerID: docker://2bcaf93323a25ec76da607ea3b5e5415d2ec43242ebb5850933a5e8a0749febb
          exitCode: 0
          finishedAt: 2018-11-08T06:17:27Z
          reason: Completed
          startedAt: 2018-11-08T06:17:22Z
      Name: step-02
      Namespace: default
      PodName: sample-workflow-7kndd-0
      Status: Terminated
    - ContainerState:
        terminated:
          containerID: docker://6e51b84dacb3f427625c605d5630b41f15e19c4db906597eff00d65a1469916a
          exitCode: 0
          finishedAt: 2018-11-08T06:17:36Z
          reason: Completed
          startedAt: 2018-11-08T06:17:31Z
      Name: step-03
      Namespace: default
      PodName: sample-workflow-7kndd-0
      Status: Terminated
  - - ContainerState:
        terminated:
          containerID: docker://665b4b2a78c53e41fc2b39e89fabc8098f16f4b3cdffa59b88cc01811cfed5bb
          exitCode: 0
          finishedAt: 2018-11-08T06:18:02Z
          reason: Completed
          startedAt: 2018-11-08T06:17:57Z
      Name: step-04
      Namespace: default
      PodName: sample-workflow-7kndd-1
      Status: Terminated
  - - ContainerState:
        terminated:
          containerID: docker://ec205f9d8b5d705b2727ec5fe330a61c26228be41233873489c5501017a6d147
          exitCode: 0
          finishedAt: 2018-11-08T06:18:18Z
          reason: Completed
          startedAt: 2018-11-08T06:18:13Z
      Name: step-05
      Namespace: default
      PodName: sample-workflow-7kndd-1
      Status: Terminated
    - ContainerState:
        terminated:
          containerID: docker://136a77255a28ec50fadf2461af0147de89aed96e716bdd1550deda4c711ef554
          exitCode: 0
          finishedAt: 2018-11-08T06:18:27Z
          reason: Completed
          startedAt: 2018-11-08T06:18:22Z
      Name: step-06
      Namespace: default
      PodName: sample-workflow-7kndd-1
      Status: Terminated
  - - ContainerState:
        terminated:
          containerID: docker://bfe0cc512d7a45f418621f8f75db1b4f4789f2309895584f6dfd24ecc925ef40
          exitCode: 0
          finishedAt: 2018-11-08T06:18:47Z
          reason: Completed
          startedAt: 2018-11-08T06:18:47Z
      Name: cleanup-step
      Namespace: default
      PodName: sample-workflow-7kndd-2
      Status: Terminated
  taskIndex: -1
```

## Cleanup

```console
$ kubectl delete -f docs/examples/dag-execution/
serviceaccount "wf-sa" deleted
clusterrole.rbac.authorization.k8s.io "wf-role" deleted
rolebinding.rbac.authorization.k8s.io "wf-role-binding" deleted
clusterrolebinding.rbac.authorization.k8s.io "operator-role-binding" deleted
workflow.engine.kube.ci "sample-workflow" deleted
Error from server (NotFound): error when deleting "docs/examples/dag-execution/trigger.yaml": the server could not find the requested resource
```
