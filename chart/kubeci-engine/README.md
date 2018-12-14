# KubeCI Engine
[KubeCI Engine by AppsCode](https://github.com/kube-ci/engine) - Kubernetes Native Workflow Engine

## TL;DR;

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm install appscode/kubeci-engine --name kubeci-engine --namespace kube-system
```

## Introduction

This chart bootstraps a [KubeCI engine controller](https://github.com/kube-ci/engine) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.9+

## Installing the Chart

To install the chart with the release name `kubeci-engine`:

```console
$ helm install appscode/kubeci-engine --name kubeci-engine
```

The command deploys KubeCI engine operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `kubeci-engine`:

```console
$ helm delete kubeci-engine
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the KubeCI engine chart and their default values.

| Parameter                            | Description                                                       | Default            |
| ------------------------------------ | ----------------------------------------------------------------- | ------------------ |
| `replicaCount`                       | Number of KubeCI engine replicas to create (only 1 is supported)  | `1`                |
| `operator.registry`                  | Docker registry used to pull operator image                       | `kubeci`           |
| `operator.repository`                | Operator container image                                          | `kubeci-engine`    |
| `operator.tag`                       | Operator container image tag                                      | `0.7.0`            |
| `cleaner.registry`                   | Docker registry used to pull Webhook cleaner image                | `appscode`         |
| `cleaner.repository`                 | Webhook cleaner container image                                   | `kubectl`          |
| `cleaner.tag`                        | Webhook cleaner container image tag                               | `v1.11`            |
| `imagePullPolicy`                    | Container image pull policy                                       | `IfNotPresent`     |
| `criticalAddon`                      | If true, installs KubeCI engine operator as critical addon        | `false`            |
| `logLevel`                           | Log level for operator                                            | `3`                |
| `affinity`                           | Affinity rules for pod assignment                                 | `{}`               |
| `annotations`                        | Annotations applied to operator pod(s)                            | `{}`               |
| `nodeSelector`                       | Node labels for pod assignment                                    | `{}`               |
| `tolerations`                        | Tolerations used pod assignment                                   | `{}`               |
| `rbac.create`                        | If `true`, create and use RBAC resources                          | `true`             |
| `serviceAccount.create`              | If `true`, create a new service account                           | `true`             |
| `serviceAccount.name`                | Service account to be used. If not set and `serviceAccount.create` is `true`, a name is generated using the fullname template | `` |
| `apiserver.groupPriorityMinimum`     | The minimum priority the group should have.                       | 10000              |
| `apiserver.versionPriority`          | The ordering of this API inside of the group.                     | 15                 |
| `apiserver.enableValidatingWebhook`  | Enable validating webhooks for KubeCI engine CRDs                 | true               |
| `apiserver.enableMutatingWebhook`    | Enable mutating webhooks for Kubernetes workloads                 | true               |
| `apiserver.ca`                       | CA certificate used by main Kubernetes api server                 | `not-ca-cert`      |
| `apiserver.disableStatusSubresource` | If true, disables status sub resource for crds. Otherwise enables based on Kubernetes version | `false`            |
| `enableAnalytics`                    | Send usage events to Google Analytics                             | `true`             |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```console
$ helm install --name kubeci-engine --set image.tag=v0.2.1 appscode/kubeci-engine
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while
installing the chart. For example:

```console
$ helm install --name kubeci-engine --values values.yaml appscode/kubeci-engine
```

## RBAC

By default the chart will not install the recommended RBAC roles and rolebindings.

You need to have the flag `--authorization-mode=RBAC` on the api server. See the following document for how to enable [RBAC](https://kubernetes.io/docs/admin/authorization/rbac/).

To determine if your cluster supports RBAC, run the following command:

```console
$ kubectl api-versions | grep rbac
```

If the output contains "beta", you may install the chart with RBAC enabled (see below).

### Enable RBAC role/rolebinding creation

To enable the creation of RBAC resources (On clusters with RBAC). Do the following:

```console
$ helm install --name kubeci-engine appscode/kubeci-engine --set rbac.create=true
```
