package endpoint

import "testing"

func TestAccountingSpecNormalization(t *testing.T) {
	tests := []struct {
		name    string
		spec    AccountingSpec
		want    AccountingSpec
		enabled bool
	}{
		{
			name:    "zero defaults",
			spec:    AccountingSpec{},
			want:    AccountingSpec{Cost: CostAccountingDefault},
			enabled: true,
		},
		{
			name:    "explicit default",
			spec:    AccountingSpec{Cost: " default "},
			want:    AccountingSpec{Cost: CostAccountingDefault},
			enabled: true,
		},
		{
			name:    "enabled",
			spec:    AccountingSpec{Cost: " enabled "},
			want:    AccountingSpec{Cost: CostAccountingEnabled},
			enabled: true,
		},
		{
			name:    "disabled",
			spec:    AccountingSpec{Cost: " disabled "},
			want:    AccountingSpec{Cost: CostAccountingDisabled},
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeAccountingSpec(tt.spec)
			if got != tt.want {
				t.Fatalf("NormalizeAccountingSpec() = %#v, want %#v", got, tt.want)
			}
			if got.CostAccountingEnabled() != tt.enabled {
				t.Fatalf("CostAccountingEnabled() = %t, want %t", got.CostAccountingEnabled(), tt.enabled)
			}
		})
	}
}

func TestAccountingSpecRejectsUnknownCostMode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NormalizeAccountingSpec did not panic")
		}
	}()

	NormalizeAccountingSpec(AccountingSpec{Cost: "metered"})
}

func TestAccountingSpecWithDefaults(t *testing.T) {
	got := AccountingSpec{}.WithDefaults(AccountingSpec{
		Cost: CostAccountingDisabled,
	})
	if got.Cost != CostAccountingDisabled {
		t.Fatalf("inherited cost = %q, want %q", got.Cost, CostAccountingDisabled)
	}

	got = AccountingSpec{Cost: CostAccountingEnabled}.WithDefaults(AccountingSpec{
		Cost: CostAccountingDisabled,
	})
	if got.Cost != CostAccountingEnabled {
		t.Fatalf("explicit cost = %q, want %q", got.Cost, CostAccountingEnabled)
	}
}

func TestOperationSpecWithDefaults(t *testing.T) {
	got := OperationSpec{
		ID: "create_order",
	}.WithDefaults(OperationSpec{
		ID:      "group_operation",
		Summary: "Orders endpoint",
		Accounting: AccountingSpec{
			Cost: CostAccountingDisabled,
		},
	})
	if got.ID != "create_order" {
		t.Fatalf("operation id = %q, want create_order", got.ID)
	}
	if got.Summary != "Orders endpoint" {
		t.Fatalf("summary = %q, want Orders endpoint", got.Summary)
	}
	if got.Accounting.Cost != CostAccountingDisabled {
		t.Fatalf("accounting = %#v, want disabled", got.Accounting)
	}

	got = OperationSpec{
		ID:      "create_order",
		Summary: "Create order",
		Accounting: AccountingSpec{
			Cost: CostAccountingEnabled,
		},
	}.WithDefaults(OperationSpec{
		Summary: "Orders endpoint",
		Accounting: AccountingSpec{
			Cost: CostAccountingDisabled,
		},
	})
	if got.Summary != "Create order" {
		t.Fatalf("explicit summary = %q, want Create order", got.Summary)
	}
	if got.Accounting.Cost != CostAccountingEnabled {
		t.Fatalf("explicit accounting = %#v, want enabled", got.Accounting)
	}
}

func TestEndpointGroupOperationDefaultsAndEndpointOverrides(t *testing.T) {
	group := EndpointGroup{
		Operation: OperationSpec{
			ID:      "group_operation",
			Summary: "Orders endpoint",
			Accounting: AccountingSpec{
				Cost: CostAccountingEnabled,
			},
		},
		Endpoints: []Endpoint{
			DefineEndpoint(EndpointSpec{
				Method:  POST,
				Path:    "/inherited",
				Handler: noopTranscriptionHandler,
			}),
			DefineEndpoint(EndpointSpec{
				Method:  POST,
				Path:    "/disabled",
				Handler: noopTranscriptionHandler,
				Operation: OperationSpec{
					ID:      "disabled_operation",
					Summary: "Disabled operation",
					Accounting: AccountingSpec{
						Cost: CostAccountingDisabled,
					},
				},
			}),
		},
	}

	resolved := group.ResolvedEndpoints()
	if len(resolved) != 2 {
		t.Fatalf("resolved endpoints = %d, want 2", len(resolved))
	}
	if got := resolved[0].Operation().ID; got != "" {
		t.Fatalf("inherited operation id = %q, want empty", got)
	}
	if got := resolved[0].Operation().Summary; got != "Orders endpoint" {
		t.Fatalf("inherited summary = %q, want Orders endpoint", got)
	}
	if got := resolved[0].Operation().Accounting.Cost; got != CostAccountingEnabled {
		t.Fatalf("inherited cost accounting = %q, want %q", got, CostAccountingEnabled)
	}
	if !resolved[0].CostAccountingEnabled() {
		t.Fatal("inherited endpoint did not enable cost accounting")
	}
	if got := resolved[1].Operation().ID; got != "disabled_operation" {
		t.Fatalf("endpoint operation id = %q, want disabled_operation", got)
	}
	if got := resolved[1].Operation().Summary; got != "Disabled operation" {
		t.Fatalf("endpoint summary = %q, want Disabled operation", got)
	}
	if got := resolved[1].Operation().Accounting.Cost; got != CostAccountingDisabled {
		t.Fatalf("endpoint cost accounting = %q, want %q", got, CostAccountingDisabled)
	}
	if resolved[1].CostAccountingEnabled() {
		t.Fatal("explicitly disabled endpoint enabled cost accounting")
	}
}

func TestEndpointGroupConfigureOperationSpecUpdatesInheritedEndpoints(t *testing.T) {
	group := EndpointGroup{}
	group.Add(DefineEndpoint(EndpointSpec{
		Method:  POST,
		Path:    "/inherited",
		Handler: noopTranscriptionHandler,
	}))
	group.Add(DefineEndpoint(EndpointSpec{
		Method:  POST,
		Path:    "/explicit",
		Handler: noopTranscriptionHandler,
		Operation: OperationSpec{
			ID:      "explicit_operation",
			Summary: "Explicit operation",
			Accounting: AccountingSpec{
				Cost: CostAccountingEnabled,
			},
		},
	}))

	group.ConfigureOperationSpec(OperationSpec{
		ID:      "group_operation",
		Summary: "Group operation",
		Accounting: AccountingSpec{
			Cost: CostAccountingDisabled,
		},
	})

	if got := group.Endpoints[0].Operation().ID; got != "" {
		t.Fatalf("configured inherited operation id = %q, want empty", got)
	}
	if got := group.Endpoints[0].Operation().Summary; got != "Group operation" {
		t.Fatalf("configured inherited summary = %q, want Group operation", got)
	}
	if got := group.Endpoints[0].Operation().Accounting.Cost; got != CostAccountingDisabled {
		t.Fatalf("configured inherited cost accounting = %q, want %q", got, CostAccountingDisabled)
	}
	if got := group.Endpoints[1].Operation().ID; got != "explicit_operation" {
		t.Fatalf("configured explicit operation id = %q, want explicit_operation", got)
	}
	if got := group.Endpoints[1].Operation().Summary; got != "Explicit operation" {
		t.Fatalf("configured explicit summary = %q, want Explicit operation", got)
	}
	if got := group.Endpoints[1].Operation().Accounting.Cost; got != CostAccountingEnabled {
		t.Fatalf("configured explicit cost accounting = %q, want %q", got, CostAccountingEnabled)
	}
}
