# KubeCI Engine

KubeCI engine by AppsCode is a Kubernetes native workflow engine.

## Features

- Configure a set of containerized steps using workflow.
- Run steps in serial or, parallel order by resolving dependencies for each step.
- Trigger workflows through create/update/delete events of any Kubernetes object.
- Trigger workflows manually with fake create events.
- Shared `workspace` and `home` directory among all steps of a workflow.
- Credential initializer for Docker and Git.
- APIs for collecting status and logs of each step.
