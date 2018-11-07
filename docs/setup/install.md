# Installation Guide

Kubeci-engine can be installed via a script or as a Helm chart.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="true">Script</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="helm-tab" data-toggle="tab" href="#helm" role="tab" aria-controls="helm" aria-selected="false">Helm</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using Script

To install Kubeci-engine in your Kubernetes cluster, run the following command:

```console
$ curl -fsSL https://raw.githubusercontent.com/kube-ci/engine/0.1.0/hack/deploy/install.sh | bash
```

After successful installation, you should have a `kubeci-engine-***` pod running in the `kube-system` namespace.

```console
$ kubectl get pods -n kube-system | grep kubeci-engine
kubeci-engine-846d47f489-jrb58       1/1       Running   0          48s
```

#### Customizing Installer

The installer script and associated yaml files can be found in the [/hack/deploy](https://github.com/kube-ci/engine/tree/0.1.0/hack/deploy) folder. You can see the full list of flags available to installer using `-h` flag.

```console
$ curl -fsSL https://raw.githubusercontent.com/kube-ci/engine/0.1.0/hack/deploy/install.sh | bash -s -- -h
kubeci-engine.sh - install kubeci-engine operator

kubeci-engine.sh [options]

options:
-h, --help                         show brief help
-n, --namespace=NAMESPACE          specify namespace (default: kube-system)
    --rbac                         create RBAC roles and bindings (default: true)
    --docker-registry              docker registry used to pull kubeci-engine images (default: appscode)
    --image-pull-secret            name of secret used to pull kubeci-engine operator images
    --run-on-master                run kubeci-engine operator on master
    --enable-validating-webhook    enable/disable validating webhooks for kubeci-engine crds
    --enable-mutating-webhook      enable/disable mutating webhooks for Kubernetes workloads
    --enable-status-subresource    If enabled, uses status sub resource for crds
    --enable-analytics             send usage events to Google Analytics (default: true)
    --uninstall                    uninstall kubeci-engine
    --purge                        purges kubeci-engine crd objects and crds
```

If you would like to run kubeci-engine operator pod in `master` instances, pass the `--run-on-master` flag:

```console
$ curl -fsSL https://raw.githubusercontent.com/kube-ci/engine/0.1.0/hack/deploy/install.sh \
    | bash -s -- --run-on-master [--rbac]
```

Kubeci-engine operator will be installed in a `kube-system` namespace by default. If you would like to run Kubci operator pod in `kubeci-engine` namespace, pass the `--namespace=kubeci-engine` flag:

```console
$ kubectl create namespace kubeci-engine
$ curl -fsSL https://raw.githubusercontent.com/kube-ci/engine/0.1.0/hack/deploy/install.sh \
    | bash -s -- --namespace=kubeci-engine [--run-on-master] [--rbac]
```

If you are using a private Docker registry, you need to pull the following image:

 - [kubeci/kubeci-engine](https://hub.docker.com/r/kubeci/kubeci-engine)

To pass the address of your private registry and optionally a image pull secret use flags `--docker-registry` and `--image-pull-secret` respectively.

```console
$ kubectl create namespace kubeci-engine
$ curl -fsSL https://raw.githubusercontent.com/kube-ci/engine/0.1.0/hack/deploy/install.sh \
    | bash -s -- --docker-registry=MY_REGISTRY [--image-pull-secret=SECRET_NAME] [--rbac]
```

Kubeci-engine implements [validating admission webhooks](https://kubernetes.io/docs/admin/admission-controllers/#validatingadmissionwebhook-alpha-in-18-beta-in-19) to validate Kubeci-engine CRDs. This is enabled by default for Kubernetes 1.9.0 or later releases. To disable this feature, pass the `--enable-validating-webhook=false` flag.

```console
$ curl -fsSL https://raw.githubusercontent.com/kube-ci/engine/0.1.0/hack/deploy/install.sh \
    | bash -s -- --enable-validating-webhook=false [--rbac]
```

Kubeci-engine 0.1.0 or later releases can use status sub resource for CustomResourceDefinitions. This is enabled by default for Kubernetes 1.11.0 or later releases. To disable this feature, pass the `--enable-status-subresource=false` flag.

</div>
<div class="tab-pane fade" id="helm" role="tabpanel" aria-labelledby="helm-tab">

## Using Helm
Kubeci-engine can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kube-ci/engine/tree/0.1.0/chart/kubeci) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `my-release`:

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/kubeci-engine
NAME            CHART VERSION APP VERSION DESCRIPTION
appscode/kubeci-engine  0.1.0    0.1.0  Kubeci-engine by AppsCode - Kuberenetes native CI system

$ helm install appscode/kubeci-engine --name kubeci-engine --version 0.1.0 --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kube-ci/engine/tree/master/chart/kubeci-engine).

</div>

### Installing in GKE Cluster

If you are installing Kubeci-engine on a GKE cluster, you will need cluster admin permissions to install Kubeci-engine operator. Run the following command to grant admin permission to the cluster.

```console
$ kubectl create clusterrolebinding "cluster-admin-$(whoami)" \
  --clusterrole=cluster-admin \
  --user="$(gcloud config get-value core/account)"
```


## Verify installation
To check if Kubeci-engine operator pods have started, run the following command:
```console
$ kubectl get pods --all-namespaces -l app=kubeci-engine --watch

NAMESPACE     NAME                              READY     STATUS    RESTARTS   AGE
kube-system   kubeci-engine-859d6bdb56-m9br5           2/2       Running   2          5s
```

Once the operator pods are running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:
```console
$ kubectl get crd -l app=kubeci-engine

NAME                                 AGE
workflows.engine.kube.ci             5s
workplans.engine.kube.ci             5s
workflowtemplates.engine.kube.ci     5s
```

Now, you are ready to [run your first workflow](/docs/guides/README.md) using Kubeci-engine.


## Configuring RBAC
Kubeci-engine introduces resources, such as, `Workflow`, `Workplan`, `WorkflowTemplate`,  `Trigger` and `WorkplanLog`. Kubeci-engine installer will create 2 user facing cluster roles:

| ClusterRole                 | Aggregates To | Description                            |
|-----------------------------|---------------|----------------------------------------|
| appscode:kubeci-engine:edit | admin, edit   | Allows edit access to Kubeci-engine CRDs, intended to be granted within a namespace using a RoleBinding. |
| appscode:kubeci-engine:view | view          | Allows read-only access to Kubeci-engine CRDs, intended to be granted within a namespace using a RoleBinding. |

These user facing roles supports [ClusterRole Aggregation](https://kubernetes.io/docs/admin/authorization/rbac/#aggregated-clusterroles) feature in Kubernetes 1.9 or later clusters.


## Using kubectl for Workflow
```console
# List all Workflow objects
$ kubectl get workflows --all-namespaces

# List Workflow objects for a namespace
$ kubectl get workflows -n <namespace>

# Get Workflow YAML
$ kubectl get workflow -n <namespace> <name> -o yaml

# Describe Workflow. Very useful to debug problems.
$ kubectl describe workflow -n <namespace> <name>
```


## Using kubectl for Workplan
```console
# List all Workplan objects
$ kubectl get workplans --all-namespaces

# List Workplan objects for a namespace
$ kubectl get workplans -n <namespace>

# Get Workplan YAML
$ kubectl get workplan -n <namespace> <name> -o yaml

# Describe Workplan. Very useful to debug problems.
$ kubectl describe workplan -n <namespace> <name>
```


## Detect Kubeci-engine version
To detect Kubeci-engine version, exec into the operator pod and run `kubeci-engine version` command.

```console
$ POD_NAMESPACE=kube-system
$ POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app=kubeci-engine -o jsonpath={.items[0].metadata.name})
$ kubectl exec -it $POD_NAME -c operator -n $POD_NAMESPACE kubeci-engine version

Version = 0.1.0
VersionStrategy = tag
Os = alpine
Arch = amd64
CommitHash = 85b0f16ab1b915633e968aac0ee23f877808ef49
GitBranch = release-0.1.0
GitTag = 0.1.0
CommitTimestamp = 2018-10-10T05:24:23
```