package logs

import (
	"io"

	"github.com/appscode/kutil/tools/clientcmd"
	api "github.com/kube-ci/engine/apis/engine/v1alpha1"
	"github.com/kube-ci/engine/apis/extensions/v1alpha1"
	cs "github.com/kube-ci/engine/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

type LogController struct {
	kubeClient   kubernetes.Interface
	kubeciClient cs.Interface
}

type Query struct {
	Namespace string `survey:"namespace"`
	Workflow  string `survey:"workflow"`
	Workplan  string `survey:"workplan"`
	Step      string `survey:"step"`
}

func NewLogController(kubeConfig string) (*LogController, error) {
	clientConfig, err := clientcmd.BuildConfigFromContext(kubeConfig, "")
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	kubeciClient, err := cs.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	return &LogController{
		kubeClient:   kubeClient,
		kubeciClient: kubeciClient,
	}, nil
}

func (c *LogController) getWorkplanStatus(query Query) (api.WorkplanStatus, error) {
	workplan, err := c.kubeciClient.EngineV1alpha1().Workplans(query.Namespace).Get(query.Workplan, metav1.GetOptions{})
	if err != nil {
		return api.WorkplanStatus{}, err
	}
	return workplan.Status, nil
}

func (c *LogController) getLogReader(query Query) (io.ReadCloser, error) {
	opts := &v1alpha1.WorkplanLogOptions{
		Step:   query.Step,
		Follow: true,
	}
	return c.kubeciClient.ExtensionsV1alpha1().RESTClient().Get().
		Resource("workplanlogs").
		Namespace(query.Namespace).
		Name(query.Workplan).
		VersionedParams(opts, scheme.ParameterCodec).
		Stream()
}
