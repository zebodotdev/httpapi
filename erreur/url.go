package erreur

import "sync"

type ErrorDoc struct {
	Code    string
	Type    string
	Cause   string
	FixCode string
}

type URLBuilder func(ErrorDoc) string

var urlBuilderState = struct {
	sync.RWMutex
	builder URLBuilder
}{
	builder: defaultURLBuilder,
}

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

func URL(code string) string {
	return URLFor(code, "", "", "")
}

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
