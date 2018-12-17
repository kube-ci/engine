package controller

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	discovery_util "github.com/appscode/kutil/discovery"
	dynamicclientset "github.com/appscode/kutil/dynamic/clientset"
	dynamicdiscovery "github.com/appscode/kutil/dynamic/discovery"
	dynamicinformer "github.com/appscode/kutil/dynamic/informer"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kube-ci/engine/apis/engine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/jsonpath"
)

type ResourceIdentifier struct {
	Object             map[string]interface{} // required for json path data
	ObjectReference    api.ObjectReference
	ResourceGeneration *types.IntHash

	// TODO: remove extra fields
	Group    string
	Resource string
	Version  string
	Labels   map[string]string
}

func (res ResourceIdentifier) String() string {
	return fmt.Sprintf("%s/%s:%s/%s",
		res.ObjectReference.APIVersion, res.ObjectReference.Kind,
		res.ObjectReference.Namespace, res.ObjectReference.Name)
}

func (res ResourceIdentifier) GetEnvFromPath(paths map[string]string) []corev1.EnvVar {
	if paths == nil {
		return nil
	}
	var envVars []corev1.EnvVar
	for env, path := range paths {
		envVars = append(envVars, corev1.EnvVar{
			Name:  env,
			Value: jsonPathData(path, res.Object),
		})
	}
	return envVars
}

func jsonPathData(path string, data interface{}) string {
	j := jsonpath.New("kubeci-engine")
	j.AllowMissingKeys(true) // TODO: true or false ? ignore errors ?

	if err := j.Parse(path); err != nil {
		return ""
	}

	buf := new(bytes.Buffer)
	if err := j.Execute(buf, data); err != nil {
		return ""
	}

	return buf.String()
}

func (c *Controller) objToResourceIdentifier(obj interface{}) (*ResourceIdentifier, error) {
	o, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("failed to convert object to unstructured")
	}
	if o == nil {
		return nil, nil
	}

	apiVersion := o.GetAPIVersion()
	kind := o.GetKind()
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}
	gvr, err := discovery_util.ResourceForGVK(c.kubeClient.Discovery(), gv.WithKind(kind))
	if err != nil {
		return nil, err
	}

	return &ResourceIdentifier{
		Object:   o.Object,
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: gvr.Resource,
		Labels:   o.GetLabels(),
		ObjectReference: api.ObjectReference{
			APIVersion: apiVersion,
			Kind:       kind,
			Namespace:  o.GetNamespace(),
			Name:       o.GetName(),
		},
		ResourceGeneration: types.NewIntHash(o.GetGeneration(), meta_util.ObjectHash(o)),
	}, nil
}

func (c *Controller) initDynamicWatcher() {
	// set discovery-interval and resync-period
	discoveryInterval := c.DiscoveryInterval
	informerRelist := c.ResyncPeriod

	// Periodically refresh discovery to pick up newly-installed resources.
	dc := discovery.NewDiscoveryClientForConfigOrDie(c.clientConfig)
	resources := dynamicdiscovery.NewResourceMap(dc)
	// We don't care about stopping this cleanly since it has no external effects.
	resources.Start(discoveryInterval)

	// Create dynamic clientset (factory for dynamic clients).
	c.dynClient = dynamicclientset.New(c.clientConfig, resources)
	// Create dynamic informer factory (for sharing dynamic informers).
	c.dynInformersFactory = dynamicinformer.NewSharedInformerFactory(c.dynClient, informerRelist)

	log.Infof("Waiting for sync")
	if !resources.HasSynced() {
		time.Sleep(time.Second)
	}
}

type dynamicInformers struct {
	lock  sync.Mutex
	items map[string]informerStore
}

type informerStore struct {
	informer          *dynamicinformer.ResourceInformer
	workflows         sets.String
	DeletionTimestamp time.Time
}

// periodically check for unused informers after resync period
func (c *Controller) runInformerGC(stopCh <-chan struct{}) {
	log.Infof("Starting informer GC")
	go wait.JitterUntil(c.closeInformers, c.ResyncPeriod, 0.0, true, stopCh)
}

// close informers which are unused for more than resync period
func (c *Controller) closeInformers() {
	c.dynamicInformers.lock.Lock()
	defer c.dynamicInformers.lock.Unlock()

	for resourceKey, store := range c.dynamicInformers.items {
		if !store.DeletionTimestamp.IsZero() && time.Since(store.DeletionTimestamp) > c.ResyncPeriod {
			log.Infof("Closing unused informer for resource %s", resourceKey)
			store.informer.Close()
			delete(c.dynamicInformers.items, resourceKey)
		}
	}
}

func (c *Controller) reconcileInformers(workflowKey string, triggers []api.Trigger) error {
	// we need lock for the whole process
	c.dynamicInformers.lock.Lock()
	defer c.dynamicInformers.lock.Unlock()

	requiredInformers := sets.NewString()

	for _, trigger := range triggers {
		requiredInformers.Insert(trigger.ResourceKey())

		store, ok := c.dynamicInformers.items[trigger.ResourceKey()]
		if !ok { // informer not created
			var err error
			if store.informer, err = c.dynInformersFactory.Resource(trigger.APIVersion, trigger.Resource); err != nil {
				return err
			}
			store.informer.Informer().AddEventHandler(c.handlerForDynamicInformer())
		}
		// store workflow
		if store.workflows == nil {
			store.workflows = sets.NewString(workflowKey)
		} else {
			store.workflows.Insert(workflowKey)
		}
		store.DeletionTimestamp = time.Time{}                   // set DeletionTimestamp to zero
		c.dynamicInformers.items[trigger.ResourceKey()] = store // save in map

		log.Infof("Created informer for resource %s", trigger.ResourceKey())
	}

	createdInformers := sets.StringKeySet(c.dynamicInformers.items)
	diff := createdInformers.Difference(requiredInformers)
	for resourceKey := range diff {
		store := c.dynamicInformers.items[resourceKey]
		if store.workflows.Has(workflowKey) {
			store.workflows.Delete(workflowKey) // delete previously inserted workflow
			if store.workflows.Len() == 0 {     // unused informer, set deletion timestamp, don't delete immediately
				store.DeletionTimestamp = time.Now()
			}
		}
		c.dynamicInformers.items[resourceKey] = store // save in map
	}

	return nil
}

// common handler for all events
func (c *Controller) handlerForDynamicInformer() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Debugln("Updated resource", obj)
			if err := c.handleTrigger(obj, []string{"*"}, false, false); err != nil {
				log.Errorf(err.Error())
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Debugln("Updated resource", newObj)
			if err := c.handleTrigger(newObj, []string{"*"}, false, false); err != nil {
				log.Errorf(err.Error())
			}
		},
		DeleteFunc: func(obj interface{}) {
			log.Debugln("Deleted resource", obj)
			if err := c.handleTrigger(obj, []string{"*"}, true, false); err != nil {
				log.Errorf(err.Error())
			}
		},
	}
}
