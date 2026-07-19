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

func TestOperationMarshalYAMLEmitsExtensionsInline(t *testing.T) {
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

	encoded, err := operation.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML() error = %v", err)
	}
	object, ok := encoded.(map[string]any)
	if !ok {
		t.Fatalf("MarshalYAML() = %T, want map[string]any", encoded)
	}
	if _, ok := object["x-example-backend"]; !ok {
		t.Fatalf("yaml object missing extension: %#v", object)
	}
	if _, ok := object["Extensions"]; ok {
		t.Fatalf("yaml object leaked extension container: %#v", object)
	}
	if _, ok := object["extensions"]; ok {
		t.Fatalf("yaml object leaked extension container: %#v", object)
	}
}

func TestOperationMarshalJSONEmitsEmptyResponsesObject(t *testing.T) {
	encoded, err := json.Marshal(Operation{OperationID: "empty_responses"})
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if !strings.Contains(string(encoded), `"responses":{}`) {
		t.Fatalf("encoded operation responses = %s, want empty object", encoded)
	}
}

func TestOperationExtensionTrimsName(t *testing.T) {
	var operation Operation
	if err := operation.SetExtension(" x-example ", "ok"); err != nil {
		t.Fatalf("SetExtension() error = %v", err)
	}

	value, ok := operation.Extension("x-example")
	if !ok {
		t.Fatal("Extension() missing trimmed extension name")
	}
	if value != "ok" {
		t.Fatalf("Extension() = %v, want ok", value)
	}
}

func TestOperationSetExtensionReplacesAllNormalizedDuplicates(t *testing.T) {
	operation := Operation{
		Extensions: Extensions{
			" x-example": "first",
			"x-example ": "second",
		},
	}

	if err := operation.SetExtension("x-example", "final"); err != nil {
		t.Fatalf("SetExtension() error = %v", err)
	}
	if len(operation.Extensions) != 1 {
		t.Fatalf("extension count = %d, want 1: %#v", len(operation.Extensions), operation.Extensions)
	}
	value, ok := operation.Extension("x-example")
	if !ok || value != "final" {
		t.Fatalf("Extension() = %#v, %t; want final, true", value, ok)
	}
}

func TestOperationRejectsInvalidExtensionName(t *testing.T) {
	var operation Operation

	err := operation.SetExtension("backend", "https://service.example.internal")
	if !errors.Is(err, ErrInvalidExtensionName) {
		t.Fatalf("error = %v, want ErrInvalidExtensionName", err)
	}
	err = operation.SetExtension("X-example", "https://service.example.internal")
	if !errors.Is(err, ErrInvalidExtensionName) {
		t.Fatalf("uppercase error = %v, want ErrInvalidExtensionName", err)
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

func TestOperationRejectsDuplicateNormalizedExtensionName(t *testing.T) {
	operation := Operation{
		Extensions: Extensions{
			"x-example":  "first",
			" x-example": "second",
		},
	}

	_, err := json.Marshal(operation)
	if !errors.Is(err, ErrDuplicateExtensionName) {
		t.Fatalf("marshal error = %v, want ErrDuplicateExtensionName", err)
	}
}

func TestPathsAddOperationAndMergeValidateDestination(t *testing.T) {
	var paths Paths
	err := paths.AddOperation("/tasks", "POST", Operation{})
	if err == nil {
		t.Fatal("AddOperation() error = nil, want error")
	}

	err = paths.Merge(Paths{"/tasks": {Post: &Operation{}}})
	if err == nil {
		t.Fatal("Merge() error = nil, want error")
	}
}

func TestPathItemSupportsStandardOpenAPIMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
		assert func(PathItem) bool
	}{
		{name: "get", method: "get", assert: func(item PathItem) bool { return item.Get != nil }},
		{name: "put", method: "PUT", assert: func(item PathItem) bool { return item.Put != nil }},
		{name: "post", method: "POST", assert: func(item PathItem) bool { return item.Post != nil }},
		{name: "delete", method: "DELETE", assert: func(item PathItem) bool { return item.Delete != nil }},
		{name: "options", method: "OPTIONS", assert: func(item PathItem) bool { return item.Options != nil }},
		{name: "head", method: "HEAD", assert: func(item PathItem) bool { return item.Head != nil }},
		{name: "patch", method: "PATCH", assert: func(item PathItem) bool { return item.Patch != nil }},
		{name: "trace", method: "TRACE", assert: func(item PathItem) bool { return item.Trace != nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var item PathItem
			if err := item.SetOperation(tt.method, Operation{}); err != nil {
				t.Fatalf("SetOperation() error = %v", err)
			}
			if !tt.assert(item) {
				t.Fatalf("SetOperation(%q) did not set expected operation: %#v", tt.method, item)
			}
		})
	}
}
