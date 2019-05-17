[![Go Report Card](https://goreportcard.com/badge/github.com/kube-ci/engine)](https://goreportcard.com/report/github.com/kube-ci/engine)
[![Build Status](https://travis-ci.org/kube-ci/engine.svg?branch=master)](https://travis-ci.org/kube-ci/engine)
[![codecov](https://codecov.io/gh/kube-ci/engine/branch/master/graph/badge.svg)](https://codecov.io/gh/kube-ci/engine)
[![Docker Pulls](https://img.shields.io/docker/pulls/kubeci/kubeci-engine.svg)](https://hub.docker.com/r/kubeci/kubeci-engine/)
[![Slack](https://slack.appscode.com/badge.svg)](https://slack.appscode.com)
[![Twitter](https://img.shields.io/twitter/follow/thekubeci.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=TheKubeCi)

# KubeCI Engine

KubeCI engine by AppsCode is a Kubernetes native workflow engine.

## Features

- Configure a set of containerized steps using workflow.
- Run steps in serial or, parallel order by resolving dependencies for each step.
- Trigger workflows through create/update/delete events of any kubernetes object.
- Trigger workflows manually with fake create events.
- Shared `workspace` and `home` directory among all steps of a workflow.
- Credential initializer for Docker and Git.
- APIs for collecting status and logs of each step.

## Supported Versions

Please pick a version of KubeCI engine that matches your Kubernetes installation.

| KubeCI engine Version                                                                      | Docs                                                            | Kubernetes Version |
|------------------------------------------------------------------------------------|-----------------------------------------------------------------|--------------------|
| [0.1.0](https://github.com/kube-ci/engine/releases/tag/0.1.0) (uses CRD) | [User Guide](https://kube.ci/products/engine/0.1.0)    | 1.9.x+             |

## Installation

To install KubeCI engine, please follow the guide [here](https://kube.ci/products/engine/0.1.0/setup/install).

## Using KubeCI engine

Want to learn how to use KubeCI engine? Please start [here](https://kube.ci/products/engine/0.1.0).

## KubeCI engine API Clients

You can use KubeCI engine api clients to programmatically access its objects. Here are the supported clients:

- Go: [https://github.com/kube-ci/engine](/client/clientset/versioned)

## Contribution guidelines

Want to help improve KubeCI engine? Please start [here](https://kube.ci/products/engine/0.1.0/welcome/contributing).

---

**KubeCI binaries collects anonymous usage statistics to help us learn how the software is being used and how we can improve it. To disable stats collection, run the operator with the flag** `--enable-analytics=false`.

---

## Acknowledgement

 - Credential initializer part is adopted from [knative/build](https://github.com/knative/build).

## Support

We use Slack for public discussions. To chit chat with us or the rest of the community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8NCX6N23/details/) channel `#kubeci`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

If you have found a bug with KubeCI engine or want to request for new features, please [file an issue](https://github.com/kube-ci/project/issues/new).
