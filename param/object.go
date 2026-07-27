package param

import (
	"fmt"
	"strings"
)

// ObjectParser converts accepted object parameters into the object's final
// domain value.
type ObjectParser[T any] func(Values) (T, error)

// ObjectShape defines an accepted JSON object shape.
type ObjectShape[T any] struct {
	def *objectDef[T]
}

// Object returns a JSON object shape that can be used for a request parameter.
func Object[T any]() *ObjectShape[T] {
	return &ObjectShape[T]{def: &objectDef[T]{}}
}

// Param adds an accepted parameter to this object shape.
func (shape *ObjectShape[T]) Param(parameter AcceptedParam) *ObjectShape[T] {
	shape.def.addParam(parameter)
	return shape
}

// ExactlyOne requires this object to include exactly one of names.
func (shape *ObjectShape[T]) ExactlyOne(names ...string) *ObjectShape[T] {
	shape.def.addRule(ExactlyOne(names...))
	return shape
}

// AtLeastOne requires this object to include at least one of names.
func (shape *ObjectShape[T]) AtLeastOne(names ...string) *ObjectShape[T] {
	shape.def.addRule(AtLeastOne(names...))
	return shape
}

// AtMostOne requires this object to include at most one of names.
func (shape *ObjectShape[T]) AtMostOne(names ...string) *ObjectShape[T] {
	shape.def.addRule(AtMostOne(names...))
	return shape
}

// MutuallyExclusive requires this object not to include more than one of names.
func (shape *ObjectShape[T]) MutuallyExclusive(names ...string) *ObjectShape[T] {
	shape.def.addRule(MutuallyExclusive(names...))
	return shape
}

// Parse attaches the object-level parser and returns this shape.
func (shape *ObjectShape[T]) Parse(parser ObjectParser[T]) *ObjectShape[T] {
	if parser == nil {
		panic("httpapi/param: object parser is required")
	}
	shape.def.parser = parser
	shape.def.normalize()
	return shape
}

func (shape *ObjectShape[T]) parseShape(ctx parseContext, path string, raw any) (T, *Error) {
	object, ok := raw.(map[string]any)
	if !ok {
		return zeroValue[T](), typeMismatch(path, TypeObject)
	}
	return shape.def.parseObject(ctx, path, jsonObject(object))
}

func (shape *ObjectShape[T]) wireType() Type {
	return TypeObject
}

type objectDef[T any] struct {
	params []AcceptedParam
	rules  []rule
	parser ObjectParser[T]
	names  map[string]AcceptedParam
}

func (def *objectDef[T]) addParam(parameter AcceptedParam) {
	if parameter == nil {
		panic("httpapi/param: parameter is required")
	}
	def.params = append(def.params, parameter)
}

func (def *objectDef[T]) addRule(rule rule) {
	if rule == nil {
		panic("httpapi/param: rule is required")
	}
	def.rules = append(def.rules, normalizeRule(rule))
}

func (def *objectDef[T]) normalize() {
	def.names = make(map[string]AcceptedParam, len(def.params))
	for _, parameter := range def.params {
		name := parameter.paramName()
		if _, ok := def.names[name]; ok {
			panic(fmt.Sprintf("httpapi/param: duplicate parameter %q", name))
		}
		def.names[name] = parameter
	}
	for _, rule := range def.rules {
		for _, name := range rule.names() {
			if _, ok := def.names[name]; !ok {
				panic(fmt.Sprintf("httpapi/param: rule references unknown parameter %q", name))
			}
		}
	}
	if len(def.params) > 0 && def.parser == nil {
		panic("httpapi/param: object parser is required when object parameters are declared")
	}
}

func (def *objectDef[T]) parseObject(
	ctx parseContext,
	path string,
	object jsonObject,
) (T, *Error) {
	def.normalize()

	for name := range object {
		parameter, ok := def.names[name]
		paramPath := joinPath(path, name)
		if !ok ||
			(!parameter.paramAvailable(ctx.config.caller) &&
				parameter.paramPresent(object[name])) {
			return zeroValue[T](), unexpectedError(paramPath)
		}
	}

	values := Values{
		path:   path,
		params: make(map[string]parsedParam, len(def.params)),
		ctx:    ctx.config.ctx,
		extra:  ctx.config.values,
	}
	for _, parameter := range def.params {
		if !parameter.paramAvailable(ctx.config.caller) {
			continue
		}

		parsed, err := parameter.parseParameter(ctx, path, object)
		if err != nil {
			return zeroValue[T](), err
		}
		values.params[parameter.paramName()] = parsed
	}

	for _, rule := range def.rules {
		if err := rule.apply(values); err != nil {
			return zeroValue[T](), err
		}
	}

	if def.parser == nil {
		parsed, err := decodeJSONLike[T](map[string]any(object))
		if err != nil {
			return zeroValue[T](), parserError(pathOrBody(path), err)
		}
		return parsed, nil
	}

	parsed, err := def.parser(values)
	if err != nil {
		return zeroValue[T](), parserError(pathOrBody(path), err)
	}
	return parsed, nil
}

func normalizeRule(r rule) rule {
	switch typed := r.(type) {
	case presenceRule:
		return normalizePresenceRule(typed)
	default:
		return r
	}
}

func unexpectedError(path string) *Error {
	return paramError(
		path,
		CodeUnexpected,
		fmt.Sprintf("`%s` is an unexpected parameter", path),
		nil,
	)
}

func pathOrBody(path string) string {
	if strings.TrimSpace(path) == "" {
		return "request_body"
	}
	return path
}
