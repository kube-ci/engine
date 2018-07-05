package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkflow     = "Workflow"
	ResourceSingularWorkflow = "workflow"
	ResourcePluralWorkflow   = "workflows"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

type WorkflowSpec struct {
	Triggers []Trigger `json:"triggers,omitempty"`
	Steps    []Step    `json:"steps,omitempty"`
}

type Trigger struct {
	Name             string               `json:"name,omitempty"`
	ApiVersion       string               `json:"apiVersion,omitempty"`
	Kind             string               `json:"kind,omitempty"`
	Resource         string               `json:"resource,omitempty"`
	Namespace        string               `json:"namespace,omitempty"`
	Selector         metav1.LabelSelector `json:"selector,omitempty"`
	OnDelete         bool                 `json:"onDelete,omitempty"`
	OnCreateOrUpdate bool                 `json:"onAddOrUpdate,omitempty"`
}

type Step struct {
	Name       string   `json:"name,omitempty"`
	Image      string   `json:"image,omitempty"`
	Commands   []string `json:"commands,omitempty"`
	Args       []string `json:"args,omitempty"`
	Dependency []string `json:"dependency,omitempty"`
}

type WorkflowStatus struct {
	LastObservedGeneration         *int64           `json:"lastObservedGeneration"`
	LastObservedResourceGeneration map[string]int64 `json:"lastObservedResourceGeneration"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Workflow `json:"items"`
}
