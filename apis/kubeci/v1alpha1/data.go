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
				ApiVersion: "mycrd.k8s.io/v1alpha1",
				Kind:       "Foo",
				Resource:   "foos",
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"workflow": "workflow-01",
					},
				},
			},
		},
		Steps: []Step{
			{
				Name:     "get-file",
				Image:    "kubeci/file-op",
				Commands: []string{"download"},
				Args: []string{
					"--url=http://someurl.com",
					"--path=/kubeci/dir-01",
				},
			},
			{
				Name:     "modify-file",
				Image:    "kubeci/file-op",
				Commands: []string{"copy-and-modify"},
				Args: []string{
					"--copy-from=/kubeci/dir-01",
					"--copy-to=/kubeci/dir-02",
					"--operation=to-upper",
				},
				Dependency: []string{"get-file"},
			},
			{
				Name:     "print-diff",
				Image:    "kubeci/file-op",
				Commands: []string{"diff"},
				Args: []string{
					"--source-dir=/kubeci/dir-01",
					"--modified-dir=/kubeci/dir-02",
					"--store-diff=/kubci/diff.txt",
				},
				Dependency: []string{
					"get-file",
					"modify-file",
				},
			},
			{
				Name:     "publish-output",
				Image:    "kubeci/file-op",
				Commands: []string{"upload-s3"},
				Args: []string{
					"--path=/kubeci/diff.txt",
				},
				Dependency: []string{"print-diff"},
			},
		},
	},
}

var Wf02 = &Workflow{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "workflow-02",
		Namespace: "default",
	},
	Spec: WorkflowSpec{
		Steps: []Step{
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
	},
}

var Wf03 = &Workflow{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "workflow-03",
		Namespace: "default",
	},
	Spec: WorkflowSpec{
		Triggers: []Trigger{
			{
				ApiVersion: "v1",
				Kind:       "Service",
				Resource:   "services",
			},
		},
		Steps: []Step{
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
	},
}

var Wf04 = &Workflow{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "workflow-04",
		Namespace: "default",
	},
	Spec: WorkflowSpec{
		Triggers: []Trigger{
			{
				ApiVersion:       "v1",
				Kind:             "Pod",
				Resource:         "pods",
				Name:             "my-trigger",
				Namespace:        "default",
				OnCreateOrUpdate: true,
				OnDelete:         true,
			},
		},
		Steps: []Step{
			{
				Name:     "echo",
				Image:    "alpine",
				Commands: []string{"ls"},
				Args:     []string{"-la"},
			},
		},
	},
}

var Wf05 = &Workflow{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "workflow-05",
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
