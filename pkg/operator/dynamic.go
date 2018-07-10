package operator

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	dynamicclientset "github.com/appscode/kutil/dynamic/clientset"
	dynamicdiscovery "github.com/appscode/kutil/dynamic/discovery"
	dynamicinformer "github.com/appscode/kutil/dynamic/informer"
	api "github.com/kube-ci/experiments/apis/kubeci/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/cache"
)

type ResourceIdentifier struct {
	Name              string
	ApiVersion        string
	Kind              string
	Namespace         string
	UID               ktypes.UID
	Generation        int64
	DeletionTimestamp *metav1.Time
	Labels            map[string]string
}

func (res ResourceIdentifier) String() string {
	return fmt.Sprintf("%s/%s:%s/%s", res.ApiVersion, res.Kind, res.Namespace, res.Name)
}

func objToResourceIdentifier(obj interface{}) ResourceIdentifier {
	o := obj.(*unstructured.Unstructured)
	return ResourceIdentifier{
		ApiVersion:        o.GetAPIVersion(),
		Kind:              o.GetKind(),
		Namespace:         o.GetNamespace(),
		Name:              o.GetName(),
		UID:               o.GetUID(),
		Generation:        o.GetGeneration(),
		Labels:            o.GetLabels(),
		DeletionTimestamp: o.GetDeletionTimestamp(),
	}
}

func (op *Operator) initDynamicWatcher() {
	// resync periods
	discoveryInterval := 5 * time.Second
	informerRelist := 5 * time.Minute

	// Periodically refresh discovery to pick up newly-installed resources.
	dc := discovery.NewDiscoveryClientForConfigOrDie(op.ClientConfig)
	resources := dynamicdiscovery.NewResourceMap(dc)
	// We don't care about stopping this cleanly since it has no external effects.
	resources.Start(discoveryInterval)

	// Create dynamic clientset (factory for dynamic clients).
	op.dynClient = dynamicclientset.New(op.ClientConfig, resources)
	// Create dynamic informer factory (for sharing dynamic informers).
	op.dynInformersFactory = dynamicinformer.NewSharedInformerFactory(op.dynClient, informerRelist)

	log.Infof("Waiting for sync")
	if !resources.HasSynced() {
		time.Sleep(time.Second)
	}
}

func (op *Operator) createInformer(wf *api.Workflow) error {
	for _, trigger := range wf.Spec.Triggers {
		informer, err := op.dynInformersFactory.Resource(trigger.ApiVersion, trigger.Resource)
		if err != nil {
			return err
		}
		informer.Informer().AddEventHandler(op.handlerForDynamicInformer())
		log.Infof("Created informer for resource %s/%s", trigger.ApiVersion, trigger.Resource)
	}
	return nil
}

// common handler for all events
func (op *Operator) handlerForDynamicInformer() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			res := objToResourceIdentifier(obj)
			log.Debugln("Added resource", res)
			op.handleTrigger(res)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			res := objToResourceIdentifier(newObj)
			log.Debugln("Updated resource", res)
			op.handleTrigger(res)
		},
		DeleteFunc: func(obj interface{}) {
			res := objToResourceIdentifier(obj)
			log.Debugln("Deleted resource", res)
			op.handleTrigger(res)
		},
	}
}
