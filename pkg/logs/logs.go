package logs

import (
	"io"

	api "github.com/kube-ci/engine/apis/engine/v1alpha1"
	"github.com/kube-ci/engine/apis/extensions/v1alpha1"
	cs "github.com/kube-ci/engine/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type LogController struct {
	KubeClient   kubernetes.Interface
	KubeciClient cs.Interface
}

type Query struct {
	Namespace string `survey:"namespace"`
	Workflow  string `survey:"workflow"`
	Workplan  string `survey:"workplan"`
	Step      string `survey:"step"`
}

func NewLogController(clientConfig *rest.Config) (*LogController, error) {
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	kubeciClient, err := cs.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	return &LogController{
		KubeClient:   kubeClient,
		KubeciClient: kubeciClient,
	}, nil
}

func (c *LogController) WorkplanStatus(query Query) (api.WorkplanStatus, error) {
	workplan, err := c.KubeciClient.EngineV1alpha1().Workplans(query.Namespace).Get(query.Workplan, metav1.GetOptions{})
	if err != nil {
		return api.WorkplanStatus{}, err
	}
	return workplan.Status, nil
}

func (c *LogController) LogReader(query Query) (io.ReadCloser, error) {
	opts := &v1alpha1.WorkplanLogOptions{
		Step:   query.Step,
		Follow: true,
	}
	return c.KubeciClient.ExtensionsV1alpha1().RESTClient().Get().
		Resource("workplanlogs").
		Namespace(query.Namespace).
		Name(query.Workplan).
		VersionedParams(opts, scheme.ParameterCodec).
		Stream()
}
