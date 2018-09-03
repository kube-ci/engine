package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	ResourceKindTrigger = "Trigger"
	ResourceTriggers    = "triggers"
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
