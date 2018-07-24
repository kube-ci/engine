# Kube CI Design Doc

## Workflow CRD

```yaml
apiVersion: kube.ci/v1alpha1
kind: Workflow
metadata:
  name: my-workflow
  namespace: default
spec:
  steps:
  - args:
    - -c
    - touch /kubeci/file-1
    commands:
    - sh
    image: alpine
    name: step-1
  - args:
    - -c
    - touch /kubeci/file-2
    commands:
    - sh
    dependency:
    - step-1
    image: alpine
    name: step-2
  - args:
    - -c
    - touch /kubeci/file-3
    commands:
    - sh
    dependency:
    - step-1
    image: alpine
    name: step-3
  - args:
    - -c
    - touch /kubeci/file-4
    commands:
    - sh
    dependency:
    - step-2
    - step-3
    image: alpine
    name: step-4
  - args:
    - -c
    - echo step-5; ls /kubeci
    commands:
    - sh
    dependency:
    - step-1
    - step-4
    image: alpine
    name: step-5
  triggers:
  - apiVersion: v1
    kind: ConfigMap
    name: my-config
    namespace: default
    onAddOrUpdate: true
    onDelete: true
    resource: configmaps
    selector: {}
status:
  lastObservedGeneration:         # workplan generation
  lastObservedResourceGeneration: # map[resource-uid]resource-generation
```

### Shared Volume

- all pods of the workflow will be scheduled in a specific node
- resources/outputs will be stored in node's hostpath
- we can use local PV as shared volume
- we can also use file sharing techniques (like torrent) among the nodes

### Triggers

- list of kubernetes resources, identified by apiVersion and kind
- user can also specify namespace, selector or name to filter resources
- a list of events must be specified for each resource. Workflow will be triggered for any of this events
- workflow-operator will watch this resources using `dynamic-informer`
- when a workflow event is triggered, workflow-operator will create a workplan custom resource by resolving dependency. Then the workplan-operator will run and manage the required pods

### Steps

- a step should specify necessary fields (image, commands, args) for running a container
- a step can be dependent of other steps. Init-containers can be used for running sequential steps and sidecar-containers can be used for running parallel steps
- all pods will share inputs/outputs through a common hostpath volume (`/kubeci/{workflow-name}/{workplan-name}`). This volume will be mounted in a specific path (`/kubeci`) and containers should put their outputs in this specific folder
- when a workplan completed/failed, workplan-operator will clean-up the hostpath. So user should publish required outputs elsewhere (s3/gcs)

## Workplan CRD

```yaml
apiVersion: kube.ci/v1alpha1
kind: Workplan
metadata:
  generateName: my-workflow-
  namespace: default
  ownerReferences: # workflow object reference
  - apiVersion: kube.ci/v1alpha1 
    kind: Workflow
    name: my-workflow
spec:
  tasks:
  - SerialSteps:
    - args:
      - -c
      - touch /kubeci/file-1
      commands:
      - sh
      image: alpine
      name: step-1
    ParallelSteps:
    - args:
      - -c
      - touch /kubeci/file-2
      commands:
      - sh
      dependency:
      - step-1
      image: alpine
      name: step-2
    - args:
      - -c
      - touch /kubeci/file-3
      commands:
      - sh
      dependency:
      - step-1
      image: alpine
      name: step-3   
  - SerialSteps:
    - args:
      - -c
      - touch /kubeci/file-4
      commands:
      - sh
      dependency:
      - step-2
      - step-3
      image: alpine
      name: step-4
    ParallelSteps:
    - args:
      - -c
      - echo step-5; ls /kubeci
      commands:
      - sh
      dependency:
      - step-1
      - step-4
      image: alpine
      name: step-5
status:
  phase:     # pending, running, completed, failed
  reason:    # messages and errors
  taskIndex: # current running task
```