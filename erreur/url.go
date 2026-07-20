package erreur

import "sync"

// ErrorDoc is the structured input passed to a URLBuilder.
type ErrorDoc struct {
	// Code is the specific machine-readable error code.
	Code string

	// Type is the broad error type.
	Type string

	// Cause is the underlying failure cause.
	Cause string

	// FixCode is the remediation code.
	FixCode string
}

// URLBuilder builds documentation URLs for structured errors.
type URLBuilder func(ErrorDoc) string

var urlBuilderState = struct {
	sync.RWMutex
	builder URLBuilder
}{
	builder: defaultURLBuilder,
}

// ConfigureURLBuilder installs the package-level error documentation URL
// builder.
//
// Passing nil restores the default builder, which returns an empty URL. The
// returned function restores the previous builder.
func ConfigureURLBuilder(builder URLBuilder) func() {
	if builder == nil {
		builder = defaultURLBuilder
	}

	urlBuilderState.Lock()
	prev := urlBuilderState.builder
	urlBuilderState.builder = builder
	urlBuilderState.Unlock()

	return func() {
		urlBuilderState.Lock()
		urlBuilderState.builder = prev
		urlBuilderState.Unlock()
	}
}

// URL returns a documentation URL for an error code using the configured
// builder.
func URL(code string) string {
	return URLFor(code, "", "", "")
}

// URLFor returns a documentation URL for a fully described error using the
// configured builder.
func URLFor(code, typ, cause, fixCode string) string {
	urlBuilderState.RLock()
	builder := urlBuilderState.builder
	urlBuilderState.RUnlock()

	return builder(ErrorDoc{
		Code:    code,
		Type:    typ,
		Cause:   cause,
		FixCode: fixCode,
	})
}

func defaultURLBuilder(doc ErrorDoc) string {
	return ""
}
