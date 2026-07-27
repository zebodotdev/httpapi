package param

import (
	"fmt"
	"strings"
	"time"
)

// TrimmedString removes leading and trailing whitespace from value.
func TrimmedString(value string) (string, error) {
	return strings.TrimSpace(value), nil
}

// NonEmptyTrimmedString removes leading and trailing whitespace from value and
// rejects the parameter when the resulting string is empty.
func NonEmptyTrimmedString(path string, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", Invalid(path, fmt.Sprintf("`%s` is required", path))
	}
	return value, nil
}

// TrimmedStringList removes leading and trailing whitespace from each item and
// drops items that become empty.
func TrimmedStringList(values []string) ([]string, error) {
	parsed := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			parsed = append(parsed, value)
		}
	}
	return parsed, nil
}

// RFC3339Timestamp removes leading and trailing whitespace from value and
// parses it as an RFC3339 timestamp.
func RFC3339Timestamp(path string, value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, invalidRFC3339Timestamp(path)
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, invalidRFC3339Timestamp(path)
	}
	return parsed, nil
}

// OptionalRFC3339Timestamp removes leading and trailing whitespace from value
// and parses non-empty values as RFC3339 timestamps. Empty values parse to the
// zero time.
func OptionalRFC3339Timestamp(path string, value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	return RFC3339Timestamp(path, value)
}

// OptionalRFC3339TimestampPointer removes leading and trailing whitespace from
// value and parses non-empty values as RFC3339 timestamps. Empty values parse
// to nil, which is useful for optional filter bounds.
func OptionalRFC3339TimestampPointer(path string, value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := RFC3339Timestamp(path, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func invalidRFC3339Timestamp(path string) *Error {
	return Invalid(path, fmt.Sprintf("`%s` must be an RFC3339 timestamp", path))
}
