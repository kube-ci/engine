#!/usr/bin/env bash
set -xe

REPO=${GOPATH}/src/kube.ci/git-apiserver
pushd ${REPO}

export APPSCODE_ENV=dev

# codegen
./hack/codegen.sh

# make.py
./hack/make.py

# build docker
./hack/docker/setup.sh

# delete old deploy
kubectl delete deploy -n kube-system git-apiserver || true

# load to minikube
minducker kubeci/git-apiserver:initial

# deploy
./hack/deploy/install.sh
