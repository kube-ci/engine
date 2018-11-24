> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# Implicit Environment Variables

By default `HOME`(shared `HOME` directory), `NAMESPACE` (namespace of workflow/workplan/pod) and `WORKPLAN` (name of the workplan) environment variables are set to all step-containers. If user specify environment variables with same name using `workflow.spec.envVar`, they will be replaced by default values.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

## Configure RBAC

First, create a service-account for the workflow. Then, create a cluster-role with ConfigMap `list` and `watch` permissions. Now, bind it with service-accounts of both workflow and operator.

```console
$ kubectl apply -f ./docs/examples/implicit-env-var/rbac.yaml
serviceaccount/wf-sa created
clusterrole.rbac.authorization.k8s.io/wf-role created
rolebinding.rbac.authorization.k8s.io/wf-role-binding created
clusterrolebinding.rbac.authorization.k8s.io/operator-role-binding created
```

## Create Workflow

```console
$ kubectl apply -f ./docs/examples/implicit-env-var/workflow.yaml
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
  executionOrder: Serial
  allowForceTrigger: true
  steps:
  - name: step-one
    image: alpine
    commands:
    - sh
    args:
    - -c
    - echo HOME=$HOME; echo NAMESPACE=$NAMESPACE; echo WORKPLAN=$WORKPLAN;
  - name: step-two
    image: alpine
    commands:
    - sh
    args:
    - -c
    - echo HOME=$HOME; echo NAMESPACE=$NAMESPACE; echo WORKPLAN=$WORKPLAN;
```

## Trigger Workflow

Now trigger the workflow by creating a `Trigger` custom-resource which contains a complete ConfigMap resource inside `.request` section.

```console
$ kubectl apply -f ./docs/examples/implicit-env-var/trigger.yaml
trigger.extension.kube.ci/sample-trigger created
```

Whenever a workflow is triggered, a workplan is created and respective pods are scheduled.

```console
$ kubectl get workplan -l workflow=sample-workflow
NAME                    CREATED AT
sample-workflow-gwd7c   5s
```

```console
$ kubectl get pods -l workplan=sample-workflow-gwd7c
NAME                      READY   STATUS      RESTARTS   AGE
sample-workflow-gwd7c-0   0/1     Completed   0          25s
```

## Check Logs

The `step-one` and `step-two` prints the values of `HOME`, `NAMESPACE` and `WORKPLAN` environment variables.

```console
$ kubectl get --raw '/apis/extension.kube.ci/v1alpha1/namespaces/default/workplanlogs/sample-workflow-gwd7c?step=step-one'
HOME=/kubeci/home
NAMESPACE=default
WORKPLAN=sample-workflow-gwd7c
```

```console
$ kubectl get --raw '/apis/extension.kube.ci/v1alpha1/namespaces/default/workplanlogs/sample-workflow-gwd7c?step=step-two'
HOME=/kubeci/home
NAMESPACE=default
WORKPLAN=sample-workflow-gwd7c
```

Here, we can see that, `HOME`, `NAMESPACE` and `WORKPLAN` environment variables are available in both containers.

## Cleanup

```console
$ kubectl delete -f docs/examples/implicit-env-var/
serviceaccount "wf-sa" deleted
clusterrole.rbac.authorization.k8s.io "wf-role" deleted
rolebinding.rbac.authorization.k8s.io "wf-role-binding" deleted
clusterrolebinding.rbac.authorization.k8s.io "operator-role-binding" deleted
workflow.engine.kube.ci "sample-workflow" deleted
Error from server (NotFound): error when deleting "docs/examples/implicit-env-var/trigger.yaml": the server could not find the requested resource
```
