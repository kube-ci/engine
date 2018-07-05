package v1alpha1

import (
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/go-openapi/spec"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/kube-openapi/pkg/common"
)

func (wf Workflow) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Version:       SchemeGroupVersion.Version,
		Plural:        ResourcePluralWorkflow,
		Singular:      ResourceSingularWorkflow,
		Kind:          ResourceKindWorkflow,
		ShortNames:    []string{"wf"},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "kubeci"},
		},
		SpecDefinitionName: "github.com/kubeci/experiments/apis/kubeci/v1beta1.Workflow",
		// EnableValidation:        true,
		// GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: true,
	}, setNameSchema)
}

func (wp Workplan) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Version:       SchemeGroupVersion.Version,
		Plural:        ResourcePluralWorkplan,
		Singular:      ResourceSingularWorkplan,
		Kind:          ResourceKindWorkplan,
		ShortNames:    []string{"wp"},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "kubeci"},
		},
		SpecDefinitionName: "github.com/kubeci/experiments/apis/kubeci/v1beta1.Workplan",
		// EnableValidation:        true,
		// GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: true,
	}, setNameSchema)
}

func setNameSchema(openapiSpec map[string]common.OpenAPIDefinition) {
	// ref: https://github.com/kubedb/project/issues/166
	// https://github.com/kubernetes/apimachinery/blob/94ebb086c69b9fec4ddbfb6a1433d28ecca9292b/pkg/util/validation/validation.go#L153
	var maxLength int64 = 63
	openapiSpec["k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"].Schema.SchemaProps.Properties["name"] = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
			Type:        []string{"string"},
			Format:      "",
			Pattern:     `^[a-z]([-a-z0-9]*[a-z0-9])?$`,
			MaxLength:   &maxLength,
		},
	}
}