package dependency

import (
	"testing"

	"github.com/TamalSaha/go-oneliners"
	"github.com/kube-ci/experiments/apis/kubeci/v1alpha1"
)

func TestResolveDependency(t *testing.T) {
	if tasks, err := ResolveDependency(v1alpha1.Wf02); err != nil {
		t.Errorf(err.Error())
	} else {
		oneliners.PrettyJson(tasks)
	}
}

func TestDagToLayers(t *testing.T) {
	stepsMap := make(map[string]v1alpha1.Step, 0)
	for _, step := range v1alpha1.Wf02.Spec.Steps {
		stepsMap[step.Name] = step
	}

	if layers, err := dagToLayers(stepsMap); err != nil {
		t.Errorf(err.Error())
	} else {
		oneliners.PrettyJson(layers)
	}
}
