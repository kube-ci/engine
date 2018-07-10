package operator

import (
	"github.com/appscode/go/log"
	api "github.com/kube-ci/experiments/apis/kubeci/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (op *Operator) updateWorkflowLastObservedGen(name, namespace string, generation int64) error {
	wf, err := op.ApiClient.KubeciV1alpha1().Workflows(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	wf.Status.LastObservedGeneration = &generation

	_, err = op.ApiClient.KubeciV1alpha1().Workflows(namespace).UpdateStatus(wf)
	if err != nil {
		log.Errorf("failed to update status of workflow %s/%s, reason: %s", namespace, name, err.Error())
	}

	return err
}

func (op *Operator) updateWorkflowLastObservedResourceGen(name, namespace, uid string, generation *int64) error {
	wf, err := op.ApiClient.KubeciV1alpha1().Workflows(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if wf.Status.LastObservedResourceGeneration == nil {
		wf.Status.LastObservedResourceGeneration = make(map[string]int64, 0)
	}

	if generation == nil { // delete
		delete(wf.Status.LastObservedResourceGeneration, uid)
	} else { // add or update
		wf.Status.LastObservedResourceGeneration[uid] = *generation
	}

	_, err = op.ApiClient.KubeciV1alpha1().Workflows(wf.Namespace).UpdateStatus(wf)
	if err != nil {
		log.Errorf("failed to update status of workflow %s/%s, reason: %s", namespace, name, err.Error())
	}

	return err
}

func (op *Operator) updateWorkplanStatus(name, namespace string, status api.WorkplanStatus) error {
	wp, err := op.ApiClient.KubeciV1alpha1().Workplans(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Errorf("failed to update status of workplan %s/%s, reason: %s", namespace, name, err.Error())
		return err
	}

	wp.Status = status

	_, err = op.ApiClient.KubeciV1alpha1().Workplans(namespace).UpdateStatus(wp)
	if err != nil {
		log.Errorf("failed to update status of workplan %s/%s, reason: %s", namespace, name, err.Error())
	}

	return err
}
