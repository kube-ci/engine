package operator

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	api "github.com/kube-ci/experiments/apis/kubeci/v1alpha1"
	"github.com/kube-ci/experiments/pkg/dependency"
	"github.com/kube-ci/experiments/pkg/eventer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (op *Operator) handleTrigger(res ResourceIdentifier) {
	workflows, err := op.wfLister.Workflows(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		panic(err)
	}

	for _, wf := range workflows {
		if op.shouldHandleTrigger(res, wf) {
			log.Infof("Triggering workflow %s for resource %s", wf.Name, res)

			if err = op.createWorkplan(wf); err != nil {
				log.Errorf("Trigger failed for resource %v, reason: %s", res, err.Error())
				op.recorder.Eventf(
					wf.ObjectReference(),
					core.EventTypeWarning,
					eventer.EventReasonWorkflowTriggerFailed,
					"Trigger failed for resource %v, reason: %s", res, err.Error(),
				)
				return
			}

			log.Infof("Successfully triggered workflow %s for resource %s", wf.Name, res)
			op.recorder.Eventf(
				wf.ObjectReference(),
				core.EventTypeNormal,
				eventer.EventReasonWorkflowTriggered,
				"Successfully triggered workflow %s for resource %s", wf.Name, res,
			)
		}
	}
}

func (op *Operator) shouldHandleTrigger(res ResourceIdentifier, wf *api.Workflow) bool {
	for _, trigger := range wf.Spec.Triggers {
		if trigger.ApiVersion != res.ApiVersion {
			continue
		}
		if trigger.Kind != res.Kind {
			continue
		}

		// match name and namespace if specified
		if trigger.Name != "" && trigger.Name != res.Name {
			continue
		}
		if trigger.Namespace != "" && trigger.Namespace != res.Namespace {
			continue
		}

		// match label-selector if specified
		if selector, err := metav1.LabelSelectorAsSelector(&trigger.Selector); err != nil ||
			!selector.Matches(labels.Set(res.Labels)) {
			continue
		}

		// check generation to prevent duplicate trigger
		// also update/delete generation even if events not matched

		var gen int64
		var ok bool // false if map is nil or key not found
		if wf.Status.LastObservedResourceGeneration != nil {
			gen, ok = wf.Status.LastObservedResourceGeneration[string(res.UID)]
		}

		if res.DeletionTimestamp != nil {
			if !ok {
				continue
			}
			op.updateWorkflowLastObservedResourceGen(wf.Name, wf.Namespace, string(res.UID), nil)
		} else {
			if ok && gen >= res.Generation {
				continue
			}
			op.updateWorkflowLastObservedResourceGen(wf.Name, wf.Namespace, string(res.UID), &res.Generation)
		}

		// match events
		if res.DeletionTimestamp != nil && !trigger.OnDelete {
			continue
		}
		if res.DeletionTimestamp == nil && !trigger.OnCreateOrUpdate {
			continue
		}

		return true
	}
	return false
}

func (op *Operator) createWorkplan(wf *api.Workflow) error {
	tasks, err := dependency.ResolveDependency(wf)
	if err != nil {
		return err
	}
	wp := &api.Workplan{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: wf.Name + "-",
			Namespace:    wf.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         api.SchemeGroupVersion.Group + "/" + api.SchemeGroupVersion.Version,
					Kind:               api.ResourceKindWorkflow,
					Name:               wf.Name,
					UID:                wf.UID,
					BlockOwnerDeletion: types.TrueP(),
				},
			},
		},
		Spec: api.WorkplanSpec{
			Tasks: tasks,
		},
	}
	log.Infof("Creating workplan workflow %s", wf.Name)
	if wp, err = op.ApiClient.KubeciV1alpha1().Workplans(wp.Namespace).Create(wp); err != nil {
		return fmt.Errorf("failed to create workplan for workflow %s", wf.Name)
	}
	return nil
}
