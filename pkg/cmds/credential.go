package cmds

import (
	"flag"
	"fmt"

	"github.com/appscode/go/log"
	"github.com/spf13/cobra"
	"kube.ci/engine/pkg/credentials"
	"kube.ci/engine/pkg/credentials/dockercreds"
	"kube.ci/engine/pkg/credentials/gitcreds"
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

	// add credential initializer flags
	pfs := flag.NewFlagSet("credential", flag.ExitOnError)
	dockercreds.Flags(pfs)
	gitcreds.Flags(pfs)
	cmd.Flags().AddGoFlagSet(pfs)

	return cmd
}
