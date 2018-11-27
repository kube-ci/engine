# Trigger

## What is Trigger

A `Trigger` is a representation of a Kubernetes object with the help of [Aggregated API Servers](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/aggregated-api-servers.md). User can manually trigger a workflow by creating a `Trigger` resource. For this, `workflow.spec.allowManualTrigger` must be `true`. Note that, only `create` verb is available for this custom resource.

## Trigger structure

As with all other Kubernetes objects, a Trigger needs `apiVersion`, `kind`, and `metadata` fields. It also includes `.workflows` and `.request` sections. Below is an example Trigger object:

```yaml
apiVersion: extensions.kube.ci/v1alpha1
kind: Trigger
metadata:
  name: testing-force-trigger
  namespace: default
workflows:
- wf-force-trigger
request:
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: my-config
    namespace: default
  data:
    hello: world
```

Here, we are going to describe some important sections of `Trigger` object.

### .workflows

A list of workflows in the same namespace which will be considered for trigger. If not specified, all workflows in the same namespace will be considered.

### .request

A complete representation of a Kubernetes object along with `apiVersion`, `kind`, and `metadata` fields. When a trigger is created, it will act as a fake create event for the object and can only be used to manually trigger a workflow without actually creating the object.
