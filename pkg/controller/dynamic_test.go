package controller

import (
	"encoding/json"
	"testing"

	meta_util "github.com/appscode/kutil/meta"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube.ci/kubeci/apis/kubeci/v1alpha1"
)

func TestJsonPathData(t *testing.T) {
	jsonData := `{
	"Title": "we",
	"Users": {
		"you": "me"
	}
}`

	var structData interface{}
	if err := json.Unmarshal([]byte(jsonData), &structData); err != nil {
		t.Error(err)
	}

	if out := jsonPathData("{$.Users.you}", structData); out != "me" {
		t.Errorf("expected %s, actual %s", "me", out)
	}
}

func TestObjectHashWorkflow(t *testing.T) {
	obj := &v1alpha1.Workflow{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubeci.kube.ci/v1alpha1",
			Kind:       "Workflow",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "wf-01",
			Namespace:  "default",
			Generation: 3,
			Annotations: map[string]string{
				"hello": "world",
			},
			Labels: map[string]string{
				"you": "me",
			},
		},
		Spec: v1alpha1.WorkflowSpec{
			ServiceAccount: "sa-01",
		},
		Status: v1alpha1.WorkflowStatus{
			ObservedGeneration: 2,
		},
	}

	hash := meta_util.ObjectHash(obj)

	// generation changed, hash should change
	objNew := obj.DeepCopy()
	objNew.Generation = 2
	hashNew := meta_util.ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("generation changed, hash should change")
	}

	// annotation changed, hash should change
	objNew = obj.DeepCopy()
	objNew.Annotations["hello"] = "hell"
	hashNew = meta_util.ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("annotation changed, hash should change")
	}

	// labels changed, hash should change
	objNew = obj.DeepCopy()
	objNew.Labels["you"] = "not-me"
	hashNew = meta_util.ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("labels changed, hash should change")
	}

	// spec changed, hash should change
	objNew = obj.DeepCopy()
	objNew.Spec.ServiceAccount = "sa-02"
	hashNew = meta_util.ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("spec changed, hash should change")
	}

	// status changed, hash should not change
	objNew = obj.DeepCopy()
	objNew.Status.ObservedGeneration = 3
	hashNew = meta_util.ObjectHash(objNew)
	if hash != hashNew {
		t.Errorf("status changed, hash should not changee")
	}
}

func TestObjectHashConfigmap(t *testing.T) {
	obj := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "cfg-01",
			Namespace:  "default",
			Generation: 3,
			Annotations: map[string]string{
				"hello": "world",
			},
			Labels: map[string]string{
				"you": "me",
			},
		},
		Data: map[string]string{
			"performance": "average",
		},
	}

	hash := meta_util.ObjectHash(obj)

	// data changed, hash should change
	objNew := obj.DeepCopy()
	objNew.Data["performance"] = "excellent"
	hashNew := meta_util.ObjectHash(objNew)
	if hash == hashNew {
		t.Errorf("data changed, hash should change")
	}
}
