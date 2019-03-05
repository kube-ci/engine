package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/appscode/go/log"
	gort "github.com/appscode/go/runtime"
	"github.com/go-openapi/spec"
	api_install "github.com/kube-ci/engine/apis/engine/install"
	v1alpha1 "github.com/kube-ci/engine/apis/engine/v1alpha1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
	crdutils "kmodules.xyz/client-go/apiextensions/v1beta1"
	"kmodules.xyz/client-go/openapi"
)

func generateCRDDefinitions() {
	filename := gort.GOPath() + "/src/github.com/kube-ci/engine/apis/engine/v1alpha1/crds.yaml"
	os.Remove(filename)

	err := os.MkdirAll(filepath.Join(gort.GOPath(), "/src/github.com/kube-ci/engine/api/crds"), 0755)
	if err != nil {
		log.Fatal(err)
	}

	crds := []*crd_api.CustomResourceDefinition{
		v1alpha1.Workflow{}.CustomResourceDefinition(),
		v1alpha1.Workplan{}.CustomResourceDefinition(),
		v1alpha1.WorkflowTemplate{}.CustomResourceDefinition(),
	}
	for _, crd := range crds {
		filename := filepath.Join(gort.GOPath(), "/src/github.com/kube-ci/engine/api/crds", crd.Spec.Names.Singular+".yaml")
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
		crdutils.MarshallCrd(f, crd, "yaml")
		f.Close()
	}
}

func generateSwaggerJson() {
	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)

	api_install.Install(Scheme)

	apispec, err := openapi.RenderOpenAPISpec(openapi.Config{
		Scheme: Scheme,
		Codecs: Codecs,
		Info: spec.InfoProps{
			Title:   "Kubeci-engine",
			Version: "v0.1.0",
			Contact: &spec.ContactInfo{
				Name:  "AppsCode Inc.",
				URL:   "https://appscode.com",
				Email: "hello@appscode.com",
			},
			License: &spec.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			v1alpha1.GetOpenAPIDefinitions,
		},
		Resources: []openapi.TypeInfo{
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourceWorkflows, v1alpha1.ResourceKindWorkflow, true},
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourceWorkplans, v1alpha1.ResourceKindWorkplan, true},
			{v1alpha1.SchemeGroupVersion, v1alpha1.ResourceWorkflowTemplates, v1alpha1.ResourceKindWorkflowTemplate, true},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	filename := gort.GOPath() + "/src/github.com/kube-ci/engine/api/openapi-spec/swagger.json"
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(filename, []byte(apispec), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	generateCRDDefinitions()
	generateSwaggerJson()
}
