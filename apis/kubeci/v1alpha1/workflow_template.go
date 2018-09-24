package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkflowTemplate = "WorkflowTemplate"
	ResourceWorkflowTemplates    = "workflowtemplates"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkflowTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WorkflowTemplateSpec `json:"spec,omitempty"`
}

type WorkflowTemplateSpec struct {
	Steps []Step `json:"steps,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkflowTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []WorkflowTemplate `json:"items"`
}
