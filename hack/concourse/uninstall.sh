#!/usr/bin/env bash

set -x

./hack/deploy/install.sh --docker-registry=$DOCKER_REGISTRY --uninstall --purge
