package controller

import (
	"path"

	core "k8s.io/api/core/v1"
	api "kube.ci/kubeci/apis/kubeci/v1alpha1"
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

func getImplicitVolumes(wpName string) []core.Volume {
	hostPathType := core.HostPathDirectoryOrCreate
	return []core.Volume{{
		Name: "home",
		VolumeSource: core.VolumeSource{
			HostPath: &core.HostPathVolumeSource{
				Path: path.Join("/kubeci", wpName, "home"),
				Type: &hostPathType,
			},
		},
	}, {
		Name: "workspace",
		VolumeSource: core.VolumeSource{
			HostPath: &core.HostPathVolumeSource{
				Path: path.Join("/kubeci", wpName, "workspace"),
				Type: &hostPathType,
			},
		},
	}}
}
