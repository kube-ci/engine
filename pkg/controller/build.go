package controller

import (
	"fmt"
	"path"

	api "github.com/kube-ci/engine/apis/engine/v1alpha1"
	"github.com/kube-ci/engine/pkg/credentials"
	"github.com/kube-ci/engine/pkg/credentials/dockercreds"
	"github.com/kube-ci/engine/pkg/credentials/gitcreds"
	"github.com/kube-ci/engine/pkg/docker"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/plugin/pkg/admission/serviceaccount"
)

const (
	implicitHomeDir    = "/kubeci/home"
	implicitWorkingDir = "/kubeci/workspace"
)

var (
	cleanupStep = api.Step{
		Name:     "cleanup-step",
		Image:    "alpine",
		Commands: []string{"sh"},
		Args:     []string{"-c", "echo deleting files/folders; ls /kubeci; rm -rf /kubeci/home/*; rm -rf /kubeci/workspace/*"},
	}

	implicitEnvVars = []core.EnvVar{{
		Name:  "HOME",
		Value: implicitHomeDir,
	}}

	implicitVolumeMounts = []core.VolumeMount{{
		Name:      "home",
		MountPath: implicitHomeDir,
	}, {
		Name:      "workspace",
		MountPath: implicitWorkingDir,
	}}
)

func getImplicitEnvVars(wpName string) []core.EnvVar {
	downwardEnvVars := []core.EnvVar{
		{
			Name: "NAMESPACE",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "WORKPLAN",
			Value: wpName,
		},
	}
	return append(implicitEnvVars, downwardEnvVars...)
}

func getImplicitVolumes(wpName string) []core.Volume {
	hostPathType := core.HostPathDirectoryOrCreate
	return []core.Volume{{
		Name: "home",
		VolumeSource: core.VolumeSource{
			HostPath: &core.HostPathVolumeSource{
				Path: path.Join("/var/run/kubeci", wpName, "home"),
				Type: &hostPathType,
			},
		},
	}, {
		Name: "workspace",
		VolumeSource: core.VolumeSource{
			HostPath: &core.HostPathVolumeSource{
				Path: path.Join("/var/run/kubeci", wpName, "workspace"),
				Type: &hostPathType,
			},
		},
	}}
}

func (c *Controller) credentialInitializer(wf *api.Workflow) (*api.Step, []core.Volume, error) {
	serviceAccountName := wf.Spec.ServiceAccount
	if serviceAccountName == "" {
		serviceAccountName = serviceaccount.DefaultServiceAccountName
	}
	sa, err := c.kubeClient.CoreV1().ServiceAccounts(wf.Namespace).Get(serviceAccountName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	var (
		args         []string
		volumes      []core.Volume
		volumeMounts []core.VolumeMount
	)

	builders := []credentials.Builder{dockercreds.NewBuilder(), gitcreds.NewBuilder()}

	for _, secretEntry := range sa.Secrets {
		secret, err := c.kubeClient.CoreV1().Secrets(wf.Namespace).Get(secretEntry.Name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}

		matched := false
		for _, b := range builders {
			if urlSecrets := b.MatchingAnnotations(secret); len(urlSecrets) > 0 {
				matched = true
				args = append(args, urlSecrets...)
			}
		}

		if matched {
			name := fmt.Sprintf("secret-volume-%s", secret.Name)
			volumeMounts = append(volumeMounts, core.VolumeMount{
				Name:      name,
				MountPath: credentials.VolumeName(secret.Name),
			})
			volumes = append(volumes, core.Volume{
				Name: name,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secret.Name,
					},
				},
			})
		}
	}

	if len(args) > 0 {
		step := &api.Step{
			Name: "credential-initializer",
			Image: docker.Docker{
				Image:    docker.ImageKubeci,
				Registry: c.DockerRegistry,
				Tag:      c.KubeciImageTag,
			}.ToContainerImage(),
			Args:         append([]string{"credential"}, args...),
			VolumeMounts: volumeMounts,
		}
		return step, volumes, nil
	}

	return nil, nil, nil
}
