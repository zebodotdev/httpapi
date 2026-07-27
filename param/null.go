package param

// NullPolicy defines how a present JSON null is interpreted for a parameter.
type NullPolicy string

const (
	// NullAsAbsent treats null as if the key was not supplied. It is the
	// default policy.
	NullAsAbsent NullPolicy = "as_absent"

	// NullRejected rejects a present null value.
	NullRejected NullPolicy = "rejected"

	// NullAccepted accepts null as a present value. Parsers receive the zero
	// value for the parameter's final type.
	NullAccepted NullPolicy = "accepted"
)

func normalizeNullPolicy(policy NullPolicy) NullPolicy {
	switch policy {
	case "", NullAsAbsent:
		return NullAsAbsent
	case NullRejected, NullAccepted:
		return policy
	default:
		panic("httpapi/param: unsupported null policy")
	}
}
