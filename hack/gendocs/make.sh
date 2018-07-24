#!/usr/bin/env bash

pushd $GOPATH/src/kube.ci/kubeci/hack/gendocs
go run main.go
popd
