package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kubernetes-webhook-util/admission"
	hooks "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	webhook "github.com/appscode/kubernetes-webhook-util/admission/v1beta1/generic"
	"github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/queue"
	kubeci "github.com/kube-ci/engine/apis/engine"
	api "github.com/kube-ci/engine/apis/engine/v1alpha1"
	"github.com/kube-ci/engine/client/clientset/versioned/typed/engine/v1alpha1/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (c *Controller) NewWorkplanMutatingWebhook() hooks.AdmissionHook {
	return webhook.NewGenericWebhook(
		schema.GroupVersionResource{
			Group:    "validators.engine.kube.ci",
			Version:  "v1alpha1",
			Resource: "workplans",
		},
		"workplan",
		[]string{kubeci.GroupName},
		api.SchemeGroupVersion.WithKind("Workplan"),
		nil,
		&admission.ResourceHandlerFuncs{
			// should not allow spec update
			UpdateFunc: func(oldObj, newObj runtime.Object) (runtime.Object, error) {
				oldWp := oldObj.(*api.Workplan)
				newWp := newObj.(*api.Workplan)
				if !meta.Equal(oldWp.Spec, newWp.Spec) {
					return nil, fmt.Errorf("workplan spec is immutable")
				}
				return nil, nil
			},
		},
	)
}

// process only add and delete events
// uninitialized: newly created
// running: previously created, but operator restarted before it succeeded

func (c *Controller) initWorkplanWatcher() {
	c.wpInformer = c.kubeciInformerFactory.Engine().V1alpha1().Workplans().Informer()
	c.wpQueue = queue.New("Workplan", c.MaxNumRequeues, c.NumThreads, c.runWorkplanInjector)
	c.wpInformer.AddEventHandler(queue.NewEventHandler(c.wpQueue.GetQueue(), func(oldObj, newObj interface{}) bool {
		wpOld := oldObj.(*api.Workplan).DeepCopy()
		wpNew := newObj.(*api.Workplan).DeepCopy()
		// handle update only for initial status update
		return wpOld.Status.Phase == "" && wpNew.Status.Phase == api.WorkplanUninitialized
	}))
	c.wpLister = c.kubeciInformerFactory.Engine().V1alpha1().Workplans().Lister()
}

func (c *Controller) runWorkplanInjector(key string) error {
	obj, exist, err := c.wpInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		log.Warningf("Workplan %s does not exist anymore\n", key)
	} else {
		wp := obj.(*api.Workplan).DeepCopy()
		if wp.Status.Phase == api.WorkplanUninitialized || wp.Status.Phase == api.WorkplanRunning {
			log.Infof("Sync/Add/Update for workplan %s", key)
			if err := c.reconcileForWorkplan(wp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Controller) reconcileForWorkplan(wp *api.Workplan) error {
	go c.executeWorkplan(wp)
	return nil
}

func (c *Controller) executeWorkplan(wp *api.Workplan) {
	var err error
	if wp.Status.Phase == api.WorkplanUninitialized {
		log.Infof("Executing workplan %s", wp.Name)
		if wp, err = util.UpdateWorkplanStatus(
			c.kubeciClient.EngineV1alpha1(),
			wp,
			func(r *api.WorkplanStatus) *api.WorkplanStatus {
				r.Phase = api.WorkplanPending
				r.TaskIndex = -1
				r.Reason = "Initializing tasks"
				return r
			},
			api.EnableStatusSubresource,
		); err != nil {
			log.Errorf(err.Error())
			return
		}
	} else if wp.Status.Phase == api.WorkplanRunning {
		log.Infof("Resuming workplan %s", wp.Name)
	}

	if err = c.runTasks(wp); err != nil {
		log.Errorf("Failed to execute workplan: %s, reason: %s", wp.Name, err.Error())
		// get latest before status update, since workplan status is changed inside runTasks()
		wp, e2 := c.kubeciClient.EngineV1alpha1().Workplans(wp.Namespace).Get(wp.Name, metav1.GetOptions{})
		if e2 != nil {
			log.Error(err)
			return
		}
		// set failed status
		wp, e3 := util.UpdateWorkplanStatus(
			c.kubeciClient.EngineV1alpha1(),
			wp,
			func(r *api.WorkplanStatus) *api.WorkplanStatus {
				r.Phase = api.WorkplanFailed
				r.TaskIndex = -1
				r.Reason = err.Error()
				return r
			},
			api.EnableStatusSubresource,
		)
		if e3 != nil {
			log.Errorf(err.Error())
		}
		return
	}

	log.Infof("Workplan %s completed successfully", wp.Name)
	// get latest before status update, since workplan status is changed inside runTasks()
	if wp, err = c.kubeciClient.EngineV1alpha1().Workplans(wp.Namespace).Get(wp.Name, metav1.GetOptions{}); err != nil {
		log.Error(err)
		return
	}
	// set succeeded status
	if wp, err = util.UpdateWorkplanStatus(
		c.kubeciClient.EngineV1alpha1(),
		wp,
		func(r *api.WorkplanStatus) *api.WorkplanStatus {
			r.Phase = api.WorkplanSucceeded
			r.TaskIndex = -1
			r.Reason = "All tasks completed successfully"
			return r
		},
		api.EnableStatusSubresource,
	); err != nil {
		log.Errorf(err.Error())
	}
}
