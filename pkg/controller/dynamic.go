package controller

import (
	"bytes"
	"fmt"
	"time"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	discovery_util "github.com/appscode/kutil/discovery"
	dynamicclientset "github.com/appscode/kutil/dynamic/clientset"
	dynamicdiscovery "github.com/appscode/kutil/dynamic/discovery"
	dynamicinformer "github.com/appscode/kutil/dynamic/informer"
	meta_util "github.com/appscode/kutil/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/jsonpath"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
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

func (res ResourceIdentifier) GetData(paths map[string]string) map[string]string {
	if paths == nil {
		return nil
	}
	data := make(map[string]string, 0)
	for env, path := range paths {
		data[env] = jsonPathData(path, res.Object)
	}
	return data
}

func jsonPathData(path string, data interface{}) string {
	j := jsonpath.New("kubeci")
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

func (c *Controller) objToResourceIdentifier(obj interface{}) (ResourceIdentifier, error) {
	o := obj.(*unstructured.Unstructured)

	apiVersion := o.GetAPIVersion()
	kind := o.GetKind()
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return ResourceIdentifier{}, err
	}
	gvr, err := discovery_util.ResourceForGVK(c.kubeClient.Discovery(), gv.WithKind(kind))
	if err != nil {
		return ResourceIdentifier{}, err
	}

	return ResourceIdentifier{
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
	// resync periods
	discoveryInterval := c.ResyncPeriod
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

func (c *Controller) createInformer(wf *api.Workflow) error {
	for _, trigger := range wf.Spec.Triggers {
		informer, err := c.dynInformersFactory.Resource(trigger.APIVersion, trigger.Resource)
		if err != nil {
			return err
		}
		informer.Informer().AddEventHandler(c.handlerForDynamicInformer())
		log.Infof("Created informer for resource %s/%s", trigger.APIVersion, trigger.Resource)
	}
	return nil
}

// common handler for all events
func (c *Controller) handlerForDynamicInformer() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Debugln("Updated resource", obj)
			c.handleTrigger(obj, []string{}, false, false)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Debugln("Updated resource", newObj)
			c.handleTrigger(newObj, []string{}, false, false)
		},
		DeleteFunc: func(obj interface{}) {
			log.Debugln("Deleted resource", obj)
			c.handleTrigger(obj, []string{}, true, false)
		},
	}
}
