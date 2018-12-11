package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	ResourceKindTrigger     = "Trigger"
	ResourceTriggers        = "triggers"
	ResourceKindWorkplanLog = "WorkplanLog"
	ResourceWorkplanLogs    = "workplanlogs"
)

// +genclient
// +genclient:skipVerbs=list,get,update,patch,delete,deleteCollection,watch
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Trigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Workflows []string                   `json:"workflows,omitempty"`
	Request   *unstructured.Unstructured `json:"request,omitempty"`
}

// +genclient
// +genclient:skipVerbs=get,list,create,update,patch,delete,deleteCollection,watch
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkplanLog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkplanLogOptions struct {
	metav1.TypeMeta
	Step   string `json:"step,omitempty"`
	Follow bool   `json:"follow,omitempty"`
}
