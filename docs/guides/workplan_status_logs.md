> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# Workplan Status and Logs

This tutorial will show you how to use interactive `workplan-logs` CLI to collect logs of different steps and `workplan-viewer` web-ui to view workplan-status and logs.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md).

First create and trigger a workflow. You can follow any one of previous guides to do so. Let say, you have following workflow and workplans:

```console
$ kubectl get workflow
NAME              AGE
sample-workflow   5h
```

```console
$ kubectl get workplan
NAME                    AGE
sample-workflow-krx8p   2h
sample-workflow-v8skm   5h
```

## Workplan Logs CLI

You need to provide namespace, workplan and step using flags. For example:

```console
$ kubeci-engine workplan-logs --namespace default --workplan sample-workflow-v8skm --step step-test

fetch http://dl-cdn.alpinelinux.org/alpine/v3.8/main/x86_64/APKINDEX.tar.gz
fetch http://dl-cdn.alpinelinux.org/alpine/v3.8/community/x86_64/APKINDEX.tar.gz
(1/4) Installing nghttp2-libs (1.32.0-r0)
(2/4) Installing libssh2 (1.8.0-r3)
(3/4) Installing libcurl (7.61.1-r0)
(4/4) Installing curl (7.61.1-r0)
Executing busybox-1.28.4-r1.trigger
OK: 6 MiB in 18 packages
setting pending status...
running go tests...
ok  	github.com/diptadas/kubeci-gpig	0.001s
waiting for 30s...
setting succeed/failed status...
removing ok-to-test label...
done
```

Alternatively, we can choose them interactively:

```console
$ kubeci-engine workplan-logs

? Choose a Namespace: default
? Choose a Workflow: sample-workflow
? Choose a Workplan: sample-workflow-v8skm
? Choose a running/terminated step: step-test
fetch http://dl-cdn.alpinelinux.org/alpine/v3.8/main/x86_64/APKINDEX.tar.gz
fetch http://dl-cdn.alpinelinux.org/alpine/v3.8/community/x86_64/APKINDEX.tar.gz
(1/4) Installing nghttp2-libs (1.32.0-r0)
(2/4) Installing libssh2 (1.8.0-r3)
(3/4) Installing libcurl (7.61.1-r0)
(4/4) Installing curl (7.61.1-r0)
Executing busybox-1.28.4-r1.trigger
OK: 6 MiB in 18 packages
setting pending status...
running go tests...
ok  	github.com/diptadas/kubeci-gpig	0.001s
waiting for 30s...
setting succeed/failed status...
removing ok-to-test label...
done
```

## Workplan Viewer

The web-ui is deployed as a sidecar container during installation process. You can expose it with port forwarding. The status page will refresh after every 30 seconds. For `Running` and `Terminated` steps, there are links for accessing logs.

- Status URL: `http://127.0.0.1:9090/namespaces/{namespace}/workplans/{workplan-name}`
- Logs URL: `http://127.0.0.1:9090/namespaces/{namespace}/workplans/{workplan-name}/steps/{step-name}`

First, get the operator pod:

```console
$ kubectl get pods --all-namespaces | grep kubeci-engine
kube-system   kubeci-engine-5944797bb5-qxn5n                 3/3     Running     0          65m
```

Now, forward `9090` port of operator pod:

```console
$ kubectl port-forward -n kube-system kubeci-engine-5944797bb5-qxn5n 9090:9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Go to following URL to get current of status workplan `sample-workflow-v8skm`:

`http://127.0.0.1:9090/namespaces/default/workplans/sample-workflow-v8skm`

Go to following URL to get logs of step `step-test`:

`http://127.0.0.1:9090/namespaces/default/workplans/sample-workflow-v8skm/steps/step-test`
