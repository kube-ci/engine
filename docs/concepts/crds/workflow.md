# Workflow

## What is Workflow

A `Workflow` is a Kubernetes `CustomResourceDefinition` (CRD). It provides configuration for a set of sequential and/or parallel tasks and conditions for triggering them.

## Workflow Spec

As with all other Kubernetes objects, a Workflow needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example Workflow object:

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
    name: my-config
    onCreateOrUpdate: true
  steps:
  - name: step-01
    image: alpine
    commands:
    - echo
    args:
    - step-one
  - name: step-02
    image: alpine
    commands:
    - echo
    args:
    - step-two
```

The `.spec` section has following parts:

### spec.triggers

Specifies a set of Kubernetes resources along with their events (create/delete/update). The workflow will be triggered for any of those events. A trigger can have following fields to specify resources and filter events:

- `apiVersion` (required): API version of the resource. For example: v1alpha1, v1beta2, etc.
- `kind` (required): Kind of the resource. For example: ConfigMap, Secret, etc.
- `resource` (required): For example: configmaps, secrets, etc.
- `name`: If specified, only resources with this name will be considered.
- `namespace`: If specified, only resources in this namespace will be considered. Otherwise, all namespaces will be qualified.
- `selector`: If specified, resources with matching label selectors will be considered. Otherwise, any labels will be qualified.
- `onCreateOrUpdate`: If true workflow will be triggered when specified resource is created or, updated.
- `onDelete`: If true workflow will be triggered when specified resource is deleted. Note that, you should set at least one of `onCreateOrUpdate` and `onDelete` to true.
- `envFromPath`: Task containers might need some information about the resource for which the workflow was triggered. Those can be make available to task containers by means of environment variables. Using `envFromPath` you can specify a set of key value pairs indicating which json-path data to be mapped to which environment variable.

For example, let's say your workflow is triggered for following configmap:

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: my-config
  namespace: default
data:
  example.property.1: hello
  example.property.2: world
```

And you have specified following `envFromPath`:

```yaml
envFromPath:
  DATA_ONE: '{$.data.example\.property\.1}'
```

In this case, all containers will have `DATA_ONE` environment variable with value `hello`.

### spec.steps

Specifies a set of tasks which will be executed in serial or, parallel order. A step can have following fields:

- `name`: An unique name across the workflow steps.
- `image`: Container image for the task.
- `commands`: Set of commands to be executed.
- `args`: Set of arguments to the commands.
- `dependency`: Set of step-names which required to be completed before running this step. It should only be specified when `spec.executionOrder` is `DAG`.
- `volumeMounts`: Workflow volumes (specified in `workflow.spec.volumes`) to mount into the container's filesystem. For more details see [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#volumemount-v1-core).
- `securityContext`: Specifies security context for the container associated with this step. For more details see [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#securitycontext-v1-core).

### spec.template

- `name`: Name of the workflow template to be invoked.
- `arguments`: Set of key value pairs to be used as arguments to substitute template placeholders. For more details on substitution see [here](https://github.com/drone/docs/blob/v0.8.0/content/usage/config/substitution.md).

Note that, you can not specify `spec.steps` while using template.

### spec.executionOrder

Specifies the order in which the steps will be executed. Possible values are:

- `Serial`: Steps will be executed one by one in the given order.
- `Parallel`: All the steps will be executed simultaneously.
- `DAG`: Steps will be executed sequentially or, simultaneously based on dependency of each steps.

Note that, if nothing specified, `Serial` will be used as default execution order.

### spec.allowManualTrigger

When true, manual/fake trigger wil be allowed. See [here](/docs/guides/force_trigger.md) for more details about how to perform this.

### spec.serviceAccount

Name of the service-account to ensure RBAC for the workflow. This service-account along with operator's service-account must have `list` and `watch` permissions for the resources specified in `spec.triggers`. This service-account is also used to run all pods associated with the workflow.

### spec.envVar

List of environment variables to set in all task containers. For more details see [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#envvar-v1-core).

### spec.envFrom

List of sources to populate environment variables in all task containers. For more details see [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#envfromsource-v1-core).

### spec.volumes

List of volumes that can be mounted by the task containers belonging to the workflow. For more details see [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#volume-v1-core).

### spec.securityContext

Specifies security context for all pods associated with this workflow. For more details see [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#podsecuritycontext-v1-core).