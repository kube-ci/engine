> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# Volumes

This tutorial will show you how to specify explicit volumes and mount them inside step-containers using volume-mounts.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

## Configure RBAC

First, create a service-account for the workflow. Then, create a cluster-role with ConfigMap `list` and `watch` permissions. Now, bind it with service-accounts of both workflow and operator.

```console
$ kubectl apply -f ./docs/examples/volumes/rbac.yaml
serviceaccount/wf-sa created
clusterrole.rbac.authorization.k8s.io/wf-role created
rolebinding.rbac.authorization.k8s.io/wf-role-binding created
clusterrolebinding.rbac.authorization.k8s.io/operator-role-binding created
```

## Create Workflow

```console
$ kubectl apply -f ./docs/examples/volumes/workflow.yaml
workflow.engine.kube.ci/wf-volumes created
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
  allowManualTrigger: true
  steps:
  - name: step-one
    image: alpine
    commands:
    - touch
    args:
    - /path-one/file-01
    volumeMounts: # explicit volume mounts
    - name: sample-volume
      mountPath: /path-one
  - name: step-two
    image: alpine
    commands:
    - ls
    args:
    - /path-two
    volumeMounts: # explicit volume mounts
    - name: sample-volume
      mountPath: /path-two
  volumes: # explicit volumes
  - name: sample-volume
    hostPath:
      path: /tmp/sample-volume
      type: DirectoryOrCreate
```

## Trigger Workflow

Now trigger the workflow by creating a `Trigger` custom-resource which contains a complete ConfigMap resource inside `.request` section.

```console
$ kubectl apply -f ./docs/examples/volumes/trigger.yaml
trigger.extensions.kube.ci/sample-trigger created
```

Whenever a workflow is triggered, a workplan is created and respective pods are scheduled.

```console
$ kubectl get workplan -l workflow=sample-workflow
NAME                    CREATED AT
sample-workflow-gmjrl   13s
```

```console
$ kubectl get pods -l workplan=sample-workflow-gmjrl
NAME                      READY   STATUS      RESTARTS   AGE
sample-workflow-gmjrl-0   0/1     Completed   0          29s
```

## Check Logs

Both `step-one` and `step-two` mounts the same hostpath volume but in different paths. So the content created by `step-one` in path `/path-one` is also available to `step-two` in path `/path-two`.

```console
$ kubectl get --raw '/apis/extensions.kube.ci/v1alpha1/namespaces/default/workplanlogs/sample-workflow-gmjrl?step=step-two'
file-01
```

## Cleanup

```console
$ kubectl delete -f docs/examples/volumes/
serviceaccount "wf-sa" deleted
clusterrole.rbac.authorization.k8s.io "wf-role" deleted
rolebinding.rbac.authorization.k8s.io "wf-role-binding" deleted
clusterrolebinding.rbac.authorization.k8s.io "operator-role-binding" deleted
workflow.engine.kube.ci "sample-workflow" deleted
Error from server (NotFound): error when deleting "docs/examples/force-trigger/trigger.yaml": the server could not find the requested resource
```
