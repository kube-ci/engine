package v1alpha1

import (
	"strings"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdutils "kmodules.xyz/client-go/apiextensions/v1beta1"
)

var (
	EnableStatusSubresource bool
)

func (c Workflow) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourceWorkflows,
		Singular:      strings.ToLower(ResourceKindWorkflow),
		Kind:          ResourceKindWorkflow,
		ShortNames:    []string{"wf"},
		Categories:    []string{"kubeci", "ci", "appscode", "all"},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{
				Name:    SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
			},
		},
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "kubeci-engine"},
		},
		SpecDefinitionName:      "github.com/kube-ci/engine/apis/engine/v1alpha1.Workflow",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: EnableStatusSubresource,
	})
}

func (c Workplan) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourceWorkplans,
		Singular:      strings.ToLower(ResourceKindWorkplan),
		Kind:          ResourceKindWorkplan,
		ShortNames:    []string{"wp"},
		Categories:    []string{"kubeci", "ci", "appscode", "all"},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{
				Name:    SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
			},
		},
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "kubeci-engine"},
		},
		SpecDefinitionName:      "github.com/kube-ci/engine/apis/engine/v1alpha1.Workplan",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: EnableStatusSubresource,
	})
}

func (c WorkflowTemplate) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourceWorkflowTemplates,
		Singular:      strings.ToLower(ResourceKindWorkflowTemplate),
		Kind:          ResourceKindWorkflowTemplate,
		ShortNames:    []string{"wt"},
		Categories:    []string{"kubeci", "ci", "appscode", "all"},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{
				Name:    SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
			},
		},
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "kubeci-engine"},
		},
		SpecDefinitionName:      "github.com/kube-ci/engine/apis/engine/v1alpha1.WorkflowTemplate",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: EnableStatusSubresource,
	})
}
