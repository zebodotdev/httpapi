package cost

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	// MaxQuantityScale is the largest decimal scale accepted by Quantity.
	//
	// Quantity stores an int64 mantissa and a base-10 scale. Limiting scale to
	// eighteen decimal places keeps string parsing and rendering bounded while
	// still covering common cloud-provider fractional units such as CPU seconds,
	// memory GiB-seconds, request units, and provider-call counts.
	MaxQuantityScale = 18
)

// Quantity is an exact decimal quantity suitable for usage observations.
//
// A Quantity stores value * 10^-scale. For example, Decimal(125, 2) represents
// 1.25 and Whole(3) represents 3. The type deliberately avoids float64 so cost
// sinks can aggregate and price usage without binary floating-point drift.
//
// The zero value represents 0.
type Quantity struct {
	value int64
	scale int32
}

// Whole returns an integer Quantity.
func Whole(value int64) Quantity {
	return Quantity{value: value}
}

// Decimal returns a decimal Quantity from an integer mantissa and decimal
// scale.
//
// Decimal(125, 2) represents 1.25. Decimal returns an error when scale is
// negative or larger than MaxQuantityScale.
func Decimal(value int64, scale int) (Quantity, error) {
	if err := validateQuantityScale(scale); err != nil {
		return Quantity{}, err
	}
	return Quantity{value: value, scale: int32(scale)}, nil
}

// MustDecimal returns Decimal(value, scale) and panics when the scale is
// invalid.
//
// MustDecimal is useful for package-level constants and tests where invalid
// scale values should fail during startup.
func MustDecimal(value int64, scale int) Quantity {
	quantity, err := Decimal(value, scale)
	if err != nil {
		panic(err.Error())
	}
	return quantity
}

// ParseQuantity parses an exact base-10 decimal string into a Quantity.
//
// ParseQuantity accepts forms such as "1", "1.25", ".5", "-0.5", and
// "+10.000". Scientific notation is intentionally not accepted because usage
// observations should be easy to inspect in durable logs and audit records.
func ParseQuantity(raw string) (Quantity, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Quantity{}, fmt.Errorf("httpapi/cost: quantity is required")
	}

	sign := ""
	switch raw[0] {
	case '+':
		raw = raw[1:]
	case '-':
		sign = "-"
		raw = raw[1:]
	}
	if raw == "" {
		return Quantity{}, fmt.Errorf("httpapi/cost: quantity is required")
	}

	parts := strings.Split(raw, ".")
	if len(parts) > 2 {
		return Quantity{}, fmt.Errorf("httpapi/cost: quantity %q is not a decimal", raw)
	}

	whole := parts[0]
	fractional := ""
	if len(parts) == 2 {
		fractional = parts[1]
	}
	if whole == "" && fractional == "" {
		return Quantity{}, fmt.Errorf("httpapi/cost: quantity is required")
	}
	if !decimalDigits(whole) || !decimalDigits(fractional) {
		return Quantity{}, fmt.Errorf("httpapi/cost: quantity %q contains non-decimal digits", raw)
	}
	if err := validateQuantityScale(len(fractional)); err != nil {
		return Quantity{}, err
	}

	digits := strings.TrimLeft(whole+fractional, "0")
	if digits == "" {
		digits = "0"
	}
	if sign == "-" && digits != "0" {
		digits = "-" + digits
	}
	value, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return Quantity{}, fmt.Errorf("httpapi/cost: quantity %q is out of range: %w", raw, err)
	}

	return Quantity{value: value, scale: int32(len(fractional))}, nil
}

// Value returns the integer mantissa for q.
func (q Quantity) Value() int64 {
	return q.value
}

// Scale returns the base-10 decimal scale for q.
func (q Quantity) Scale() int {
	return int(q.scale)
}

// IsZero reports whether q represents zero.
func (q Quantity) IsZero() bool {
	return q.value == 0
}

// Valid returns an error when q cannot be represented by httpapi's decimal
// quantity format.
func (q Quantity) Valid() error {
	return validateQuantityScale(int(q.scale))
}

// String returns q as a base-10 decimal string.
func (q Quantity) String() string {
	if q.scale == 0 {
		return strconv.FormatInt(q.value, 10)
	}

	scale := int(q.scale)
	digits := strconv.FormatInt(q.value, 10)
	sign := ""
	if strings.HasPrefix(digits, "-") {
		sign = "-"
		digits = strings.TrimPrefix(digits, "-")
	}

	for len(digits) <= scale {
		digits = "0" + digits
	}
	point := len(digits) - scale
	return sign + digits[:point] + "." + digits[point:]
}

// MarshalText encodes q as its base-10 decimal string.
func (q Quantity) MarshalText() ([]byte, error) {
	if err := q.Valid(); err != nil {
		return nil, err
	}
	return []byte(q.String()), nil
}

// UnmarshalText decodes q from a base-10 decimal string.
func (q *Quantity) UnmarshalText(text []byte) error {
	if q == nil {
		return fmt.Errorf("httpapi/cost: cannot unmarshal into nil Quantity")
	}
	quantity, err := ParseQuantity(string(text))
	if err != nil {
		return err
	}
	*q = quantity
	return nil
}

// MarshalJSON encodes q as a JSON string containing its base-10 decimal value.
func (q Quantity) MarshalJSON() ([]byte, error) {
	text, err := q.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(text))
}

// UnmarshalJSON decodes q from a JSON string containing a base-10 decimal
// value.
func (q *Quantity) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("httpapi/cost: quantity must be a JSON string: %w", err)
	}
	return q.UnmarshalText([]byte(raw))
}

func validateQuantityScale(scale int) error {
	if scale < 0 {
		return fmt.Errorf("httpapi/cost: quantity scale cannot be negative: %d", scale)
	}
	if scale > MaxQuantityScale {
		return fmt.Errorf(
			"httpapi/cost: quantity scale %d exceeds max scale %d",
			scale,
			MaxQuantityScale,
		)
	}
	return nil
}

func decimalDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
