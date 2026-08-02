package endpoint

import (
	"fmt"
	"strings"
)

// CostAccountingMode declares whether operation cost accounting is expected.
type CostAccountingMode string

const (
	// CostAccountingDefault keeps httpapi's default behavior. By default,
	// request cost usage recorded on the request is included in endpoint
	// completion events.
	CostAccountingDefault CostAccountingMode = "default"

	// CostAccountingEnabled explicitly marks an operation as cost-accounted.
	CostAccountingEnabled CostAccountingMode = "enabled"

	// CostAccountingDisabled explicitly marks an operation as not cost-accounted.
	CostAccountingDisabled CostAccountingMode = "disabled"
)

// AccountingSpec carries provider-neutral operation accounting metadata.
//
// It does not define provider, cloud, billing-account, invoice, price-list, or
// reconciliation policy. Cost accounting operation identity comes from
// OperationSpec.ID when present.
type AccountingSpec struct {
	// Cost declares whether usage recorded during endpoint handling should be
	// emitted as a completion cost event.
	Cost CostAccountingMode `json:"cost,omitempty" yaml:"cost,omitempty"`
}

// WithDefaults returns spec with unset accounting fields filled from defaults.
func (spec AccountingSpec) WithDefaults(defaults AccountingSpec) AccountingSpec {
	defaults = NormalizeAccountingSpec(defaults)
	spec = NormalizeAccountingSpec(spec)

	if spec.Cost == CostAccountingDefault {
		spec.Cost = defaults.Cost
	}

	return NormalizeAccountingSpec(spec)
}

// CostAccountingEnabled reports whether cost accounting is enabled for spec.
func (spec AccountingSpec) CostAccountingEnabled() bool {
	return NormalizeAccountingSpec(spec).Cost.Enabled()
}

// Enabled reports whether the cost accounting mode allows cost events.
func (mode CostAccountingMode) Enabled() bool {
	return NormalizeCostAccountingMode(mode) != CostAccountingDisabled
}

// NormalizeAccountingSpec validates endpoint accounting metadata.
func NormalizeAccountingSpec(spec AccountingSpec) AccountingSpec {
	spec.Cost = NormalizeCostAccountingMode(spec.Cost)
	return spec
}

func normalizeEndpointAccountingSpec(spec AccountingSpec) AccountingSpec {
	return NormalizeAccountingSpec(spec)
}

// NormalizeCostAccountingMode trims and canonicalizes a cost accounting mode.
func NormalizeCostAccountingMode(mode CostAccountingMode) CostAccountingMode {
	switch strings.TrimSpace(strings.ToLower(string(mode))) {
	case "", string(CostAccountingDefault):
		return CostAccountingDefault
	case string(CostAccountingEnabled):
		return CostAccountingEnabled
	case string(CostAccountingDisabled):
		return CostAccountingDisabled
	default:
		panic(fmt.Sprintf("httpapi: unsupported cost accounting mode %q", mode))
	}
}
