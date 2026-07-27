package param

import (
	"context"

	callerpkg "github.com/zebodotdev/httpapi/caller"
)

// Option configures a parse call.
type Option func(*parseConfig)

type parseConfig struct {
	caller callerpkg.Caller
	ctx    context.Context
	values map[string]any
}

// WithCaller attaches caller information used by parameter availability rules.
func WithCaller(caller callerpkg.Caller) Option {
	return func(config *parseConfig) {
		config.caller = caller
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
