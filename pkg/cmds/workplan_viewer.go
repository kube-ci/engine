package cmds

import (
	"github.com/appscode/kutil/tools/clientcmd"
	workplan_viewer "github.com/kube-ci/engine/pkg/workplan-viewer"
	"github.com/spf13/cobra"
)

func NewCmdWorkplanViewer() *cobra.Command {
	var kubeConfig string
	cmd := &cobra.Command{
		Use:               "workplan-viewer",
		Short:             "Start workplan-viewer server",
		Long:              "Start workplan-viewer server",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientConfig, err := clientcmd.BuildConfigFromContext(kubeConfig, "")
			if err != nil {
				return err
			}
			return workplan_viewer.Serve(clientConfig)
		},
	}
	cmd.Flags().StringVar(&kubeConfig, "kubeconfig", "", "Kube config file.")
	return cmd
}

// engine workplan-viewer --kubeconfig ~/.kube/config
// 127.0.0.1:9090/namespaces/default/workplans/wf-pr-test-9dtw8
// 127.0.0.1:9090/namespaces/default/workplans/wf-pr-test-9dtw8/steps/step-test
