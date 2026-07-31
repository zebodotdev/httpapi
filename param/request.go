package param

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type jsonObject map[string]any

// JSONBuilder defines a JSON request parser before its object parser has been
// attached.
type JSONBuilder[T any] struct {
	def *objectDef[T]
}

// JSON starts a top-level JSON request parser definition.
func JSON[T any]() *JSONBuilder[T] {
	return &JSONBuilder[T]{def: &objectDef[T]{}}
}

// JSONShape returns a JSON request parser whose top-level body is parsed by
// shape.
//
// Use JSONShape when the whole request body is already best described by a
// Shape, such as a DiscriminatedObject whose discriminator lives at the top
// level. Use JSON when defining a plain object by adding top-level parameters
// directly.
func JSONShape[T any](shape Shape[T]) *Request[T] {
	if shape == nil {
		panic("httpapi/param: request shape is required")
	}
	if shape.wireType() != TypeObject {
		panic("httpapi/param: top-level JSON request shape must be an object")
	}
	return &Request[T]{shape: shape}
}

// Param adds an accepted top-level request parameter or parameter group.
func (builder *JSONBuilder[T]) Param(parameter ParamDefinition) *JSONBuilder[T] {
	builder.def.addParam(parameter)
	return builder
}

// ExactlyOne requires the request to include exactly one of names.
func (builder *JSONBuilder[T]) ExactlyOne(names ...string) *JSONBuilder[T] {
	builder.def.addRule(ExactlyOne(names...))
	return builder
}

// AtLeastOne requires the request to include at least one of names.
func (builder *JSONBuilder[T]) AtLeastOne(names ...string) *JSONBuilder[T] {
	builder.def.addRule(AtLeastOne(names...))
	return builder
}

// AtMostOne requires the request to include at most one of names.
func (builder *JSONBuilder[T]) AtMostOne(names ...string) *JSONBuilder[T] {
	builder.def.addRule(AtMostOne(names...))
	return builder
}

// MutuallyExclusive requires the request not to include more than one of names.
func (builder *JSONBuilder[T]) MutuallyExclusive(names ...string) *JSONBuilder[T] {
	builder.def.addRule(AtMostOne(names...))
	return builder
}

// Parse attaches the final request parser and returns the runtime parser.
func (builder *JSONBuilder[T]) Parse(parser ObjectParser[T]) *Request[T] {
	if parser == nil {
		panic("httpapi/param: request parser is required")
	}
	builder.def.parser = parser
	builder.def.normalize()
	return &Request[T]{def: builder.def}
}

// Request parses JSON request bodies into T.
type Request[T any] struct {
	def   *objectDef[T]
	shape Shape[T]
}

// Parse parses input into the request's final domain value.
//
// Input may be an httpapi request source, io.Reader, []byte, string,
// map[string]any, or a JSON object decoded by this package. When input is a
// request source, Parse reads the buffered request body and automatically
// carries its caller, context, and request object into Values.
func (request *Request[T]) Parse(input any, options ...Option) (T, *Error) {
	if request == nil || (request.def == nil && request.shape == nil) {
		panic("httpapi/param: nil request parser")
	}

	options = append(requestSourceOptions(input), options...)
	object, err := decodeInput(input)
	if err != nil {
		return zeroValue[T](), err
	}

	ctx := parseContext{config: newParseConfig(options)}
	if request.shape != nil {
		parsed, err := request.shape.parseShape(ctx, "", map[string]any(object))
		if err != nil {
			return zeroValue[T](), err
		}
		return parsed, nil
	}

	return request.def.parseObject(ctx, "", object)
}

func decodeInput(input any) (jsonObject, *Error) {
	switch value := input.(type) {
	case nil:
		return jsonObject{}, nil
	case jsonObject:
		return value, nil
	case map[string]any:
		return jsonObject(value), nil
	case RequestSource:
		return decodeBytes(value.RequestBody())
	case []byte:
		return decodeBytes(value)
	case string:
		return decodeBytes([]byte(value))
	case io.Reader:
		body, err := io.ReadAll(value)
		if err != nil {
			return nil, invalidBodyError(err)
		}
		return decodeBytes(body)
	default:
		return nil, invalidBodyError(fmt.Errorf("unsupported request body type %T", input))
	}
}

func requestSourceOptions(input any) []Option {
	source, ok := input.(RequestSource)
	if !ok {
		return nil
	}
	return []Option{WithRequest(source)}
}

func decodeBytes(body []byte) (jsonObject, *Error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return jsonObject{}, nil
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()

	var object map[string]any
	if err := dec.Decode(&object); err != nil {
		return nil, invalidBodyError(err)
	}
	if err := ensureNoTrailingJSON(dec); err != nil {
		return nil, invalidBodyError(err)
	}
	if object == nil {
		return nil, invalidBodyError(errors.New("body must be a JSON object"))
	}
	return jsonObject(object), nil
}

func ensureNoTrailingJSON(dec *json.Decoder) error {
	var trailing any
	if err := dec.Decode(&trailing); err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}
	return errors.New("body must contain exactly one JSON object")
}

func invalidBodyError(err error) *Error {
	return paramError(
		"request_body",
		CodeInvalidBody,
		"request body must be a valid JSON object",
		err,
	)
}
