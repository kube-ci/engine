package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
	"kube.ci/kubeci/client/clientset/versioned/typed/kubeci/v1alpha1/util"
)

func (c *Controller) runTasks(wp *api.Workplan) error {
	for index, task := range wp.Spec.Tasks {
		var err error
		pod := podSpecForTasks(wp, task, index)

		if wp.Status.Phase == api.WorkplanPending || index > wp.Status.TaskIndex {
			log.Infof("Running task[%d] for workplan %s", index, wp.Name)
			if pod, err = c.kubeClient.CoreV1().Pods(pod.Namespace).Create(pod); err != nil {
				return fmt.Errorf("failed to create pod %s for task[%d], reason: %s", pod.Name, index, err.Error())
			}
			if wp, err = util.UpdateWorkplanStatus(
				c.kubeciClient.KubeciV1alpha1(),
				wp,
				func(r *api.WorkplanStatus) *api.WorkplanStatus {
					r.Phase = api.WorkplanRunning
					r.TaskIndex = index
					r.Reason = fmt.Sprintf("Running task[%d]", index)
					return r
				},
				api.EnableStatusSubresource,
			); err != nil {
				return err
			}
		}

		// wait until pod completes
		for {
			if pod, err = c.kubeClient.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{}); err != nil {
				return err
			}
			if pod.Status.Phase == core.PodSucceeded {
				log.Infof("Succeeded pod %s for task[%d] of workplan %s", pod.Name, index, wp.Name)
				break
			}
			if pod.Status.Phase == core.PodFailed {
				return fmt.Errorf("failed pod %s for task[%d]", pod.Name, index)
			}
		}
	}

	return nil
}

func podSpecForTasks(wp *api.Workplan, task api.Task, index int) *core.Pod {
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%d", wp.Name, index),
			Namespace:       wp.Namespace,
			OwnerReferences: []metav1.OwnerReference{wp.Reference()},
		},
		Spec: core.PodSpec{
			RestartPolicy: core.RestartPolicyNever,
			Volumes:       core_util.UpsertVolume(wp.Spec.Volumes, getImplicitVolumes(wp.Name)...),
		},
	}

	for _, step := range task.SerialSteps {
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, containerForStep(wp, step))
	}
	for _, step := range task.ParallelSteps {
		pod.Spec.Containers = append(pod.Spec.Containers, containerForStep(wp, step))
	}

	return pod
}

func containerForStep(wp *api.Workplan, step api.Step) core.Container {
	return core.Container{
		Name:         step.Name,
		Image:        step.Image,
		Command:      step.Commands,
		Args:         step.Args,
		EnvFrom:      wp.Spec.EnvFrom,
		Env:          core_util.UpsertEnvVars(wp.Spec.EnvVar, implicitEnvVars...),
		WorkingDir:   implicitWorkingDir,
		VolumeMounts: core_util.UpsertVolumeMount(step.VolumeMounts, implicitVolumeMounts...),
	}
}
