#!/usr/bin/env bash

set -eoux pipefail

ORG_NAME=kube-ci
REPO_NAME=engine
APP_LABEL=kubeci-engine #required for `kubectl describe deploy -n kube-system -l app=$APP_LABEL`
REPO_ROOT=$GOPATH/src/kube.ci/$REPO_NAME

export APPSCODE_ENV=dev
export DOCKER_REGISTRY=appscodeci

# get concourse-common
pushd $REPO_NAME
git status # required, otherwise you'll get error `Working tree has modifications.  Cannot add.`. why?
git subtree pull --prefix hack/libbuild https://github.com/appscodelabs/libbuild.git master --squash -m 'concourse'
popd

source $REPO_NAME/hack/libbuild/concourse/init.sh

pushd $GOPATH/src/kube.ci/$REPO_NAME

# install dependencies
./hack/builddeps.sh

# build and push docker images
./hack/docker/setup.sh
./hack/docker/setup.sh push
./hack/deploy/install.sh --docker-registry=$DOCKER_REGISTRY

# run tests
./hack/make.py test e2e --selfhosted-operator

popd
