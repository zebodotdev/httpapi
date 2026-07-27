package caller

import (
	"encoding/json"
	"testing"
)

func TestDefineTrimsAndStoresName(t *testing.T) {
	api := Define(" API ")
	if api.Name() != "api" {
		t.Fatalf("name = %q", api.Name())
	}
	if !api.Defined() {
		t.Fatal("caller not defined")
	}
	if api != Define("api") {
		t.Fatal("caller definitions are not stable comparable values")
	}
}

func TestParseRejectsInvalidNames(t *testing.T) {
	tests := []string{
		"",
		" ",
		"bad caller",
		"bad\ncaller",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := Parse(input); err == nil {
				t.Fatal("Parse error = nil")
			}
		})
	}
}

func TestSetOfAllowsDefinedCallers(t *testing.T) {
	publicAPI := Define("public-api")
	worker := Define("worker")
	dashboard := Define("dashboard")

	set := SetOf(worker, dashboard)
	if set.Allows(publicAPI) {
		t.Fatal("public API unexpectedly allowed")
	}
	if !set.Allows(worker) {
		t.Fatal("worker not allowed")
	}
	if SetOf().Allows(Caller{}) != true {
		t.Fatal("unrestricted set should allow zero caller")
	}
}

func TestAvailableToRequiresCallers(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("AvailableTo did not panic")
		}
	}()

	AvailableTo()
}

func TestSetIntersectionDoesNotWidenRestrictions(t *testing.T) {
	publicAPI := Define("public-api")
	worker := Define("worker")
	dashboard := Define("dashboard")

	got := SetOf(publicAPI, worker).Intersect(SetOf(worker, dashboard))
	if got.Allows(publicAPI) {
		t.Fatal("intersection allowed public API")
	}
	if got.Allows(dashboard) {
		t.Fatal("intersection allowed dashboard")
	}
	if !got.Allows(worker) {
		t.Fatal("intersection did not allow worker")
	}
}

func TestSetIntersectionCanDenyAll(t *testing.T) {
	publicAPI := Define("public-api")
	worker := Define("worker")

	got := SetOf(publicAPI).Intersect(SetOf(worker))
	if !got.Restricted() {
		t.Fatal("disjoint intersection became unrestricted")
	}
	if got.Allows(publicAPI) {
		t.Fatal("disjoint intersection allowed public API")
	}
	if got.Allows(worker) {
		t.Fatal("disjoint intersection allowed worker")
	}
}

func TestCallerMarshalJSON(t *testing.T) {
	encoded, err := json.Marshal(Define("worker"))
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	if string(encoded) != `"worker"` {
		t.Fatalf("encoded caller = %s", encoded)
	}
}
