package controller

import (
	"context"
	"fmt"
	"io"

	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/kubernetes"
	api "kube.ci/engine/apis/engine/v1alpha1"
	"kube.ci/engine/apis/extension/v1alpha1"
	cs "kube.ci/engine/client/clientset/versioned"
)

type LogsREST struct {
	controller *Controller
}

var _ rest.GetterWithOptions = &LogsREST{}
var _ rest.Scoper = &LogsREST{}
var _ rest.GroupVersionKindProvider = &LogsREST{}
var _ rest.CategoriesProvider = &LogsREST{}

func NewLogsREST(controller *Controller) *LogsREST {
	return &LogsREST{
		controller: controller,
	}
}

func (r *LogsREST) New() runtime.Object {
	return &v1alpha1.Trigger{}
}

func (r *LogsREST) NamespaceScoped() bool {
	return true
}

func (r *LogsREST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return v1alpha1.SchemeGroupVersion.WithKind(v1alpha1.ResourceKindWorkplanLog)
}

func (r *LogsREST) Categories() []string {
	return []string{"kubeci-engine", "ci", "appscode", "all"}
}

// Get retrieves a runtime.Object that will stream the contents of the pod log
func (r *LogsREST) Get(ctx context.Context, name string, options runtime.Object) (runtime.Object, error) {
	ns, ok := request.NamespaceFrom(ctx)
	if !ok {
		return nil, apierrors.NewBadRequest("namespace not found")
	}
	wpLogOptions, ok := options.(*v1alpha1.WorkplanLogOptions)
	if !ok {
		return nil, apierrors.NewBadRequest("workplan log options not found")
	}
	if wpLogOptions.Step == "" {
		return nil, apierrors.NewBadRequest("step name not specified")
	}
	return &streamer{
		Name:         name,
		Namespace:    ns,
		Step:         wpLogOptions.Step,
		Follow:       wpLogOptions.Follow,
		KubeClient:   r.controller.kubeClient,
		KubeciClient: r.controller.kubeciClient,
	}, nil
}

func (r *LogsREST) NewGetOptions() (runtime.Object, bool, string) {
	return &v1alpha1.WorkplanLogOptions{}, false, ""
}

type streamer struct {
	KubeClient   kubernetes.Interface
	KubeciClient cs.Interface

	Name      string // workplan name
	Namespace string // workplan namespace
	Step      string
	Follow    bool
}

var _ rest.ResourceStreamer = &streamer{}

func (s *streamer) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}
func (s *streamer) DeepCopyObject() runtime.Object {
	panic("streamer does not implement DeepCopyObject")
}

func (s *streamer) InputStream(ctx context.Context, apiVersion, acceptHeader string) (stream io.ReadCloser, flush bool, mimeType string, err error) {
	reader, err := s.getLogReader()
	if err != nil {
		err = apierrors.NewBadRequest(err.Error())
	}
	return reader, s.Follow, "text/plain", err
}

func (s *streamer) getLogReader() (io.ReadCloser, error) {
	stepEntry, err := s.getStepEntry()
	if err != nil {
		return nil, err
	}
	if stepEntry.Status == api.ContainerUninitialized {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("log not not available for step %s, reason: uninitialized", s.Step))
	}
	req := s.KubeClient.CoreV1().Pods(s.Namespace).GetLogs(stepEntry.PodName, &core.PodLogOptions{
		Container: s.Step,
		Follow:    s.Follow,
	})
	return req.Stream()
}

func (s *streamer) getStepEntry() (api.StepEntry, error) {
	workplan, err := s.KubeciClient.EngineV1alpha1().Workplans(s.Namespace).Get(s.Name, metav1.GetOptions{})
	if err != nil {
		return api.StepEntry{}, err
	}
	for _, stepEntries := range workplan.Status.StepTree {
		for _, stepEntry := range stepEntries {
			if stepEntry.Name == s.Step {
				return stepEntry, nil
			}
		}
	}
	return api.StepEntry{}, fmt.Errorf("pod not found for step %s", s.Step)
}

// https://{master-ip}/apis/extension.kube.ci/v1alpha1/namespace/{namespace}/workplanlogs/{workplan-name}?step={step-name}&follow={true|false}
