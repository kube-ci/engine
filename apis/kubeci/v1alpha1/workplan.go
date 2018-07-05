package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkplan     = "Workplan"
	ResourceSingularWorkplan = "workplan"
	ResourcePluralWorkplan   = "workplans"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Workplan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkplanSpec   `json:"spec,omitempty"`
	Status WorkplanStatus `json:"status,omitempty"`
}

type Task struct { // analogous to a single pod
	SerialSteps   []Step // analogous to init-containers
	ParallelSteps []Step // analogous to sidecar-containers
}

type WorkplanSpec struct {
	Tasks []Task `json:"tasks,omitempty"`
}

type WorkplanStatus struct {
	Phase     string `json:"phase"`
	Reason    string `json:"reason"`
	TaskIndex int    `json:"taskIndex"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkplanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Workplan `json:"items"`
}
