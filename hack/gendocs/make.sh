#!/usr/bin/env bash

pushd $GOPATH/src/kube.ci/engine/hack/gendocs
go run main.go
popd
