package dependency

import (
	"testing"

	"github.com/appscode/kutil/meta"
	"github.com/kube-ci/engine/apis/engine/v1alpha1"
	"github.com/tamalsaha/go-oneliners"
)

var preSteps = []v1alpha1.Step{
	{
		Name: "pre-step-1",
	},
	{
		Name: "pre-step-2",
	},
}

var postSteps = []v1alpha1.Step{
	{
		Name: "post-step-1",
	},
	{
		Name: "post-step-2",
	},
}

var nonDagSteps = []v1alpha1.Step{
	{
		Name: "A",
	},
	{
		Name: "B",
	},
	{
		Name: "C",
	},
}

var dagStepsData = [][]v1alpha1.Step{
	{
		{
			Name: "A",
		},
		{
			Name:       "B",
			Dependency: []string{"A"},
		},
		{
			Name:       "C",
			Dependency: []string{"A"},
		},
		{
			Name:       "D",
			Dependency: []string{"A", "B"},
		},
		{
			Name:       "E",
			Dependency: []string{"B", "C"},
		},
		{
			Name:       "F",
			Dependency: []string{"A"},
		},
	},
	{
		{
			Name: "A",
		},
		{
			Name:       "B",
			Dependency: []string{"A"},
		},
		{
			Name:       "C",
			Dependency: []string{"B"},
		},
		{
			Name:       "D",
			Dependency: []string{"A", "B"},
		},
		{
			Name:       "E",
			Dependency: []string{"B", "C"},
		},
		{
			Name:       "F",
			Dependency: []string{"D"},
		},
	},
	{
		{
			Name: "A",
		},
		{
			Name:       "B",
			Dependency: []string{"A"},
		},
		{
			Name:       "C",
			Dependency: []string{"B"},
		},
		{
			Name:       "D",
			Dependency: []string{"C"},
		},
	},
}

func TestResolveDependencyForDag(t *testing.T) {
	for _, steps := range dagStepsData {
		if tasks, err := ResolveDependency(steps, preSteps, postSteps, v1alpha1.ExecutionOrderDAG); err != nil {
			t.Errorf(err.Error())
		} else {
			oneliners.PrettyJson(tasks)
		}
	}
}

func TestResolveDependencyForSerial(t *testing.T) {
	if tasks, err := ResolveDependency(nonDagSteps, preSteps, postSteps, v1alpha1.ExecutionOrderSerial); err != nil {
		t.Errorf(err.Error())
	} else {
		oneliners.PrettyJson(tasks)
	}
}

func TestResolveDependencyForParallel(t *testing.T) {
	if tasks, err := ResolveDependency(nonDagSteps, preSteps, postSteps, v1alpha1.ExecutionOrderParallel); err != nil {
		t.Errorf(err.Error())
	} else {
		oneliners.PrettyJson(tasks)
	}
}

func TestDagToLayers(t *testing.T) {
	for _, steps := range dagStepsData {
		stepsMap := make(map[string]v1alpha1.Step, 0)
		for _, step := range steps {
			stepsMap[step.Name] = step
		}

		if layers, err := dagToLayers(stepsMap); err != nil {
			t.Errorf(err.Error())
		} else {
			oneliners.PrettyJson(layers)
		}
	}
}

func TestTasksToLayers(t *testing.T) {
	for _, steps := range dagStepsData {
		stepsMap := make(map[string]v1alpha1.Step, 0)
		for _, step := range steps {
			stepsMap[step.Name] = step
		}
		layers, err := dagToLayers(stepsMap)
		if err != nil {
			t.Errorf(err.Error())
		}
		tasks := layersToTasks(layers)
		layersNew := TasksToLayers(tasks)
		if !meta.Equal(layers, layersNew) {
			t.Errorf("expectd %v, found %v", layers, layersNew)
		}
	}
}
