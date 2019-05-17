package controller

import (
	"time"

	cs "github.com/kube-ci/engine/client/clientset/versioned"
	kubeci_informers "github.com/kube-ci/engine/client/informers/externalversions"
	"github.com/kube-ci/engine/pkg/eventer"
	core "k8s.io/api/core/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	reg_util "kmodules.xyz/client-go/admissionregistration/v1beta1"
)

const (
	validatingWebhook = "validators.engine.kube.ci"
	mutatingWebhook   = "mutators.engine.kube.ci"
)

type config struct {
	EnableRBAC        bool
	KubeciImageTag    string
	DockerRegistry    string
	MaxNumRequeues    int
	NumThreads        int
	ResyncPeriod      time.Duration
	DiscoveryInterval time.Duration
}

type Config struct {
	config

	ClientConfig *rest.Config
	KubeClient   kubernetes.Interface
	KubeciClient cs.Interface
	CRDClient    crd_cs.ApiextensionsV1beta1Interface
}

func NewConfig(clientConfig *rest.Config) *Config {
	return &Config{
		ClientConfig: clientConfig,
	}
}

func (c *Config) New() (*Controller, error) {
	tweakListOptions := func(opt *metav1.ListOptions) {}
	ctrl := &Controller{
		config:                c.config,
		clientConfig:          c.ClientConfig,
		kubeClient:            c.KubeClient,
		kubeciClient:          c.KubeciClient,
		crdClient:             c.CRDClient,
		kubeInformerFactory:   informers.NewFilteredSharedInformerFactory(c.KubeClient, c.ResyncPeriod, core.NamespaceAll, tweakListOptions),
		kubeciInformerFactory: kubeci_informers.NewSharedInformerFactory(c.KubeciClient, c.ResyncPeriod),
		recorder:              eventer.NewEventRecorder(c.KubeClient, "kubeci-controller"),
		observedWorkflows:     observedWorkflows{items: make(map[string]workflowState)},
		dynamicInformers:      dynamicInformers{items: make(map[string]informerStore)},
	}

	if err := ctrl.ensureCustomResourceDefinitions(); err != nil {
		return nil, err
	}
	if err := reg_util.UpdateValidatingWebhookCABundle(ctrl.clientConfig, validatingWebhook); err != nil {
		return nil, err
	}
	if err := reg_util.UpdateMutatingWebhookCABundle(ctrl.clientConfig, mutatingWebhook); err != nil {
		return nil, err
	}

	ctrl.initWorkflowWatcher()
	ctrl.initWorkplanWatcher()
	ctrl.initDynamicWatcher()
	ctrl.initWorkflowTemplateWatcher()

	return ctrl, nil
}
