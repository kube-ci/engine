#!/bin/bash

set -x

GOPATH=$(go env GOPATH)
PACKAGE_NAME=kube.ci/engine
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"
apiGroups=(engine/v1alpha1 trigger/v1alpha1)

pushd $REPO_ROOT

mkdir -p "$REPO_ROOT"/api/api-rules

# for EAS types
# only deepcopy for webhook
docker run --rm -ti -u $(id -u):$(id -g) \
  -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
  -w "$DOCKER_REPO_ROOT" \
  appscode/gengo:release-1.12 "$DOCKER_CODEGEN_PKG"/generate-internal-groups.sh "deepcopy,defaulter,conversion" \
  kube.ci/engine/client \
  kube.ci/engine/apis \
  kube.ci/engine/apis \
  trigger:v1alpha1 \
  --go-header-file "$DOCKER_REPO_ROOT/hack/gengo/boilerplate.go.txt"

# for both CRD and EAS types
docker run --rm -ti -u $(id -u):$(id -g) \
  -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
  -w "$DOCKER_REPO_ROOT" \
  appscode/gengo:release-1.12 "$DOCKER_CODEGEN_PKG"/generate-groups.sh all \
  kube.ci/engine/client \
  kube.ci/engine/apis \
  "engine:v1alpha1 trigger:v1alpha1" \
  --go-header-file "$DOCKER_REPO_ROOT/hack/gengo/boilerplate.go.txt"

# Generate openapi
for gv in "${apiGroups[@]}"; do
  docker run --rm -ti -u $(id -u):$(id -g) \
    -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
    -w "$DOCKER_REPO_ROOT" \
    appscode/gengo:release-1.12 openapi-gen \
    --v 1 --logtostderr \
    --go-header-file "hack/gengo/boilerplate.go.txt" \
    --input-dirs "$PACKAGE_NAME/apis/${gv},k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/version,k8s.io/api/core/v1,github.com/appscode/go/encoding/json/types" \
    --output-package "$PACKAGE_NAME/apis/${gv}" \
    --report-filename api/api-rules/violation_exceptions.list
done

# Generate crds.yaml and swagger.json
go run ./hack/gencrd/main.go

popd
