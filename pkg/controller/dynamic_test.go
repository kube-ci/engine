package controller

import (
	"encoding/json"
	"testing"
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
