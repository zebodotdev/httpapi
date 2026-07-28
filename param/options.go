package param

import (
	"context"

	callerpkg "github.com/zebodotdev/httpapi/caller"
)

// Option configures a parse call.
type Option func(*parseConfig)

// CallerSource exposes the caller attached to a parsed request.
//
// request.Req implements this interface. Keeping the dependency as a tiny
// interface lets param use the request caller without importing the request
// package.
type CallerSource interface {
	// RequestCaller returns the application-defined request caller.
	RequestCaller() callerpkg.Caller
}

type parseConfig struct {
	caller callerpkg.Caller
	ctx    context.Context
	values map[string]any
}

// WithCaller attaches caller information used by parameter availability rules.
//
// Use WithCaller in tests or when parsing outside an httpapi request. Endpoint
// handlers that receive *request.Req or *endpoint.Req should usually use
// WithRequestCaller instead.
func WithCaller(caller callerpkg.Caller) Option {
	return func(config *parseConfig) {
		config.caller = caller
	}
}

// WithRequestCaller attaches caller information from source.
//
// This is the request-bound form of WithCaller. It keeps request parsing aligned
// with endpoint availability and response projection by reading the same caller
// stored on the request.
func WithRequestCaller(source CallerSource) Option {
	return func(config *parseConfig) {
		if source == nil {
			return
		}
		config.caller = source.RequestCaller()
	}
}

// WithContext attaches a context that object parsers can read from Values.
func WithContext(ctx context.Context) Option {
	return func(config *parseConfig) {
		config.ctx = ctx
	}
}

// WithValue attaches application-defined parse context.
func WithValue(key string, value any) Option {
	return func(config *parseConfig) {
		if config.values == nil {
			config.values = map[string]any{}
		}
		config.values[key] = value
	}
}

func newParseConfig(options []Option) parseConfig {
	config := parseConfig{ctx: context.Background()}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	if config.ctx == nil {
		config.ctx = context.Background()
	}
	return config
}
