#!/bin/bash
set -eou pipefail

crds=(repositories)

echo "checking kubeconfig context"
kubectl config current-context || {
  echo "Set a context (kubectl use-context <context>) out of the following:"
  echo
  kubectl config get-contexts
  exit 1
}
echo ""

# http://redsymbol.net/articles/bash-exit-traps/
function cleanup() {
  rm -rf $ONESSL ca.crt ca.key server.crt server.key
}
trap cleanup EXIT

# ref: https://github.com/appscodelabs/libbuild/blob/master/common/lib.sh#L55
inside_git_repo() {
  git rev-parse --is-inside-work-tree >/dev/null 2>&1
  inside_git=$?
  if [ "$inside_git" -ne 0 ]; then
    echo "Not inside a git repository"
    exit 1
  fi
}

detect_tag() {
  inside_git_repo

  # http://stackoverflow.com/a/1404862/3476121
  git_tag=$(git describe --exact-match --abbrev=0 2>/dev/null || echo '')

  commit_hash=$(git rev-parse --verify HEAD)
  git_branch=$(git rev-parse --abbrev-ref HEAD)
  commit_timestamp=$(git show -s --format=%ct)

  if [ "$git_tag" != '' ]; then
    TAG=$git_tag
    TAG_STRATEGY='git_tag'
  elif [ "$git_branch" != 'master' ] && [ "$git_branch" != 'HEAD' ] && [[ "$git_branch" != release-* ]]; then
    TAG=$git_branch
    TAG_STRATEGY='git_branch'
  else
    hash_ver=$(git describe --tags --always --dirty)
    TAG="${hash_ver}"
    TAG_STRATEGY='commit_hash'
  fi

  export TAG
  export TAG_STRATEGY
  export git_tag
  export git_branch
  export commit_hash
  export commit_timestamp
}

# https://stackoverflow.com/a/677212/244009
if [ -x "$(command -v onessl)" ]; then
  export ONESSL=onessl
else
  # ref: https://stackoverflow.com/a/27776822/244009
  case "$(uname -s)" in
    Darwin)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-darwin-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    Linux)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-linux-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    CYGWIN* | MINGW32* | MSYS*)
      curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-windows-amd64.exe
      chmod +x onessl.exe
      export ONESSL=./onessl.exe
      ;;
    *)
      echo 'other OS'
      ;;
  esac
fi

# ref: https://stackoverflow.com/a/7069755/244009
# ref: https://jonalmeida.com/posts/2013/05/26/different-ways-to-implement-flags-in-bash/
# ref: http://tldp.org/LDP/abs/html/comparison-ops.html

export GIT_APISERVER_NAMESPACE=kube-system
export GIT_APISERVER_SERVICE_ACCOUNT=git-apiserver
export GIT_APISERVER_ENABLE_RBAC=true
export GIT_APISERVER_RUN_ON_MASTER=0
export GIT_APISERVER_ENABLE_VALIDATING_WEBHOOK=false
export GIT_APISERVER_ENABLE_MUTATING_WEBHOOK=false
export GIT_APISERVER_DOCKER_REGISTRY=kubeci
export GIT_APISERVER_IMAGE_TAG=0.7.0
export GIT_APISERVER_IMAGE_PULL_SECRET=
export GIT_APISERVER_IMAGE_PULL_POLICY=IfNotPresent
export GIT_APISERVER_ENABLE_STATUS_SUBRESOURCE=false
export GIT_APISERVER_ENABLE_ANALYTICS=true
export GIT_APISERVER_UNINSTALL=0
export GIT_APISERVER_PURGE=0

export APPSCODE_ENV=${APPSCODE_ENV:-prod}
export SCRIPT_LOCATION="curl -fsSL https://raw.githubusercontent.com/appscode/kubeci/0.7.0/"
if [ "$APPSCODE_ENV" = "dev" ]; then
  detect_tag
  export SCRIPT_LOCATION="cat "
  export GIT_APISERVER_IMAGE_TAG=$TAG
  export GIT_APISERVER_IMAGE_PULL_POLICY=IfNotPresent
fi

KUBE_APISERVER_VERSION=$(kubectl version -o=json | $ONESSL jsonpath '{.serverVersion.gitVersion}')
$ONESSL semver --check='<1.9.0' $KUBE_APISERVER_VERSION || {
  export GIT_APISERVER_ENABLE_VALIDATING_WEBHOOK=true
  export GIT_APISERVER_ENABLE_MUTATING_WEBHOOK=true
}
$ONESSL semver --check='<1.11.0' $KUBE_APISERVER_VERSION || { export GIT_APISERVER_ENABLE_STATUS_SUBRESOURCE=true; }

show_help() {
  echo "kubeci.sh - install kubeci operator"
  echo " "
  echo "kubeci.sh [options]"
  echo " "
  echo "options:"
  echo "-h, --help                         show brief help"
  echo "-n, --namespace=NAMESPACE          specify namespace (default: kube-system)"
  echo "    --rbac                         create RBAC roles and bindings (default: true)"
  echo "    --docker-registry              docker registry used to pull kubeci images (default: appscode)"
  echo "    --image-pull-secret            name of secret used to pull kubeci operator images"
  echo "    --run-on-master                run kubeci operator on master"
  echo "    --enable-validating-webhook    enable/disable validating webhooks for KUBECI crds"
  echo "    --enable-mutating-webhook      enable/disable mutating webhooks for Kubernetes workloads"
  echo "    --enable-status-subresource    If enabled, uses status sub resource for crds"
  echo "    --enable-analytics             send usage events to Google Analytics (default: true)"
  echo "    --uninstall                    uninstall kubeci"
  echo "    --purge                        purges kubeci crd objects and crds"
}

while test $# -gt 0; do
  case "$1" in
    -h | --help)
      show_help
      exit 0
      ;;
    -n)
      shift
      if test $# -gt 0; then
        export GIT_APISERVER_NAMESPACE=$1
      else
        echo "no namespace specified"
        exit 1
      fi
      shift
      ;;
    --namespace*)
      export GIT_APISERVER_NAMESPACE=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --docker-registry*)
      export GIT_APISERVER_DOCKER_REGISTRY=$(echo $1 | sed -e 's/^[^=]*=//g')
      shift
      ;;
    --image-pull-secret*)
      secret=$(echo $1 | sed -e 's/^[^=]*=//g')
      export GIT_APISERVER_IMAGE_PULL_SECRET="name: '$secret'"
      shift
      ;;
    --enable-validating-webhook*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export GIT_APISERVER_ENABLE_VALIDATING_WEBHOOK=false
      fi
      shift
      ;;
    --enable-mutating-webhook*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export GIT_APISERVER_ENABLE_MUTATING_WEBHOOK=false
      fi
      shift
      ;;
    --enable-status-subresource*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export GIT_APISERVER_ENABLE_STATUS_SUBRESOURCE=false
      fi
      shift
      ;;
    --enable-analytics*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export GIT_APISERVER_ENABLE_ANALYTICS=false
      fi
      shift
      ;;
    --rbac*)
      val=$(echo $1 | sed -e 's/^[^=]*=//g')
      if [ "$val" = "false" ]; then
        export GIT_APISERVER_SERVICE_ACCOUNT=default
        export GIT_APISERVER_ENABLE_RBAC=false
      fi
      shift
      ;;
    --run-on-master)
      export GIT_APISERVER_RUN_ON_MASTER=1
      shift
      ;;
    --uninstall)
      export GIT_APISERVER_UNINSTALL=1
      shift
      ;;
    --purge)
      export GIT_APISERVER_PURGE=1
      shift
      ;;
    *)
      show_help
      exit 1
      ;;
  esac
done

if [ "$GIT_APISERVER_UNINSTALL" -eq 1 ]; then
  # delete webhooks and apiservices
  kubectl delete validatingwebhookconfiguration -l app=kubeci || true
  kubectl delete mutatingwebhookconfiguration -l app=kubeci || true
  kubectl delete apiservice -l app=kubeci
  # delete kubeci operator
  kubectl delete deployment -l app=kubeci --namespace $GIT_APISERVER_NAMESPACE
  kubectl delete service -l app=kubeci --namespace $GIT_APISERVER_NAMESPACE
  kubectl delete secret -l app=kubeci --namespace $GIT_APISERVER_NAMESPACE
  # delete RBAC objects, if --rbac flag was used.
  kubectl delete serviceaccount -l app=kubeci --namespace $GIT_APISERVER_NAMESPACE
  kubectl delete clusterrolebindings -l app=kubeci
  kubectl delete clusterrole -l app=kubeci
  kubectl delete rolebindings -l app=kubeci --namespace $GIT_APISERVER_NAMESPACE
  kubectl delete role -l app=kubeci --namespace $GIT_APISERVER_NAMESPACE

  echo "waiting for kubeci operator pod to stop running"
  for (( ; ; )); do
    pods=($(kubectl get pods --namespace $GIT_APISERVER_NAMESPACE -l app=kubeci -o jsonpath='{range .items[*]}{.metadata.name} {end}'))
    total=${#pods[*]}
    if [ $total -eq 0 ]; then
      break
    fi
    sleep 2
  done

  # https://github.com/kubernetes/kubernetes/issues/60538
  if [ "$GIT_APISERVER_PURGE" -eq 1 ]; then
    for crd in "${crds[@]}"; do
      pairs=($(kubectl get ${crd}.git.kube.ci --all-namespaces -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.namespace} {end}' || true))
      total=${#pairs[*]}

      # save objects
      if [ $total -gt 0 ]; then
        echo "dumping ${crd} objects into ${crd}.yaml"
        kubectl get ${crd}.git.kube.ci --all-namespaces -o yaml >${crd}.yaml
      fi

      for ((i = 0; i < $total; i += 2)); do
        name=${pairs[$i]}
        namespace=${pairs[$i + 1]}
        # delete crd object
        echo "deleting ${crd} $namespace/$name"
        kubectl delete ${crd}.git.kube.ci $name -n $namespace
      done

      # delete crd
      kubectl delete crd ${crd}.git.kube.ci || true
    done

    # delete user roles
    kubectl delete clusterroles appscode:kubeci:edit appscode:kubeci:view
  fi

  echo
  echo "Successfully uninstalled KUBECI!"
  exit 0
fi

echo "checking whether extended apiserver feature is enabled"
$ONESSL has-keys configmap --namespace=kube-system --keys=requestheader-client-ca-file extension-apiserver-authentication || {
  echo "Set --requestheader-client-ca-file flag on Kubernetes apiserver"
  exit 1
}
echo ""

export KUBE_CA=
export GIT_APISERVER_ENABLE_APISERVER=false
if [ "$GIT_APISERVER_ENABLE_VALIDATING_WEBHOOK" = true ] || [ "$GIT_APISERVER_ENABLE_MUTATING_WEBHOOK" = true ]; then
  $ONESSL get kube-ca >/dev/null 2>&1 || {
    echo "Admission webhooks can't be used when kube apiserver is accesible without verifying its TLS certificate (insecure-skip-tls-verify : true)."
    echo
    exit 1
  }
  export KUBE_CA=$($ONESSL get kube-ca | $ONESSL base64)
  export GIT_APISERVER_ENABLE_APISERVER=true
fi

env | sort | grep KUBECI*
echo ""

# create necessary TLS certificates:
# - a local CA key and cert
# - a webhook server key and cert signed by the local CA
$ONESSL create ca-cert
$ONESSL create server-cert server --domains=git-apiserver.$GIT_APISERVER_NAMESPACE.svc
export SERVICE_SERVING_CERT_CA=$(cat ca.crt | $ONESSL base64)
export TLS_SERVING_CERT=$(cat server.crt | $ONESSL base64)
export TLS_SERVING_KEY=$(cat server.key | $ONESSL base64)

${SCRIPT_LOCATION}hack/deploy/operator.yaml | $ONESSL envsubst | kubectl apply -f -

if [ "$GIT_APISERVER_ENABLE_RBAC" = true ]; then
  ${SCRIPT_LOCATION}hack/deploy/service-account.yaml | $ONESSL envsubst | kubectl apply -f -
  ${SCRIPT_LOCATION}hack/deploy/rbac-list.yaml | $ONESSL envsubst | kubectl auth reconcile -f -
  ${SCRIPT_LOCATION}hack/deploy/user-roles.yaml | $ONESSL envsubst | kubectl auth reconcile -f -
fi

if [ "$GIT_APISERVER_RUN_ON_MASTER" -eq 1 ]; then
  kubectl patch deploy git-apiserver -n $GIT_APISERVER_NAMESPACE \
    --patch="$(${SCRIPT_LOCATION}hack/deploy/run-on-master.yaml)"
fi

if [ "$GIT_APISERVER_ENABLE_APISERVER" = true ]; then
  ${SCRIPT_LOCATION}hack/deploy/apiservices.yaml | $ONESSL envsubst | kubectl apply -f -
fi
if [ "$GIT_APISERVER_ENABLE_VALIDATING_WEBHOOK" = true ]; then
  ${SCRIPT_LOCATION}hack/deploy/validating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
fi
# if [ "$GIT_APISERVER_ENABLE_MUTATING_WEBHOOK" = true ]; then
#   ${SCRIPT_LOCATION}hack/deploy/mutating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
# fi

echo
echo "waiting until kubeci operator deployment is ready"
$ONESSL wait-until-ready deployment git-apiserver --namespace $GIT_APISERVER_NAMESPACE || {
  echo "KUBECI operator deployment failed to be ready"
  exit 1
}

if [ "$GIT_APISERVER_ENABLE_APISERVER" = true ]; then
  echo "waiting until kubeci apiservice is available"
  $ONESSL wait-until-ready apiservice v1alpha1.admission.git.kube.ci || {
    echo "KUBECI apiservice failed to be ready"
    exit 1
  }
fi

echo "waiting until kubeci crds are ready"
for crd in "${crds[@]}"; do
  $ONESSL wait-until-ready crd ${crd}.git.kube.ci || {
    echo "$crd crd failed to be ready"
    exit 1
  }
done

echo
echo "Successfully installed KUBECI in $GIT_APISERVER_NAMESPACE namespace!"
