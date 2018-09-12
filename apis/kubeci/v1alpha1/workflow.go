package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkflow = "Workflow"
	ResourceWorkflows    = "workflows"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WorkflowSpec `json:"spec,omitempty"`
}

type ExecutionOrderType string

const (
	SerialExecution   ExecutionOrderType = "serial"
	ParallelExecution ExecutionOrderType = "parallel"
	DagExecution      ExecutionOrderType = "dag"
)

type WorkflowSpec struct {
	AllowForceTrigger bool               `json:"allowForceTrigger,omitempty"`
	Triggers          []Trigger          `json:"triggers,omitempty"`
	Steps             []Step             `json:"steps,omitempty"`
	ExecutionOrder    ExecutionOrderType `json:"executionOrder,omitempty"`
	// set container environment variables from configmaps and secrets
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
	// ServiceAccount with triggering-resource/configmaps/secrets watch/read permissions
	// TODO: also use this in pods ?
	ServiceAccount string `json:"serviceAccount,omitempty"`
}

type Trigger struct {
	Name       string `json:"name,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Resource   string `json:"resource,omitempty"`
	// TODO: trigger for resources with different namespaces? or remove it?
	Namespace        string               `json:"namespace,omitempty"`
	Selector         metav1.LabelSelector `json:"selector,omitempty"`
	OnDelete         bool                 `json:"onDelete,omitempty"`
	OnCreateOrUpdate bool                 `json:"onCreateOrUpdate,omitempty"`
	// environment-variable to json-path map, set them in containers
	EnvFromPath map[string]string `json:"envFromPath,omitempty"`
}

type Step struct {
	Name       string   `json:"name,omitempty"`
	Image      string   `json:"image,omitempty"`
	Commands   []string `json:"commands,omitempty"`
	Args       []string `json:"args,omitempty"`
	Dependency []string `json:"dependency,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Workflow `json:"items"`
}

func (wf *Workflow) Key() string {
	return wf.Namespace + "/" + wf.Name
}
