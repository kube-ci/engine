# Workflow Template

## What is Workflow Template

A `WorkflowTemplate` is a Kubernetes `CustomResourceDefinition` (CRD) which can be invoked by other Workflows in the same namespace.

## Workflow Template Spec

As with all other Kubernetes objects, a Workflow Template needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec.steps` section which contains a set of steps similar to [workflow-steps](workflow.md#specsteps). But here you can use placeholders which will be substituted by arguments during invocation. For more details on substitution see [here](https://github.com/drone/docs/blob/v0.8.0/content/usage/config/substitution.md).

Below is an example Workflow Template object:

```yaml
apiVersion: engine.kube.ci/v1alpha1
kind: WorkflowTemplate
metadata:
  name: wf-template
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

Here is an example Workflow which invokes the template:

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
  template:
    name: wf-template
    arguments:
      image: alpine
      cmd: echo $HOME
```

And after resolving the template steps will look like bellow:

```yaml
steps:
- name: step-echo
  image: alpine # substituted value
  commands:
  - sh          # default value
  args:
  - -c
  - echo $HOME  # substituted value
```