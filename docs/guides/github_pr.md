> New to KubeCI engine? Please start [here](/docs/concepts/README.md).

# Github Pull Request

This tutorial will show you how to use KubeCI engine and [Git API server](https://github.com/kube-ci/git-apiserver) to run tests and update commit status based on labels of a Github pull-request.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Now, install KubeCI engine in your cluster following the steps [here](/docs/setup/install.md). Also, install git-apiserver following the steps [here](https://github.com/kube-ci/git-apiserver).

## Configure Github Webhook

First, configure github webhook to send POST request of events to your Kubernetes apiserver. Also enable events for pull-requests and disable SSL verification.

Payload URL: `https://{master-ip}/apis/webhook.git.kube.ci/v1alpha1/githubevents`

## Configure Repository CRD

Now, create a repository CRD specifying the repository URL.

```console
$ kubectl apply -f ./docs/examples/github-pr/repository.yaml
repository.git.kube.ci/kubeci-gpig created
```

```yaml
apiVersion: git.kube.ci/v1alpha1
kind: Repository
metadata:
  name: kubeci-gpig
  namespace: default
spec:
  host: github
  owner: diptadas
  repo: kubeci-gpig
  cloneUrl: https://github.com/diptadas/kubeci-gpig.git
```

## Create Github Secret

Create a secret with Github API token required for updating commit status.

```console
$ kubectl create secret generic github-credential --from-literal=TOKEN={github-api-token}
secret github-credential created
```

## Configure RBAC

Create a service-account for the workflow. Then, create a cluster-role with `PullRequest` `list` and `watch` permissions. Also need Secret Create and Get permissions for environment from json-path and github-secret. Now, bind the cluster-role with service-accounts of both workflow and operator.

```console
$ kubectl apply -f ./docs/examples/github-pr/rbac.yaml
serviceaccount/wf-sa created
clusterrole.rbac.authorization.k8s.io/wf-role created
rolebinding.rbac.authorization.k8s.io/wf-role-binding created
clusterrolebinding.rbac.authorization.k8s.io/operator-role-binding created
```

## Create Workflow

```console
$ kubectl apply -f ./docs/examples/github-pr/workflow.yaml
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
  - apiVersion: git.kube.ci/v1alpha1
    kind: PullRequest
    resource: pullrequests
    namespace: default
    selector:
      matchLabels:
        repository: kubeci-gpig
        state: open
        ok-to-test:
    onCreateOrUpdate: true
    onDelete: false
    envFromPath:
      HEAD_SHA: '{$.spec.headSHA}'
      PR_NUMBER: '{$.spec.number}'
  serviceAccount: wf-sa
  envFrom:
  - secretRef:
      name: github-credential
  steps:
  - name: step-clone
    image: alpine/git
    commands:
    - sh
    args:
    - -c
    - git clone https://github.com/diptadas/kubeci-gpig.git .; git checkout $HEAD_SHA
  - name: step-test
    image: golang:1.10-alpine3.8
    commands:
    - sh
    args:
    - -c
    - cat kubeci/test.sh | sh
```

- The trigger section specifies a `PullRequest` CRD with `repository`, `state` and `ok-to-test` labels.
- The `step-clone` clones the repository and checkouts to specific SHA.
- The `step-test` runs a script already present in the repository. This script runs go-test and update commit status based on output. [Here](/docs/examples/github-pr/test.sh) is the sample script we used in this case. The `TARGET_URL` in the script points to kubeci web-ui which provides workplan status and logs. The web-ui is deployed as a sidecar container during installation process. You can expose it with port forwarding:

```console
$ kubectl port-forward -n kube-system {operator-pod-name} 9090:9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

## Trigger Workflow

To trigger the workflow, create a pull request in your repository and set `ok-to-test` label on it. This will send a POST request to Kubernetes apiserver. Git-apiserver controller will respond to it and create a `PullRequest` CRD with with `repository`, `state` and `ok-to-test` labels, which will then trigger the workflow.