package v1alpha1

import (
	htypes "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindWorkplan = "Workplan"
	ResourceWorkplans    = "workplans"
)

// +genclient
// +k8s:openapi-gen=true
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
	Workflow     string       `json:"workflow,omitempty"`
	Tasks        []Task       `json:"tasks,omitempty"`
	TriggeredFor TriggeredFor `json:"triggeredFor"`
	// set explicit environment variables
	EnvVar []corev1.EnvVar `json:"envVar,omitempty"`
	// set container environment variables from configmaps and secrets
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
	Volumes []corev1.Volume        `json:"volumes,omitempty"`
	// pod security context
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// ServiceAccount with triggering-resource/configmaps/secrets watch/read permissions.
	// Also used to run all associated pods
	ServiceAccount string                      `json:"serviceAccount,omitempty"`
	NodeSelector   map[string]string           `json:"nodeSelector,omitempty"`
	SchedulerName  string                      `json:"schedulerName,omitempty"`
	Tolerations    []corev1.Toleration         `json:"tolerations,omitempty"`
	Resources      corev1.ResourceRequirements `json:"resources,omitempty"`
}

type WorkplanPhase string

const (
	WorkplanPending       WorkplanPhase = "Pending"
	WorkplanRunning       WorkplanPhase = "Running"
	WorkplanSucceeded     WorkplanPhase = "Succeeded"
	WorkplanFailed        WorkplanPhase = "Failed"
	WorkplanUninitialized WorkplanPhase = "Uninitialized"
)

type ContainerStatus string

const (
	ContainerRunning       ContainerStatus = "Running"
	ContainerWaiting       ContainerStatus = "Waiting"
	ContainerTerminated    ContainerStatus = "Terminated"
	ContainerUninitialized ContainerStatus = "Uninitialized" // pod not exists
)

// status of a step containing enough info to collect logs
type StepEntry struct {
	Name           string                `json:"name"` // container name
	PodName        string                `json:"podName"`
	Status         ContainerStatus       `json:"status"` // simplified container status
	ContainerState corev1.ContainerState `json:"containerState"`
}

type WorkplanStatus struct {
	Phase     WorkplanPhase `json:"phase"`
	Reason    string        `json:"reason"`
	TaskIndex int           `json:"taskIndex"`
	NodeName  string        `json:"nodeName"`
	StepTree  [][]StepEntry `json:"stepTree"`
}

type TriggeredFor struct {
	ObjectReference    ObjectReference `json:"objectReference,omitempty"`
	ResourceGeneration *htypes.IntHash `json:"resourceGeneration,omitempty"`
}

type ObjectReference struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WorkplanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Workplan `json:"items"`
}

func (wp Workplan) Reference() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         SchemeGroupVersion.Group + "/" + SchemeGroupVersion.Version,
		Kind:               ResourceKindWorkplan,
		Name:               wp.Name,
		UID:                wp.UID,
		BlockOwnerDeletion: types.TrueP(),
	}
}
