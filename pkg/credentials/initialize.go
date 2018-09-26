package credentials

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// VolumePath is the path where build secrets are written.
// It is mutable and exported for testing.
var VolumePath = "/var/build-secrets"

// Builder is the interface for a credential initializer of any type.
type Builder interface {
	// MatchingAnnotations extracts flags for the credential
	// helper from the supplied secret and returns a slice (of
	// length 0 or greater) of applicable domains.
	MatchingAnnotations(*corev1.Secret) []string

	// Write writes the credentials to the correct location.
	Write() error
}

// VolumeName returns the full path to the secret, inside the VolumePath.
func VolumeName(secretName string) string {
	return fmt.Sprintf("%s/%s", VolumePath, secretName)
}
