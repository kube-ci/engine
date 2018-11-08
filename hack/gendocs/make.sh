#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kube-ci/engine/hack/gendocs
go run main.go
popd
