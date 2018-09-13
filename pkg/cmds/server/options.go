package server

import (
	"flag"
	"time"

	stringz "github.com/appscode/go/strings"
	v "github.com/appscode/go/version"
	"github.com/spf13/pflag"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
	cs "kube.ci/kubeci/client/clientset/versioned"
	"kube.ci/kubeci/pkg/controller"
	"kube.ci/kubeci/pkg/docker"
)

type ExtraOptions struct {
	EnableRBAC        bool
	KubeciImageTag    string
	DockerRegistry    string
	MaxNumRequeues    int
	NumThreads        int
	ScratchDir        string
	QPS               float64
	Burst             int
	ResyncPeriod      time.Duration
	DiscoveryInterval time.Duration
}

func NewExtraOptions() *ExtraOptions {
	api.EnableStatusSubresource = true
	return &ExtraOptions{
		DockerRegistry:    docker.ACRegistry,
		KubeciImageTag:    stringz.Val(v.Version.Version, "canary"),
		MaxNumRequeues:    5,
		NumThreads:        2,
		ScratchDir:        "/tmp",
		QPS:               100,
		Burst:             100,
		ResyncPeriod:      10 * time.Minute,
		DiscoveryInterval: 10 * time.Second,
	}
}

func (s *ExtraOptions) AddGoFlags(fs *flag.FlagSet) {
	fs.BoolVar(&s.EnableRBAC, "rbac", s.EnableRBAC, "Enable RBAC for operator")
	fs.StringVar(&s.ScratchDir, "scratch-dir", s.ScratchDir, "Directory used to store temporary files. Use an `emptyDir` in Kubernetes.")
	fs.StringVar(&s.KubeciImageTag, "image-tag", s.KubeciImageTag, "Image tag for sidecar, init-container, check-job and recovery-job")
	fs.StringVar(&s.DockerRegistry, "docker-registry", s.DockerRegistry, "Docker image registry for sidecar, init-container, check-job, recovery-job and kubectl-job")

	fs.Float64Var(&s.QPS, "qps", s.QPS, "The maximum QPS to the master from this client")
	fs.IntVar(&s.Burst, "burst", s.Burst, "The maximum burst for throttle")
	fs.DurationVar(&s.ResyncPeriod, "resync-period", s.ResyncPeriod, "If non-zero, will re-list this often. Otherwise, re-list will be delayed as long as possible (until the upstream source closes the watch or times out.")
	fs.DurationVar(&s.DiscoveryInterval, "discovery-interval", s.DiscoveryInterval, "If non-zero, will refresh discovery info this often.")

	fs.BoolVar(&api.EnableStatusSubresource, "enable-status-subresource", api.EnableStatusSubresource, "If true, uses sub resource for crds.")
}

func (s *ExtraOptions) AddFlags(fs *pflag.FlagSet) {
	pfs := flag.NewFlagSet("kubeci", flag.ExitOnError)
	s.AddGoFlags(pfs)
	fs.AddGoFlagSet(pfs)
}

func (s *ExtraOptions) ApplyTo(cfg *controller.Config) error {
	var err error

	cfg.EnableRBAC = s.EnableRBAC
	cfg.KubeciImageTag = s.KubeciImageTag
	cfg.DockerRegistry = s.DockerRegistry
	cfg.MaxNumRequeues = s.MaxNumRequeues
	cfg.NumThreads = s.NumThreads
	cfg.ResyncPeriod = s.ResyncPeriod
	cfg.DiscoveryInterval = s.DiscoveryInterval

	cfg.ClientConfig.QPS = float32(s.QPS)
	cfg.ClientConfig.Burst = s.Burst

	if cfg.KubeClient, err = kubernetes.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.KubeciClient, err = cs.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	if cfg.CRDClient, err = crd_cs.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	return nil
}
