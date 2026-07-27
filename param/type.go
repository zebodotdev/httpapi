package param

import "fmt"

// Type describes the JSON wire shape accepted before custom parsing runs.
type Type string

const (
	// TypeString accepts JSON string values.
	TypeString Type = "string"

	// TypeInt accepts JSON numbers that can be represented exactly as int.
	TypeInt Type = "int"

	// TypeInt64 accepts JSON numbers that can be represented exactly as int64.
	TypeInt64 Type = "int64"

	// TypeFloat64 accepts JSON numbers and converts them to float64.
	TypeFloat64 Type = "float64"

	// TypeBool accepts JSON boolean values.
	TypeBool Type = "bool"

	// TypeObject accepts JSON object values.
	TypeObject Type = "object"

	// TypeArray accepts JSON array values.
	TypeArray Type = "array"

	// TypeAny accepts any non-null JSON value.
	TypeAny Type = "any"
)

func typeDisplayName(typ Type) string {
	switch typ {
	case TypeString:
		return "a string"
	case TypeInt, TypeInt64:
		return "an integer"
	case TypeFloat64:
		return "a number"
	case TypeBool:
		return "a boolean"
	case TypeObject:
		return "an object"
	case TypeArray:
		return "an array"
	case TypeAny:
		return "a value"
	default:
		return fmt.Sprintf("a %s value", typ)
	}
}
