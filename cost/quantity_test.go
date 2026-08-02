package cost

import (
	"encoding/json"
	"testing"
)

func TestQuantityDecimalStringAndJSON(t *testing.T) {
	quantity := MustDecimal(125, 2)

	if got := quantity.String(); got != "1.25" {
		t.Fatalf("String() = %q, want 1.25", got)
	}

	encoded, err := json.Marshal(quantity)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
	if string(encoded) != `"1.25"` {
		t.Fatalf("json = %s, want %q", encoded, `"1.25"`)
	}

	var decoded Quantity
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if decoded != quantity {
		t.Fatalf("decoded = %v/%d, want %v/%d", decoded.Value(), decoded.Scale(), quantity.Value(), quantity.Scale())
	}
}

func TestParseQuantity(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantValue int64
		wantScale int
		wantText  string
	}{
		{
			name:      "integer",
			raw:       "42",
			wantValue: 42,
			wantText:  "42",
		},
		{
			name:      "fraction",
			raw:       ".5",
			wantValue: 5,
			wantScale: 1,
			wantText:  "0.5",
		},
		{
			name:      "negative fraction",
			raw:       "-0.25",
			wantValue: -25,
			wantScale: 2,
			wantText:  "-0.25",
		},
		{
			name:      "preserves scale",
			raw:       "+10.000",
			wantValue: 10000,
			wantScale: 3,
			wantText:  "10.000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuantity(tt.raw)
			if err != nil {
				t.Fatalf("ParseQuantity returned error: %v", err)
			}
			if got.Value() != tt.wantValue || got.Scale() != tt.wantScale {
				t.Fatalf(
					"quantity = value %d scale %d, want value %d scale %d",
					got.Value(),
					got.Scale(),
					tt.wantValue,
					tt.wantScale,
				)
			}
			if got.String() != tt.wantText {
				t.Fatalf("String() = %q, want %q", got.String(), tt.wantText)
			}
		})
	}
}

func TestParseQuantityRejectsUnsupportedForms(t *testing.T) {
	for _, raw := range []string{"", ".", "1.2.3", "1e3", "abc"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParseQuantity(raw); err == nil {
				t.Fatal("ParseQuantity returned nil error")
			}
		})
	}
}
