package main

import (
	"flag"
	"path/filepath"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/log/golog"
	"github.com/appscode/kutil/tools/clientcmd"
	"github.com/kube-ci/experiments/apis/kubeci/v1alpha1"
	"github.com/kube-ci/experiments/pkg/operator"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/util/homedir"
)

func main() {
	golog.InitLogs()
	defer golog.FlushLogs()
	flag.Parse()

	kubeConfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	clientConfig, err := clientcmd.BuildConfigFromContext(kubeConfig, "")
	if err != nil {
		panic(err)
	}

	op := operator.Operator{
		ClientConfig:   clientConfig,
		ResyncPeriod:   10 * time.Minute,
		WatchNamespace: core.NamespaceAll,
		MaxNumRequeues: 5,
		NumThreads:     2,
	}

	if err := op.InitOperator(); err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})
	op.RunInformers(stopCh)
	time.Sleep(3 * time.Second) // time for sync
	log.Infof("Operator started")

	log.Infof("Creating workflow %s", v1alpha1.Wf01.Name)
	_, err = op.ApiClient.KubeciV1alpha1().Workflows("default").Create(v1alpha1.Wf01)
	if err != nil {
		panic(err)
	}

	<-stopCh
}
