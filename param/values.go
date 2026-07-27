package param

import "context"

type parsedParam struct {
	value   any
	present bool
}

// Values contains parameters accepted and parsed for one object.
type Values struct {
	path   string
	params map[string]parsedParam
	ctx    context.Context
	extra  map[string]any
}

// Present reports whether name was supplied and accepted as present.
func (values Values) Present(name string) bool {
	value, ok := values.params[name]
	return ok && value.present
}

// Context returns the parse context configured with WithContext.
func (values Values) Context() context.Context {
	if values.ctx == nil {
		return context.Background()
	}
	return values.ctx
}

// Value returns an application-defined parse value configured with WithValue.
func (values Values) Value(key string) (any, bool) {
	value, ok := values.extra[key]
	return value, ok
}

// Get returns the parsed value for name when it was present.
func Get[T any](values Values, name string) (T, bool) {
	parsed, ok := values.params[name]
	if !ok || !parsed.present {
		return zeroValue[T](), false
	}
	value, ok := parsed.value.(T)
	if !ok {
		return zeroValue[T](), false
	}
	return value, true
}

// Must returns the parsed value for name.
//
// Must panics when the value is absent or has a different type. It is intended
// for object parsers reading parameters declared as required in the same
// payload definition.
func Must[T any](values Values, name string) T {
	value, ok := Get[T](values, name)
	if !ok {
		panic("httpapi/param: missing parsed parameter " + name)
	}
	return value
}
