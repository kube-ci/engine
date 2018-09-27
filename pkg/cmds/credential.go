package cmds

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/spf13/cobra"
	"kube.ci/kubeci/pkg/credentials"
	"kube.ci/kubeci/pkg/credentials/dockercreds"
	"kube.ci/kubeci/pkg/credentials/gitcreds"
)

func NewCmdCredential() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "credential",
		Short:             "Run credential initializer",
		Long:              "Run credential initializer",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			builders := []credentials.Builder{dockercreds.NewBuilder(), gitcreds.NewBuilder()}
			for _, c := range builders {
				if err := c.Write(); err != nil {
					return fmt.Errorf("error initializing credentials: %v", err)
				}
			}
			log.Infof("Credentials initialized.")
			return nil
		},
	}

	return cmd
}
