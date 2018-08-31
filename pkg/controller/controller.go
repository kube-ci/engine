package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	dynamicclientset "github.com/appscode/kutil/dynamic/clientset"
	dynamicinformer "github.com/appscode/kutil/dynamic/informer"
	"github.com/appscode/kutil/tools/queue"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
	cs "kube.ci/kubeci/client/clientset/versioned"
	api_informers "kube.ci/kubeci/client/informers/externalversions"
	api_listers "kube.ci/kubeci/client/listers/kubeci/v1alpha1"
)

type Controller struct {
	config

	clientConfig *rest.Config
	kubeClient   kubernetes.Interface
	kubeciClient cs.Interface
	crdClient    crd_cs.ApiextensionsV1beta1Interface
	recorder     record.EventRecorder

	kubeInformerFactory   informers.SharedInformerFactory
	kubeciInformerFactory api_informers.SharedInformerFactory

	// Workflow
	wfQueue    *queue.Worker
	wfInformer cache.SharedIndexInformer
	wfLister   api_listers.WorkflowLister

	// Workplan
	wpQueue    *queue.Worker
	wpInformer cache.SharedIndexInformer
	wpLister   api_listers.WorkplanLister

	dynClient           *dynamicclientset.Clientset
	dynInformersFactory *dynamicinformer.SharedInformerFactory

	// store observed resources for workflows
	// store triggered-for in workplans
	// initially sync from available workplans
	// key: workflow namespace/name
	observedResources map[string]map[api.ObjectReference]api.ResourceGeneration

	// TODO: close unused informers
	// only one informer is created for a specific resource (among all workflows)
	// we should close a informer when no workflow need that informer (when workflows deleted or updated)
}

func (c *Controller) ensureCustomResourceDefinitions() error {
	crds := []*crd_api.CustomResourceDefinition{
		api.Workflow{}.CustomResourceDefinition(),
		api.Workplan{}.CustomResourceDefinition(),
	}
	return crdutils.RegisterCRDs(c.crdClient, crds)
}

func (c *Controller) RunInformers(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()

	log.Info("Starting kubeci controller")
	c.kubeInformerFactory.Start(stopCh)
	c.kubeciInformerFactory.Start(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	for _, v := range c.kubeInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
			return
		}
	}
	for _, v := range c.kubeciInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
			return
		}
	}

	// sync workplans into observedResources // TODO
	c.observedResources = make(map[string]map[api.ObjectReference]api.ResourceGeneration)

	c.wfQueue.Run(stopCh)
	c.wpQueue.Run(stopCh)
}
