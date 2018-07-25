#!/usr/bin/env bash
set -xe

REPO=${GOPATH}/src/kube.ci/kubeci
pushd ${REPO}

export APPSCODE_ENV=dev

# codegen
./hack/codegen.sh

# make.py
./hack/make.py

# build docker
./hack/docker/setup.sh

# load to minikube
minducker kubeci/kubeci:initial

# deploy
./hack/deploy/install.sh
