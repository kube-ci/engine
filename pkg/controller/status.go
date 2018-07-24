package controller

import (
	"github.com/appscode/go/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
)

func (c *Controller) updateWorkflowLastObservedGen(name, namespace string, generation int64) error {
	wf, err := c.kubeciClient.KubeciV1alpha1().Workflows(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	wf.Status.LastObservedGeneration = &generation

	_, err = c.kubeciClient.KubeciV1alpha1().Workflows(namespace).UpdateStatus(wf)
	if err != nil {
		log.Errorf("failed to update status of workflow %s/%s, reason: %s", namespace, name, err.Error())
	}

	return err
}

func (c *Controller) updateWorkflowLastObservedResourceGen(name, namespace, uid string, generation *int64) error {
	wf, err := c.kubeciClient.KubeciV1alpha1().Workflows(namespace).Get(name, metav1.GetOptions{})
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

	_, err = c.kubeciClient.KubeciV1alpha1().Workflows(wf.Namespace).UpdateStatus(wf)
	if err != nil {
		log.Errorf("failed to update status of workflow %s/%s, reason: %s", namespace, name, err.Error())
	}

	return err
}

func (c *Controller) updateWorkplanStatus(name, namespace string, status api.WorkplanStatus) error {
	wp, err := c.kubeciClient.KubeciV1alpha1().Workplans(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Errorf("failed to update status of workplan %s/%s, reason: %s", namespace, name, err.Error())
		return err
	}

	wp.Status = status

	_, err = c.kubeciClient.KubeciV1alpha1().Workplans(namespace).UpdateStatus(wp)
	if err != nil {
		log.Errorf("failed to update status of workplan %s/%s, reason: %s", namespace, name, err.Error())
	}

	return err
}
