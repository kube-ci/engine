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
	"kube.ci/kubeci/client/clientset/versioned/typed/kubeci/v1alpha1/util"
	"kube.ci/kubeci/pkg/dependency"
	"kube.ci/kubeci/pkg/eventer"
)

func (c *Controller) handleTrigger(res ResourceIdentifier) {
	workflows, err := c.wfLister.Workflows(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		panic(err)
	}

	for _, wf := range workflows {
		if ok, trigger := c.shouldHandleTrigger(res, wf); ok {
			log.Infof("Triggering workflow %s for resource %s", wf.Name, res)

			// create secret with json-path data
			var secretRef *core.SecretEnvSource
			if data := res.GetData(trigger.EnvFromPath); data != nil {
				if secretRef, err = c.createSecret(wf, data); err != nil {
					log.Errorf("Trigger failed for resource %v, reason: %s", res, err.Error())
					c.recorder.Eventf(
						wf.ObjectReference(),
						core.EventTypeWarning,
						eventer.EventReasonWorkflowTriggerFailed,
						"Trigger failed for resource %v, reason: %s", res, err.Error(),
					)
					return
				}
			}

			if err = c.createWorkplan(wf, secretRef); err != nil {
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

			// update generation and hash
			resourceKey := res.Namespace + "/" + res.Name
			if _, err := util.UpdateWorkflowStatus(
				c.kubeciClient.KubeciV1alpha1(),
				wf.ObjectMeta,
				func(r *api.WorkflowStatus) *api.WorkflowStatus {
					if r.ObservedResources == nil {
						r.ObservedResources = make(map[string]api.ResourceGeneration)
					}
					r.ObservedResources[resourceKey] = api.ResourceGeneration{
						Generation: res.Generation,
						Hash:       res.Hash,
					}
					return r
				},
			); err != nil {
				log.Errorf("Failed to update status of workflow %s for resource %s, reason: %s", wf.Name, res, err.Error())
				return
			}
		}
	}
}

func (c *Controller) shouldHandleTrigger(res ResourceIdentifier, wf *api.Workflow) (bool, api.Trigger) {
	// check generation and hash to prevent duplicate trigger
	if resourceAlreadyObserved(wf.Status, res) {
		return false, api.Trigger{}
	}

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

		// match events
		if res.DeletionTimestamp != nil && !trigger.OnDelete { // TODO: how to handle delete?
			continue
		}
		if res.DeletionTimestamp == nil && !trigger.OnCreateOrUpdate {
			continue
		}

		// === check RBAC permissions ===

		// check resource watch permission
		if ok := c.checkAccess(
			authorizationapi.ResourceAttributes{
				Group:     res.Group,
				Version:   res.Version,
				Resource:  res.Resource,
				Name:      res.Name,
				Namespace: wf.Namespace,
				Verb:      "watch",
			},
			wf.Spec.ServiceAccount,
		); !ok {
			return false, api.Trigger{}
		}

		if trigger.EnvFromPath != nil {
			// check resource get permission
			if ok := c.checkAccess(
				authorizationapi.ResourceAttributes{
					Group:     res.Group,
					Version:   res.Version,
					Resource:  res.Resource,
					Name:      res.Name,
					Namespace: wf.Namespace,
					Verb:      "get",
				},
				wf.Spec.ServiceAccount,
			); !ok {
				return false, api.Trigger{}
			}
			// check secret create permission // TODO: check secret get permission also ?
			if ok := c.checkAccess(
				authorizationapi.ResourceAttributes{ // TODO: use constants
					Group:     "",
					Version:   "v1",
					Resource:  "secrets",
					Namespace: wf.Namespace,
					Verb:      "create",
				},
				wf.Spec.ServiceAccount,
			); !ok {
				return false, api.Trigger{}
			}
		}

		for _, env := range wf.Spec.EnvFrom {
			if env.ConfigMapRef != nil {
				// check configmap get permission
				if ok := c.checkAccess(
					authorizationapi.ResourceAttributes{ // TODO: use constants
						Group:     "",
						Version:   "v1",
						Resource:  "configmaps",
						Name:      env.ConfigMapRef.Name,
						Namespace: wf.Namespace,
						Verb:      "get",
					},
					wf.Spec.ServiceAccount,
				); !ok {
					return false, api.Trigger{}
				}
			}
			if env.SecretRef != nil {
				// check secret get permission
				if ok := c.checkAccess(
					authorizationapi.ResourceAttributes{ // TODO: use constants
						Group:     "",
						Version:   "v1",
						Resource:  "secrets",
						Name:      env.SecretRef.Name,
						Namespace: wf.Namespace,
						Verb:      "get",
					},
					wf.Spec.ServiceAccount,
				); !ok {
					return false, api.Trigger{}
				}
			}
		}

		return true, trigger
	}
	return false, api.Trigger{}
}

// TODO: how to handle generation for configmaps, secrets?
func resourceAlreadyObserved(wfStatus api.WorkflowStatus, res ResourceIdentifier) bool {
	if wfStatus.ObservedResources == nil {
		return false
	}

	resourceKey := res.Namespace + "/" + res.Name
	gen, ok := wfStatus.ObservedResources[resourceKey]
	if !ok || gen.Generation < res.Generation || gen.Hash != res.Hash {
		return false
	}

	return true
}

func (c *Controller) checkAccess(res authorizationapi.ResourceAttributes, serviceAccount string) bool {
	review := authorizationapi.SubjectAccessReview{ // TODO: use constants
		Spec: authorizationapi.SubjectAccessReviewSpec{
			ResourceAttributes: &res,
			User:               fmt.Sprintf("system:serviceaccount:%s:%s", res.Namespace, serviceAccount),
			Groups: []string{
				"system:serviceaccounts",
				fmt.Sprintf("system:serviceaccounts:%s", res.Namespace),
			},
		},
	}

	result, err := c.kubeClient.AuthorizationV1().SubjectAccessReviews().Create(&review)
	// oneliners.PrettyJson(result, "SubjectAccessReview Result")
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

func (c *Controller) createWorkplan(wf *api.Workflow, secretRef *core.SecretEnvSource) error {
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
			Tasks:   tasks,
			EnvFrom: wf.Spec.EnvFrom,
		},
	}

	if secretRef != nil { // secret with json-path data
		wp.Spec.EnvFrom = append(wp.Spec.EnvFrom, core.EnvFromSource{
			SecretRef: secretRef,
		})
	}

	log.Infof("Creating workplan for workflow %s", wf.Name)
	if wp, err = c.kubeciClient.KubeciV1alpha1().Workplans(wp.Namespace).Create(wp); err != nil {
		return fmt.Errorf("failed to create workplan for workflow %s", wf.Name)
	}
	return nil
}

func (c *Controller) createSecret(wf *api.Workflow, data map[string]string) (*core.SecretEnvSource, error) {
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: wf.Name + "-",
			Namespace:    wf.Namespace,
			OwnerReferences: []metav1.OwnerReference{ // TODO: use workplan as owner ?
				{
					APIVersion:         api.SchemeGroupVersion.Group + "/" + api.SchemeGroupVersion.Version,
					Kind:               api.ResourceKindWorkflow,
					Name:               wf.Name,
					UID:                wf.UID,
					BlockOwnerDeletion: types.TrueP(),
				},
			},
		},
		StringData: data,
	}

	log.Infof("Creating secret for workflow %s", wf.Name)
	secret, err := c.kubeClient.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret for workflow %s", wf.Name)
	}

	return &core.SecretEnvSource{
		LocalObjectReference: core.LocalObjectReference{
			Name: secret.Name,
		},
	}, nil
}
