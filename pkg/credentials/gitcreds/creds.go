package gitcreds

import (
	"flag"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"kube.ci/kubeci/pkg/credentials"
)

const (
	annotationPrefix = "credential.kube.ci/git-"
)

var (
	basicConfig basicGitConfig
	sshConfig   sshGitConfig
)

func flags(fs *flag.FlagSet) {
	basicConfig = basicGitConfig{entries: make(map[string]basicEntry)}
	fs.Var(&basicConfig, "basic-git", "List of secret=url pairs.")

	sshConfig = sshGitConfig{entries: make(map[string]sshEntry)}
	fs.Var(&sshConfig, "ssh-git", "List of secret=url pairs.")
}

func init() {
	flags(flag.CommandLine)
}

type gitConfigBuilder struct{}

// NewBuilder returns a new builder for Git credentials.
func NewBuilder() credentials.Builder { return &gitConfigBuilder{} }

// MatchingAnnotations extracts flags for the credential helper
// from the supplied secret and returns a slice (of length 0 or
// greater) of applicable domains.
func (*gitConfigBuilder) MatchingAnnotations(secret *corev1.Secret) []string {
	var flagName string
	var flags []string
	switch secret.Type {
	case corev1.SecretTypeBasicAuth:
		flagName = "basic-git"

	case corev1.SecretTypeSSHAuth:
		flagName = "ssh-git"

	default:
		return flags
	}

	for k, v := range secret.Annotations {
		if strings.HasPrefix(k, annotationPrefix) {
			flags = append(flags, fmt.Sprintf("--%s=%s=%s", flagName, secret.Name, v))
		}
	}
	return flags
}

func (*gitConfigBuilder) Write() error {
	if err := basicConfig.Write(); err != nil {
		return err
	}
	return sshConfig.Write()
}
