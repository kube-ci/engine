#!/bin/bash

set -x -e

apt-get update &>/dev/null
apt-get install -y git python python-pip &>/dev/null

mkdir -p $GOPATH/src/github.com/appscode
cp -r kubeci $GOPATH/src/github.com/appscode
pushd $GOPATH/src/kube.ci/git-apiserver

./hack/builddeps.sh
./hack/make.py test unit

popd
