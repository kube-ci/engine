package controller

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"github.com/appscode/go/log"
	dynamicclientset "github.com/appscode/kutil/dynamic/clientset"
	dynamicdiscovery "github.com/appscode/kutil/dynamic/discovery"
	dynamicinformer "github.com/appscode/kutil/dynamic/informer"
	"github.com/appscode/kutil/meta"
	"github.com/fatih/structs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/jsonpath"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
)

type ResourceIdentifier struct {
	Object             map[string]interface{} // required for json path data
	ObjectReference    api.ObjectReference
	ResourceGeneration api.ResourceGeneration

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

func objToResourceIdentifier(obj interface{}) ResourceIdentifier {
	o := obj.(*unstructured.Unstructured)

	apiVersion := o.GetAPIVersion()
	group, version := toGroupAndVersion(apiVersion)

	return ResourceIdentifier{
		Object:   o.Object,
		Group:    group,
		Resource: selfLinkToResource(o.GetSelfLink()),
		Version:  version,
		Labels:   o.GetLabels(),
		ObjectReference: api.ObjectReference{
			APIVersion: apiVersion,
			Kind:       o.GetKind(),
			Namespace:  o.GetNamespace(),
			Name:       o.GetName(),
		},
		ResourceGeneration: api.ResourceGeneration{
			Generation: o.GetGeneration(),
			Hash:       objectHash(o),
		},
	}
}

// TODO: how to get resource from unstructured object ?
func selfLinkToResource(selfLink string) string {
	items := strings.Split(selfLink, "/")
	return items[len(items)-2]
}

// TODO: use this instead of selfLink ?
func (c *Controller) groupVersionKindToResource(groupVersion, kind string) (string, error) {
	resources, err := c.kubeClient.Discovery().ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return "", err
	}
	for _, resource := range resources.APIResources {
		if resource.Kind == kind {
			return resource.Name, nil
		}
	}
	return "", fmt.Errorf("could not find api resource with group-version: %s and kind: %s", groupVersion, kind)
}

func toGroupAndVersion(apiVersion string) (string, string) {
	items := strings.Split(apiVersion, "/")
	if len(items) == 1 {
		return "", items[0]
	}
	return items[0], items[1]
}

func (c *Controller) initDynamicWatcher() {
	// resync periods
	discoveryInterval := 5 * time.Second
	informerRelist := 5 * time.Minute

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
			res := objToResourceIdentifier(obj)
			log.Debugln("Added resource", res)
			c.handleTrigger(res, false)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			res := objToResourceIdentifier(newObj)
			log.Debugln("Updated resource", res)
			c.handleTrigger(res, false)
		},
		DeleteFunc: func(obj interface{}) {
			res := objToResourceIdentifier(obj)
			log.Debugln("Deleted resource", res)
			c.handleTrigger(res, true)
		},
	}
}

// TODO: move to kutil?
// hash includes all top label fields (like data, spec) except TypeMeta, ObjectMeta and Status
// also includes Generation, Annotation and Labels form ObjectMeta
func objectHash(in metav1.Object) string {
	obj := make(map[string]interface{})

	obj["generation"] = in.GetGeneration()
	if len(in.GetLabels()) > 0 {
		obj["labels"] = in.GetLabels()
	}

	if len(in.GetAnnotations()) > 0 {
		data := make(map[string]string, len(in.GetAnnotations()))
		lastAppliedConfiguration := "kubectl.kubernetes.io/last-applied-configuration"
		for k, v := range in.GetAnnotations() {
			if k != lastAppliedConfiguration {
				data[k] = v
			}
		}
		obj["annotations"] = data
	}

	st := structs.New(in)
	for _, field := range st.Fields() {
		fieldName := field.Name()
		if fieldName != "ObjectMeta" && fieldName != "TypeMeta" && fieldName != "Status" {
			obj[strings.ToLower(fieldName)] = st.Field(fieldName).Value()
		}
	}

	// oneliners.PrettyJson(obj)

	h := fnv.New64a()
	meta.DeepHashObject(h, obj)
	return strconv.FormatUint(h.Sum64(), 10)
}
