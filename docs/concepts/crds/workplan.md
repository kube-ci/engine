# Workplan

## What is Workplan

A `Workplan` is a Kubernetes `CustomResourceDefinition` (CRD). It is created by the KubeCI engine when a workflow is triggered. It provides the final sate of the workflow after resolving template and dependencies. Once created, users can not update it.

## Workplan Spec

As with all other Kubernetes objects, a Workplan has `apiVersion`, `kind`, and `metadata` fields. It also has a `.spec` section. Below is an example Workflow object:

```yaml
apiVersion: engine.kube.ci/v1alpha1
kind: Workplan
metadata:
  creationTimestamp: 2018-11-07T05:42:53Z
  generateName: sample-workflow-
  generation: 1
  labels:
    workflow: sample-workflow
  name: sample-workflow-7gx5d
  namespace: default
  ownerReferences:
  - apiVersion: engine.kube.ci/v1alpha1
    blockOwnerDeletion: true
    kind: Workflow
    name: sample-workflow
    uid: 100277d5-e17c-11e8-befc-0800279e79b0
  resourceVersion: "44750"
  selfLink: /apis/engine.kube.ci/v1alpha1/namespaces/default/workplans/sample-workflow-7gx5d
  uid: f8b7d4e6-e24f-11e8-8508-0800279e79b0
spec:
  envFrom:
  - secretRef:
      name: github-credential
  - secretRef:
      name: sample-workflow-2wbj2
  tasks:
  - SerialSteps:
    - args:
      - -c
      - git clone https://github.com/diptadas/kubeci-gpig.git .; git checkout $HEAD_SHA
      commands:
      - sh
      image: alpine/git
      name: step-clone
    - args:
      - -c
      - cat kubeci/test.sh | sh
      commands:
      - sh
      image: golang:1.10-alpine3.8
      name: step-test
    ParallelSteps:
    - args:
      - -c
      - echo deleting files/folders; ls /kubeci; rm -rf /kubeci/home/*; rm -rf /kubeci/workspace/*
      commands:
      - sh
      image: alpine
      name: cleanup-step
  triggeredFor:
    objectReference:
      apiVersion: git.kube.ci/v1alpha1
      kind: PullRequest
      name: kubeci-gpig-1
      namespace: default
    resourceGeneration: 5$14035282142003186040
  workflow: sample-workflow
status:
  phase: Succeeded
  reason: All tasks completed successfully
  stepTree:
  - - ContainerState:
        terminated:
          containerID: docker://b5847edfcedf25bbaee3606034cdf30fe38911c108ed3adab594a3f2b6773c0b
          exitCode: 0
          finishedAt: 2018-11-07T05:43:15Z
          reason: Completed
          startedAt: 2018-11-07T05:43:13Z
      Name: step-clone
      Namespace: default
      PodName: sample-workflow-7gx5d-0
      Status: Terminated
  - - ContainerState:
        terminated:
          containerID: docker://6f26921dbbed419bee6909e62e3c4b5c45c978701d746dde77008bb16f3d61de
          exitCode: 0
          finishedAt: 2018-11-07T05:43:58Z
          reason: Completed
          startedAt: 2018-11-07T05:43:21Z
      Name: step-test
      Namespace: default
      PodName: sample-workflow-7gx5d-0
      Status: Terminated
  - - ContainerState:
        terminated:
          containerID: docker://ea8047ef352bc97bdffb02e27b60797cc01a65599755033eb3c1cfce63b303dc
          exitCode: 0
          finishedAt: 2018-11-07T05:44:08Z
          reason: Completed
          startedAt: 2018-11-07T05:44:08Z
      Name: cleanup-step
      Namespace: default
      PodName: sample-workflow-7gx5d-0
      Status: Terminated
  taskIndex: -1
```

The `.spec` section has following parts:

### spec.workflow

Name of the Workflow for which the Workplan was created.

### spec.tasks

It is populated from `workflow.spec.steps` by resolving dependencies. A task represents a single Kubernetes pod where serial steps are converted into init-containers and parallel steps are converted into containers.

### spec.triggeredFor

Describes the resource for which the Workflow was triggered. It has following fields:

- `objectReference`: It contains apiVersion, kind, name and namespace of the resource.
- `resourceGeneration`: Hash value of resource spec, labels, annotations and generation.

### spec.envVar

Copied from `workflow.spec.envVar`.

### spec.envFrom

Copied from `workflow.spec.envFrom`. If `envFromPath` is found in `workflow.spec.triggers`, a new secret is created with json-path data and appended with `spec.envFrom`.

### spec.volumes

Copied from `workflow.spec.volumes`.

## Workplan Status

The status section of a workplan contains enough information to describe the current phase of a workplan. It has following sections:

### status.phase

Indicates current state of the workplan. Possible values are: `Uninitialized`, `Pending`, `Running`, `Succeeded` and `Failed`.

### status.reason

Describes the reason behind the current phase of the workplan.

### status.taskIndex

Indicates the zero-based index of the task which is currently running. In case of `Pending`, `Succeeded` and `Failed`, it is set to `-1`.

### status.stepTree

Collection of step-entries organized in a two-dimensional array based on dependency. A single step-entry describes the step-container in which the step is scheduled:

- Name: Name of the step.
- Namespace: Namespace of the workplan. Note that, all the workplans and pods are created in the same namespace of the Workflow.
- PodName: Name of the pod where step-container is scheduled.
- Status: Current status of the container. Possible values are: `Uninitialized` (pod not exists), `Waiting`, `Running` and `Terminated`.
- ContainerState: Value of [ContainerState](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.10/#containerstate-v1-core) from pod-status.
