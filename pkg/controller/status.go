package controller

import (
	"github.com/appscode/go/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
)

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
