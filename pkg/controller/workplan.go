package controller

import (
	"github.com/appscode/go/log"
	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
)

func (c *Controller) initWorkplanWatcher() {
	c.wpInformer = c.kubeciInformerFactory.Kubeci().V1alpha1().Workplans().Informer()
	c.wpQueue = queue.New("Workplan", c.MaxNumRequeues, c.NumThreads, c.runWorkplanInjector)
	c.wpInformer.AddEventHandler(queue.DefaultEventHandler(c.wpQueue.GetQueue()))
	c.wpLister = c.kubeciInformerFactory.Kubeci().V1alpha1().Workplans().Lister()
}

func (c *Controller) runWorkplanInjector(key string) error {
	obj, exist, err := c.wpInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("Workplan %s does not exist anymore\n", key)
	} else {
		wp := obj.(*api.Workplan).DeepCopy()
		if wp.Status.Phase == "" {
			// not processed yet, process a workplan only once, ignore any updates
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
	log.Infof("Executing workplan %s", wp.Name)
	c.updateWorkplanStatus(wp.Name, wp.Namespace, api.WorkplanStatus{
		Phase:     "Pending",
		TaskIndex: -1,
		Reason:    "Initializing tasks",
	})

	if err := c.runTasks(wp); err != nil {
		log.Errorf("Failed to execute workplan: %s, reason: %s", wp.Name, err.Error())
		c.updateWorkplanStatus(wp.Name, wp.Namespace, api.WorkplanStatus{
			Phase:     "Failed",
			TaskIndex: -1,
			Reason:    err.Error(),
		})
		return
	}

	log.Infof("Workplan %s completed successfully", wp.Name)
	c.updateWorkplanStatus(wp.Name, wp.Namespace, api.WorkplanStatus{
		Phase:     "Completed",
		TaskIndex: -1,
		Reason:    "All tasks completed successfully",
	})
}
