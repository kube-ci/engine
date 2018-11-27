# Uninstall KubeCI engine

To uninstall KubeCI engine, run the following command:

```console
$ curl -fsSL https://raw.githubusercontent.com/kube-ci/engine/0.1.0/hack/deploy/install.sh \
    | bash -s -- --uninstall [--namespace=NAMESPACE]

validatingwebhookconfiguration.admissionregistration.k8s.io "admission.engine.kube.ci" deleted
No resources found
apiservice.apiregistration.k8s.io "v1alpha1.admission.engine.kube.ci" deleted
apiservice.apiregistration.k8s.io "v1alpha1.extensions.kube.ci" deleted
deployment.extensions "kubeci-engine" deleted
service "kubeci-engine" deleted
secret "kubeci-engine-apiserver-cert" deleted
serviceaccount "kubeci-engine" deleted
clusterrolebinding.rbac.authorization.k8s.io "kubeci-engine" deleted
clusterrolebinding.rbac.authorization.k8s.io "kubeci-engine-apiserver-auth-delegator" deleted
clusterrole.rbac.authorization.k8s.io "kubeci-engine" deleted
rolebinding.rbac.authorization.k8s.io "kubeci-engine-apiserver-extension-server-authentication-reader" deleted
No resources found
waiting for kubeci-engine operator pod to stop running

Successfully uninstalled KUBECI-ENGINE!
```

The above command will leave the KubeCI engine crd objects as-is. If you wish to **nuke** all KubeCI engine crd objects, also pass the `--purge` flag. This will keep a copy of KubeCI engine crd objects in your current directory.