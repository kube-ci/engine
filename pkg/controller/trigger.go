package controller

import (
	"context"
	"fmt"

	. "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	authorizationapi "k8s.io/api/authorization/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
	"kube.ci/kubeci/apis/trigger/v1alpha1"
	"kube.ci/kubeci/pkg/dependency"
	"kube.ci/kubeci/pkg/eventer"
)

type TriggerREST struct {
	controller *Controller
}

var _ rest.Creater = &TriggerREST{}
var _ rest.Scoper = &TriggerREST{}

func NewTriggerREST(controller *Controller) *TriggerREST {
	return &TriggerREST{
		controller: controller,
	}
}

func (r *TriggerREST) New() runtime.Object {
	return &v1alpha1.Trigger{}
}

func (r *TriggerREST) NamespaceScoped() bool {
	return true
}

func (r *TriggerREST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return v1alpha1.SchemeGroupVersion.WithKind(v1alpha1.ResourceKindTrigger)
}

func (r *TriggerREST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, includeUninitialized bool) (runtime.Object, error) {
	log.Info("Received force trigger event")
	trigger := obj.(*v1alpha1.Trigger)
	res := objToResourceIdentifier(trigger.Request)
	r.controller.handleTrigger(res, trigger.Workflows, false, true)
	return trigger, nil
}

func (c *Controller) handleTrigger(res ResourceIdentifier, wfNames []string, isDeleteEvent bool, force bool) {
	workflows, err := c.wfLister.Workflows(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		log.Errorf("Failed to list workflows, reason: %s", err.Error())
		return
	}

	// if list is not empty then only trigger for listed workflows
	var filteredWorkflows []*api.Workflow
	if len(wfNames) != 0 {
		for _, name := range wfNames {
			found := false
			for _, wf := range workflows {
				if wf.Name == name {
					filteredWorkflows = append(filteredWorkflows, wf)
					found = true
					break
				}
			}
			if !found {
				log.Errorf("Can't find workflow %s for resource %s", name, res)
			}
		}
	} else { // just copy
		for _, wf := range workflows {
			filteredWorkflows = append(filteredWorkflows, wf)
		}
	}

	for _, wf := range filteredWorkflows {
		if !force || wf.Spec.AllowForceTrigger {
			c.triggerWorkflow(wf, res, isDeleteEvent)
		}
	}
}

func (c *Controller) triggerWorkflow(wf *api.Workflow, res ResourceIdentifier, isDeleteEvent bool) {
	ok, trigger := c.shouldHandleTrigger(res, wf, isDeleteEvent)
	if !ok {
		log.Infof("should not handle trigger, resource: %s, workflow: %s/%s", res, wf.Namespace, wf.Name)
		return
	}
	data := res.GetData(trigger.EnvFromPath) // json-path-data
	triggeredFor := api.TriggeredFor{
		ObjectReference:    res.ObjectReference,
		ResourceGeneration: res.ResourceGeneration,
	}

	log.Infof("Triggering workflow %s for resource %s", wf.Name, triggeredFor)

	if err := c.performTrigger(wf, triggeredFor, data); err != nil {
		log.Errorf("Trigger failed for resource %v, reason: %s", res, err.Error())
		c.recorder.Eventf(
			wf.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonWorkflowTriggerFailed,
			"Trigger failed for resource %v, reason: %s", res, err.Error(),
		)
		return
	}

	// update generation and hash
	if c.observedResources[wf.Key()] == nil {
		c.observedResources[wf.Key()] = make(map[api.ObjectReference]*IntHash)
	}
	c.observedResources[wf.Key()][res.ObjectReference] = res.ResourceGeneration

	log.Infof("Successfully triggered workflow %s for resource %s", wf.Name, res)
	c.recorder.Eventf(
		wf.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonWorkflowTriggered,
		"Successfully triggered workflow %s for resource %s", wf.Name, res,
	)
}

// create json-path-secret and workplan
func (c *Controller) performTrigger(wf *api.Workflow, triggeredFor api.TriggeredFor, data map[string]string) error {
	var err error
	var secretRef *core.SecretEnvSource
	if data != nil && len(data) != 0 {
		if secretRef, err = c.createSecret(wf, data); err != nil {
			return err
		}
	}
	if _, err := c.createWorkplan(wf, secretRef, triggeredFor); err != nil {
		return err
	}
	return nil
}

func (c *Controller) shouldHandleTrigger(res ResourceIdentifier, wf *api.Workflow, isDeleteEvent bool) (bool, api.Trigger) {
	// check generation and hash to prevent duplicate trigger
	if c.resourceAlreadyObserved(wf.Key(), res) {
		return false, api.Trigger{}
	}

	// try to match all required conditions
	// if something not matched, continue matching with next element in trigger array
	// if everything matched, return true along with the trigger

	for _, trigger := range wf.Spec.Triggers {
		if trigger.APIVersion != res.ObjectReference.APIVersion {
			continue
		}
		if trigger.Kind != res.ObjectReference.Kind {
			continue
		}

		// match name and namespace if specified
		if trigger.Name != "" && trigger.Name != res.ObjectReference.Name {
			continue
		}
		// TODO: remove support for cross namespace trigger?
		if trigger.Namespace != "" && trigger.Namespace != res.ObjectReference.Namespace {
			continue
		}

		// match label-selector if specified
		if selector, err := metav1.LabelSelectorAsSelector(&trigger.Selector); err != nil ||
			!selector.Matches(labels.Set(res.Labels)) {
			continue
		}

		// match events
		if (isDeleteEvent && !trigger.OnDelete) || (!isDeleteEvent && !trigger.OnCreateOrUpdate) {
			continue
		}

		// === check RBAC permissions ===

		// check resource watch permission
		if ok := c.checkAccess(
			authorizationapi.ResourceAttributes{
				Group:     res.Group,
				Version:   res.Version,
				Resource:  res.Resource,
				Name:      res.ObjectReference.Name,
				Namespace: res.ObjectReference.Namespace,
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
					Name:      res.ObjectReference.Name,
					Namespace: res.ObjectReference.Namespace,
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

func (c *Controller) resourceAlreadyObserved(wfKey string, res ResourceIdentifier) bool {
	if c.observedResources == nil || c.observedResources[wfKey] == nil {
		return false
	}

	observed, ok := c.observedResources[wfKey][res.ObjectReference]
	return ok && observed.Equal(res.ResourceGeneration)
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

func (c *Controller) createWorkplan(wf *api.Workflow, secretRef *core.SecretEnvSource, triggeredFor api.TriggeredFor) (*api.Workplan, error) {
	cleanupStep := api.Step{
		Name:     "cleanup-step",
		Image:    "alpine",
		Commands: []string{"sh"},
		Args:     []string{"-c", "echo deleting files/folders; ls /kubeci; rm -rf /kubeci/*"},
	}

	tasks, err := dependency.ResolveDependency(wf.Spec.Steps, cleanupStep)
	if err != nil {
		return nil, err
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
			Workflow:     wf.Name,
			Tasks:        tasks,
			EnvFrom:      wf.Spec.EnvFrom,
			TriggeredFor: triggeredFor,
		},
	}

	if secretRef != nil { // secret with json-path data
		wp.Spec.EnvFrom = append(wp.Spec.EnvFrom, core.EnvFromSource{
			SecretRef: secretRef,
		})
	}

	log.Infof("Creating workplan for workflow %s", wf.Name)
	wp, err = c.kubeciClient.KubeciV1alpha1().Workplans(wp.Namespace).Create(wp)
	if err != nil {
		return nil, fmt.Errorf("failed to create workplan for workflow %s", wf.Name)
	}

	return wp, nil
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
