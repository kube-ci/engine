package controller

import (
	"fmt"
	"path/filepath"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
)

func (c *Controller) runTasks(wp *api.Workplan) error {
	for index, task := range wp.Spec.Tasks {
		log.Infof("Running task[%d] for workplan %s", index, wp.Name)
		c.updateWorkplanStatus(wp.Name, wp.Namespace, api.WorkplanStatus{
			Phase:     "Running",
			TaskIndex: index,
			Reason:    fmt.Sprintf("Running task[%d]", index),
		})

		pod := podSpecForTasks(wp, task, index)
		pod, err := c.kubeClient.CoreV1().Pods(pod.Namespace).Create(pod)
		if err != nil {
			return fmt.Errorf("failed to create pod %s for task[%d], reason: %s", pod.Name, index, err.Error())
		}

		// wait until pod completes
		for {
			pod, err = c.kubeClient.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
			if err != nil {
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
	hostPathType := core.HostPathDirectoryOrCreate
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", wp.Name, index),
			Namespace: wp.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         api.SchemeGroupVersion.Group + "/" + api.SchemeGroupVersion.Version,
					Kind:               api.ResourceKindWorkplan,
					Name:               wp.Name,
					UID:                wp.UID,
					BlockOwnerDeletion: types.TrueP(),
				},
			},
		},
		Spec: core.PodSpec{
			RestartPolicy: core.RestartPolicyNever,
			Volumes: []core.Volume{
				{
					Name: "kubeci-shared-volume",
					VolumeSource: core.VolumeSource{
						HostPath: &core.HostPathVolumeSource{
							Path: filepath.Join("/kubeci", wp.Name),
							Type: &hostPathType,
						},
					},
				},
			},
		},
	}

	kubeciVolumeMount := core.VolumeMount{
		Name:      "kubeci-shared-volume",
		MountPath: "/kubeci",
	}

	for _, step := range task.SerialSteps {
		container := core.Container{
			Name:         step.Name,
			Command:      step.Commands,
			Args:         step.Args,
			Image:        step.Image,
			VolumeMounts: []core.VolumeMount{kubeciVolumeMount},
			EnvFrom:      wp.Spec.EnvFrom,
		}
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, container)
	}

	for _, step := range task.ParallelSteps {
		container := core.Container{
			Name:         step.Name,
			Command:      step.Commands,
			Args:         step.Args,
			Image:        step.Image,
			VolumeMounts: []core.VolumeMount{kubeciVolumeMount},
			EnvFrom:      wp.Spec.EnvFrom,
		}
		pod.Spec.Containers = append(pod.Spec.Containers, container)
	}

	return pod
}
