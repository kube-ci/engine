package cmds

import (
	"path/filepath"

	"github.com/appscode/go/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"kube.ci/engine/pkg/logs"
)

func NewCmdWorkplanLogs() *cobra.Command {
	var (
		query      logs.Query
		kubeConfig string
	)
	cmd := &cobra.Command{
		Use:               "workplan-logs",
		Short:             "Get workplan logs",
		Long:              "Get workplan logs",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if kubeConfig == "" {
				kubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
			}
			c, err := logs.NewLogController(kubeConfig)
			if err != nil {
				log.Errorf("error initializing log-controller, reason: %v", err)
			}
			if err = c.GetLogs(query); err != nil {
				log.Errorf("error collecting log, reason: %v", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&query.Namespace, "namespace", "", "Namespace of the workflow.") // TODO: set default ns?
	cmd.Flags().StringVar(&query.Workflow, "workflow", "", "Name of the workflow.")
	cmd.Flags().StringVar(&query.Workplan, "workplan", "", "Name of the workplan.")
	cmd.Flags().StringVar(&query.Step, "step", "", "Name of the step.")
	cmd.Flags().StringVar(&kubeConfig, "kubeconfig", "", "Kube config file.")
	return cmd
}

// engine workplan-logs --namespace default --workplan wf-pr-test-9dtw8 --step step-test
// using kubectl-plugin
// cp /home/dipta/go/bin/engine /home/dipta/go/bin/kubectl-kubeci
// kubectl kubeci workplan-logs
