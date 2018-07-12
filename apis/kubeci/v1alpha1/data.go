package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Wf01 = &Workflow{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "workflow-01",
		Namespace: "default",
	},
	Spec: WorkflowSpec{
		Triggers: []Trigger{
			{
				ApiVersion:       "v1",
				Kind:             "ConfigMap",
				Resource:         "configmaps",
				Name:             "my-config",
				Namespace:        "default",
				OnCreateOrUpdate: true,
				OnDelete:         true,
			},
		},
		Steps: []Step{
			{
				Name:     "step-1",
				Image:    "alpine",
				Commands: []string{"sh"},
				Args:     []string{"-c", "touch /kubeci/file-1"},
			},
			{
				Name:       "step-2",
				Image:      "alpine",
				Commands:   []string{"sh"},
				Args:       []string{"-c", "touch /kubeci/file-2"},
				Dependency: []string{"step-1"},
			},
			{
				Name:       "step-3",
				Image:      "alpine",
				Commands:   []string{"sh"},
				Args:       []string{"-c", "touch /kubeci/file-3"},
				Dependency: []string{"step-1"},
			},
			{
				Name:       "step-4",
				Image:      "alpine",
				Commands:   []string{"sh"},
				Args:       []string{"-c", "touch /kubeci/file-4"},
				Dependency: []string{"step-2", "step-3"},
			},
			{
				Name:       "step-5",
				Image:      "alpine",
				Commands:   []string{"sh"},
				Args:       []string{"-c", "echo step-5; ls /kubeci"},
				Dependency: []string{"step-1", "step-4"},
			},
		},
	},
}
