## Development Guide
This document is intended to be the canonical source of truth for things like supported toolchain versions for building KubeCI engine.
If you find a requirement that this doc does not capture, please submit an issue on github.

This document is intended to be relative to the branch in which it is found. It is guaranteed that requirements will change over time
for the development branch, but release branches of KubeCI engine should not change.

### Build KubeCI engine
Some of the KubeCI engine development helper scripts rely on a fairly up-to-date GNU tools environment, so most recent Linux distros should
work just fine out-of-the-box.

#### Setup GO
KubeCI engine is written in Google's GO programming language. Currently, KubeCI engine is developed and tested on **go 1.10**. If you haven't set up a GO
development environment, please follow [these instructions](https://golang.org/doc/code.html) to install GO.

#### Download Source

```console
$ cd $(go env GOPATH)/src/kube.ci
$ git clone https://github.com/kube-ci/engine.git
$ cd engine
```

#### Install Dev tools
To install various dev tools for KubeCI engine, run the following command:
```console
$ ./hack/builddeps.sh
```

#### Build Binary
```
$ ./hack/make.py
$ kubeci-engine version
```

#### Run Binary Locally
```console
$ kubeci-engine run \
  --secure-port=8443 \
  --kubeconfig="$HOME/.kube/config" \
  --authorization-kubeconfig="$HOME/.kube/config" \
  --authentication-kubeconfig="$HOME/.kube/config" \
  --authentication-skip-lookup
```

#### Dependency management
KubeCI engine uses [Glide](https://github.com/Masterminds/glide) to manage dependencies. Dependencies are already checked in the `vendor` folder.
If you want to update/add dependencies, run:
```console
$ glide slow
```

#### Build Docker images
To build and push your custom Docker image, follow the steps below. To release a new version of KubeCI engine, please follow the [release guide](/docs/setup/developer-guide/release.md).

```console
# Build Docker image
$ ./hack/docker/setup.sh; ./hack/docker/setup.sh push

# Add docker tag for your repository
$ docker tag kubeci/kubeci-engine:<tag> <image>:<tag>

# Push Image
$ docker push <image>:<tag>
```

#### Generate CLI Reference Docs
```console
$ ./hack/gendocs/make.sh
```

### Testing KubeCI engine
#### Unit tests
```console
$ ./hack/make.py test unit
```

#### Run e2e tests
KubeCI engine uses [Ginkgo](http://onsi.github.io/ginkgo/) to run e2e tests.
```console
$ ./hack/make.py test e2e
```

To run e2e tests against remote backends, you need to set cloud provider credentials in `./hack/config/.env`. You can see an example file in `./hack/config/.env.example`.
