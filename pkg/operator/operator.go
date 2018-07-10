package operator

import (
	"time"

	"github.com/appscode/go/log"
	apiext_util "github.com/appscode/kutil/apiextensions/v1beta1"
	dynamicclientset "github.com/appscode/kutil/dynamic/clientset"
	dynamicinformer "github.com/appscode/kutil/dynamic/informer"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kube-ci/experiments/apis/kubeci/v1alpha1"
	cs "github.com/kube-ci/experiments/client/clientset/versioned"
	api_informer "github.com/kube-ci/experiments/client/informers/externalversions"
	api_listers "github.com/kube-ci/experiments/client/listers/kubeci/v1alpha1"
	"github.com/kube-ci/experiments/pkg/eventer"
	kext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	kext_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Operator struct {
	// configs
	ClientConfig   *rest.Config
	ResyncPeriod   time.Duration
	WatchNamespace string
	MaxNumRequeues int
	NumThreads     int

	KubeClient kubernetes.Interface
	CRDClient  kext_cs.ApiextensionsV1beta1Interface
	ApiClient  cs.Interface

	ApiInformerFactory api_informer.SharedInformerFactory
	recorder           record.EventRecorder

	wfQueue    *queue.Worker
	wfInformer cache.SharedIndexInformer
	wfLister   api_listers.WorkflowLister

	wpQueue    *queue.Worker
	wpInformer cache.SharedIndexInformer
	wpLister   api_listers.WorkplanLister

	dynClient           *dynamicclientset.Clientset
	dynInformersFactory *dynamicinformer.SharedInformerFactory

	// TODO: close unused informers
	// only one informer is created for a specific resource (among all workflows)
	// we should close a informer when no workflow need that informer (when workflows deleted or updated)
}

func (op *Operator) ensureCustomResourceDefinitions() error {
	log.Infof("Ensuring CRD registration")

	crds := []*kext.CustomResourceDefinition{
		api.Workflow{}.CustomResourceDefinition(),
		api.Workplan{}.CustomResourceDefinition(),
	}
	return apiext_util.RegisterCRDs(op.CRDClient, crds)
}

func (op *Operator) InitOperator() error {
	var err error

	if op.KubeClient, err = kubernetes.NewForConfig(op.ClientConfig); err != nil {
		return err
	}
	if op.ApiClient, err = cs.NewForConfig(op.ClientConfig); err != nil {
		return err
	}
	if op.CRDClient, err = kext_cs.NewForConfig(op.ClientConfig); err != nil {
		return err
	}

	op.ApiInformerFactory = api_informer.NewFilteredSharedInformerFactory(
		op.ApiClient,
		op.ResyncPeriod,
		op.WatchNamespace,
		nil,
	)
	op.recorder = eventer.NewEventRecorder(op.KubeClient, "kubeci-operator")

	if err := op.ensureCustomResourceDefinitions(); err != nil {
		return err
	}

	op.initWorkflowWatcher()
	op.initWorkplanWatcher()
	op.initDynamicWatcher()

	return nil
}

func (op *Operator) RunInformers(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	op.ApiInformerFactory.Start(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	for t, v := range op.ApiInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for caches to sync", t)
			return
		}
	}

	op.wfQueue.Run(stopCh)
	op.wpQueue.Run(stopCh)
}
