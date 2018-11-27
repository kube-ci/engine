package controller

import (
	"github.com/appscode/go/log"
	"github.com/appscode/kubernetes-webhook-util/admission"
	hooks "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	webhook "github.com/appscode/kubernetes-webhook-util/admission/v1beta1/generic"
	"github.com/appscode/kutil/tools/queue"
	"github.com/kube-ci/engine/apis/engine"
	api "github.com/kube-ci/engine/apis/engine/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (c *Controller) NewWorkflowValidatingWebhook() hooks.AdmissionHook {
	return webhook.NewGenericWebhook(
		schema.GroupVersionResource{
			Group:    "validators.engine.kube.ci",
			Version:  "v1alpha1",
			Resource: "workflows",
		},
		"workflow",
		[]string{kubeci.GroupName},
		api.SchemeGroupVersion.WithKind("Workflow"),
		nil,
		&admission.ResourceHandlerFuncs{
			CreateFunc: func(obj runtime.Object) (runtime.Object, error) {
				return nil, obj.(*api.Workflow).IsValid()
			},
			UpdateFunc: func(oldObj, newObj runtime.Object) (runtime.Object, error) {
				return nil, newObj.(*api.Workflow).IsValid()
			},
		},
	)
}

func (c *Controller) NewWorkflowMutatingWebhook() hooks.AdmissionHook {
	return webhook.NewGenericWebhook(
		schema.GroupVersionResource{
			Group:    "mutators.engine.kube.ci",
			Version:  "v1alpha1",
			Resource: "workflows",
		},
		"workflow",
		[]string{kubeci.GroupName},
		api.SchemeGroupVersion.WithKind("Workflow"),
		nil,
		&admission.ResourceHandlerFuncs{
			CreateFunc: func(obj runtime.Object) (runtime.Object, error) {
				return obj.(*api.Workflow).SetDefaults()
			},
			UpdateFunc: func(oldObj, newObj runtime.Object) (runtime.Object, error) {
				return newObj.(*api.Workflow).SetDefaults()
			},
		},
	)
}

func (c *Controller) initWorkflowWatcher() {
	c.wfInformer = c.kubeciInformerFactory.Engine().V1alpha1().Workflows().Informer()
	c.wfQueue = queue.New("Workflow", c.MaxNumRequeues, c.NumThreads, c.runWorkflowInjector)
	c.wfInformer.AddEventHandler(queue.NewEventHandler(c.wfQueue.GetQueue(), func(oldObj, newObj interface{}) bool {
		return !c.observedWorkflows.alreadyObserved(newObj.(*api.Workflow))
	}))
	c.wfLister = c.kubeciInformerFactory.Engine().V1alpha1().Workflows().Lister()
}

// always reconcile for add events, it will create required dynamic-informers
// for update events, reconcile if observed-generation is changed
// workflow observed-generation stored in operator memory instead of status
func (c *Controller) runWorkflowInjector(key string) error {
	obj, exist, err := c.wfInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}
	if !exist {
		log.Warningf("Workflow %s does not exist anymore\n", key)
		if err := c.deleteForWorkflow(key); err != nil {
			return err
		}
	} else {
		log.Infof("Sync/Add/Update for Workflow %s\n", key)
		wf := obj.(*api.Workflow).DeepCopy()
		if err := c.reconcileForWorkflow(wf); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) reconcileForWorkflow(wf *api.Workflow) error {
	if err := c.reconcileInformers(wf.Key(), wf.Spec.Triggers); err != nil {
		return err
	}
	c.observedWorkflows.set(wf)
	return nil
}

func (c *Controller) deleteForWorkflow(key string) error {
	if err := c.reconcileInformers(key, nil); err != nil {
		return err
	}
	c.observedWorkflows.delete(key)
	return nil
}
