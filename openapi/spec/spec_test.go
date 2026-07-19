package spec

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestOperationMarshalJSONEmitsExtensionsInline(t *testing.T) {
	operation := Operation{
		OperationID: "create_task",
		Responses: map[string]Response{
			"default": {Description: "fallback"},
		},
	}
	if err := operation.SetExtension("x-example-backend", map[string]any{
		"address": "https://service.example.internal",
	}); err != nil {
		t.Fatalf("SetExtension() error = %v", err)
	}

	encoded, err := json.Marshal(operation)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	value := string(encoded)
	if !strings.Contains(value, `"x-example-backend"`) {
		t.Fatalf("encoded operation missing extension: %s", value)
	}
	if strings.Contains(value, `"Extensions"`) || strings.Contains(value, `"extensions"`) {
		t.Fatalf("encoded operation leaked extension container: %s", value)
	}
}

func TestOperationRejectsInvalidExtensionName(t *testing.T) {
	var operation Operation

	err := operation.SetExtension("backend", "https://service.example.internal")
	if !errors.Is(err, ErrInvalidExtensionName) {
		t.Fatalf("error = %v, want ErrInvalidExtensionName", err)
	}

	operation = Operation{
		Extensions: Extensions{
			"backend": "https://service.example.internal",
		},
	}
	_, err = json.Marshal(operation)
	if !errors.Is(err, ErrInvalidExtensionName) {
		t.Fatalf("marshal error = %v, want ErrInvalidExtensionName", err)
	}
}
