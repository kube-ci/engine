package operator

import (
	"github.com/appscode/go/log"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kube-ci/experiments/apis/kubeci/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

func (op *Operator) initWorkflowWatcher() {
	op.wfInformer = op.ApiInformerFactory.Kubeci().V1alpha1().Workflows().Informer()
	op.wfQueue = queue.New("Workflow", op.MaxNumRequeues, op.NumThreads, op.reconcileWorkflow)
	op.wfInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			wf := obj.(*api.Workflow).DeepCopy()
			log.Debugln("Added workflow", wf.Name)
			queue.Enqueue(op.wfQueue.GetQueue(), obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			wf := newObj.(*api.Workflow).DeepCopy()
			log.Debugln("Updated workflow", wf.Name)
			queue.Enqueue(op.wfQueue.GetQueue(), newObj)
		},
		DeleteFunc: func(obj interface{}) {
			wf := obj.(*api.Workflow).DeepCopy()
			log.Debugln("Deleted workflow", wf.Name)
			queue.Enqueue(op.wfQueue.GetQueue(), obj)
		},
	})
	op.wfLister = op.ApiInformerFactory.Kubeci().V1alpha1().Workflows().Lister()
}

func (op *Operator) reconcileWorkflow(key string) error {
	obj, exists, err := op.wfInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}
	if !exists {
		log.Errorf("Workflow %s does not exist anymore", key)
		return nil
	}

	wf := obj.(*api.Workflow).DeepCopy()
	if wf.Status.LastObservedGeneration == nil || wf.Generation > *wf.Status.LastObservedGeneration {
		log.Infof("Sync/Add/Update for workflow %s", key)
		op.updateWorkflowLastObservedGen(wf.Name, wf.Namespace, wf.Generation)
		if err := op.createInformer(wf); err != nil {
			return err
		}
	}

	return nil
}
