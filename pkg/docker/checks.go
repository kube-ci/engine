package docker

const (
	ACRegistry  = "kubeci"
	ImageKubeci = "kubeci-engine"
)

type Docker struct {
	Registry, Image, Tag string
}

func (docker Docker) ToContainerImage() string {
	return docker.Registry + "/" + docker.Image + ":" + docker.Tag
}
