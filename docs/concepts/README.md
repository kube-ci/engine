# Concepts

Concepts help you learn about the different parts of the KubeCI engine and the abstractions it uses.

- What is KubeCI engine?
  - [Overview](/docs/concepts/what-is-kubeci-engine/overview.md). Provides a conceptual introduction to KubeCI engine, including the problems it solves and its high-level architecture.
- Custom Resource Definitions
  - [Workflow](/docs/concepts/crds/workflow.md). Introduces the concept of `Workflow` for configuring a set of tasks in a Kubernetes native way.
  - [Workflow Template](/docs/concepts/crds/workflow_template.md). Introduces the concept of `WorkflowTemplate` to invoke a template with arguments for different workflows.
  - [Workplan](/docs/concepts/crds/workplan.md). Introduces the concept of `Workplan` that represents the final state of a workflow after it is triggered.
  - [Trigger](/docs/concepts/crds/trigger.md). Introduces the concept of `Trigger` that represents a fake create event for a Kubernetes resource to trigger workflows.
  - [Workplan Log](/docs/concepts/crds/workplan_log.md). Introduces the concept of `WorkplanLog` that can be used to collect logs of any workplan step.
