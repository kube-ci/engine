package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	authorizationapi "k8s.io/api/authorization/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
	"kube.ci/kubeci/pkg/dependency"
	"kube.ci/kubeci/pkg/eventer"
)

func (c *Controller) handleTrigger(res ResourceIdentifier) {
	workflows, err := c.wfLister.Workflows(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		panic(err)
	}

	for _, wf := range workflows {
		if c.shouldHandleTrigger(res, wf) {
			log.Infof("Triggering workflow %s for resource %s", wf.Name, res)

			if err = c.createWorkplan(wf); err != nil {
				log.Errorf("Trigger failed for resource %v, reason: %s", res, err.Error())
				c.recorder.Eventf(
					wf.ObjectReference(),
					core.EventTypeWarning,
					eventer.EventReasonWorkflowTriggerFailed,
					"Trigger failed for resource %v, reason: %s", res, err.Error(),
				)
				return
			}

			log.Infof("Successfully triggered workflow %s for resource %s", wf.Name, res)
			c.recorder.Eventf(
				wf.ObjectReference(),
				core.EventTypeNormal,
				eventer.EventReasonWorkflowTriggered,
				"Successfully triggered workflow %s for resource %s", wf.Name, res,
			)
		}
	}
}

func (c *Controller) shouldHandleTrigger(res ResourceIdentifier, wf *api.Workflow) bool {
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
			c.updateWorkflowLastObservedResourceGen(wf.Name, wf.Namespace, string(res.UID), nil)
		} else {
			if ok && gen >= res.Generation {
				continue
			}
			c.updateWorkflowLastObservedResourceGen(wf.Name, wf.Namespace, string(res.UID), &res.Generation)
		}

		// match events
		if res.DeletionTimestamp != nil && !trigger.OnDelete {
			continue
		}
		if res.DeletionTimestamp == nil && !trigger.OnCreateOrUpdate {
			continue
		}

		// === check RBAC permissions ===

		// check resource watch permission
		if ok := c.checkAccess(
			authorizationapi.ResourceAttributes{
				Group:     res.ApiVersion, // TODO: split into group/version
				Version:   res.ApiVersion,
				Resource:  res.Kind,
				Name:      res.Name,
				Namespace: wf.Namespace,
				Verb:      "watch",
			},
			wf.Spec.ServiceAccount,
		); !ok {
			return false
		}

		if trigger.EnvFromPath != nil {
			// check resource get permission
			if ok := c.checkAccess(
				authorizationapi.ResourceAttributes{
					Group:     res.ApiVersion, // TODO: split into group/version
					Version:   res.ApiVersion,
					Resource:  res.Kind,
					Name:      res.Name,
					Namespace: wf.Namespace,
					Verb:      "get",
				},
				wf.Spec.ServiceAccount,
			); !ok {
				return false
			}
			// check secret create permission // TODO: check secret get permission also ?
			if ok := c.checkAccess(
				authorizationapi.ResourceAttributes{ // TODO: check
					Group:     "core",
					Version:   "v1",
					Resource:  "Secret",
					Namespace: wf.Namespace,
					Verb:      "create",
				},
				wf.Spec.ServiceAccount,
			); !ok {
				return false
			}
		}

		for _, env := range wf.Spec.EnvFrom {
			if env.ConfigMapRef != nil {
				// check configmap get permission
				if ok := c.checkAccess(
					authorizationapi.ResourceAttributes{ // TODO: use constants
						Group:     "core",
						Version:   "v1",
						Resource:  "Configmap",
						Name:      env.ConfigMapRef.Name,
						Namespace: wf.Namespace,
						Verb:      "get",
					},
					wf.Spec.ServiceAccount,
				); !ok {
					return false
				}
			}
			if env.SecretRef != nil {
				// check secret get permission
				if ok := c.checkAccess(
					authorizationapi.ResourceAttributes{ // TODO: use constants
						Group:     "core",
						Version:   "v1",
						Resource:  "Secret",
						Name:      env.SecretRef.Name,
						Namespace: wf.Namespace,
						Verb:      "get",
					},
					wf.Spec.ServiceAccount,
				); !ok {
					return false
				}
			}
		}

		return true
	}
	return false
}

func (c *Controller) checkAccess(res authorizationapi.ResourceAttributes, serviceAccount string) bool {
	result, err := c.kubeClient.AuthorizationV1().SubjectAccessReviews().Create(
		&authorizationapi.SubjectAccessReview{
			Spec: authorizationapi.SubjectAccessReviewSpec{
				ResourceAttributes: &res,
				User:               "",         // TODO: fix
				Groups:             []string{}, // TODO: fix
			},
		},
	)
	if err != nil {
		log.Errorln(err)
		return false
	}
	if !result.Status.Allowed {
		log.Errorf("No permission, service-account: %s resource: %v", serviceAccount, res)
		return false
	}
	return true
}

func (c *Controller) createWorkplan(wf *api.Workflow) error {
	cleanupStep := api.Step{
		Name:     "cleanup-step",
		Image:    "alpine",
		Commands: []string{"sh"},
		Args:     []string{"-c", "echo deleting files/folders; ls /kubeci; rm -rf /kubeci/*"},
	}

	tasks, err := dependency.ResolveDependency(wf.Spec.Steps, cleanupStep)
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
	log.Infof("Creating workplan for workflow %s", wf.Name)
	if wp, err = c.kubeciClient.KubeciV1alpha1().Workplans(wp.Namespace).Create(wp); err != nil {
		return fmt.Errorf("failed to create workplan for workflow %s", wf.Name)
	}
	return nil
}
