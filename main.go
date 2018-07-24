package main

import (
	"os"
	"runtime"

	"github.com/appscode/go/log"
	logs "github.com/appscode/go/log/golog"
	_ "k8s.io/client-go/kubernetes/fake"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "kube.ci/kubeci/client/clientset/versioned/fake"
	"kube.ci/kubeci/pkg/cmds"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if err := cmds.NewRootCmd().Execute(); err != nil {
		log.Fatalln("Error in kubeci Main:", err)
	}
	log.Infoln("Exiting kubeci Main")
	os.Exit(0)
}
