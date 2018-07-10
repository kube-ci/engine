package operator

import (
	"github.com/appscode/go/log"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kube-ci/experiments/apis/kubeci/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

func (op *Operator) initWorkplanWatcher() {
	op.wpInformer = op.ApiInformerFactory.Kubeci().V1alpha1().Workplans().Informer()
	op.wpQueue = queue.New("Workplan", op.MaxNumRequeues, op.NumThreads, op.reconcileWorkplan)
	op.wpInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			wp := obj.(*api.Workplan).DeepCopy()
			log.Debugln("Added workplan", wp.Name)
			queue.Enqueue(op.wpQueue.GetQueue(), obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			wp := newObj.(*api.Workplan).DeepCopy()
			log.Debugln("Updated workplan", wp.Name)
			queue.Enqueue(op.wpQueue.GetQueue(), newObj)
		},
		DeleteFunc: func(obj interface{}) {
			wp := obj.(*api.Workplan).DeepCopy()
			log.Debugln("Deleted workplan", wp.Name)
			queue.Enqueue(op.wpQueue.GetQueue(), obj)
		},
	})
	op.wpLister = op.ApiInformerFactory.Kubeci().V1alpha1().Workplans().Lister()
}

func (op *Operator) reconcileWorkplan(key string) error {
	obj, exists, err := op.wpInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}
	if !exists {
		log.Errorf("Workplan %s does not exist anymore", key)
		return nil
	}

	wp := obj.(*api.Workplan).DeepCopy()
	if wp.Status.Phase == "" {
		// not processed yet, process a workplan only once, ignore any updates
		log.Infof("Sync/Add/Update for workplan %s", key)
		go op.executeWorkplan(wp)
	}

	return nil
}

func (op *Operator) executeWorkplan(wp *api.Workplan) {
	log.Infof("Executing workplan %s", wp.Name)
	op.updateWorkplanStatus(wp.Name, wp.Namespace, api.WorkplanStatus{
		Phase:     "Pending",
		TaskIndex: -1,
		Reason:    "Initializing tasks",
	})

	if err := op.runTasks(wp); err != nil {
		log.Errorf("Failed to execute workplan: %s, reason: %s", wp.Name, err.Error())
		op.updateWorkplanStatus(wp.Name, wp.Namespace, api.WorkplanStatus{
			Phase:     "Failed",
			TaskIndex: -1,
			Reason:    err.Error(),
		})
		return
	}

	log.Infof("Workplan %s completed successfully", wp.Name)
	op.updateWorkplanStatus(wp.Name, wp.Namespace, api.WorkplanStatus{
		Phase:     "Completed",
		TaskIndex: -1,
		Reason:    "All tasks completed successfully",
	})
}
