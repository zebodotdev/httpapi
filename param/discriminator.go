package param

import (
	"fmt"
	"strings"
)

// DiscriminatedObject returns an object shape whose discriminator parameter
// selects the accepted object variant.
//
// Variant object shapes should omit the discriminator parameter. The
// discriminator is parsed once, validated against the declared variants, and
// stripped before the selected variant shape parses the remaining object.
func DiscriminatedObject[T any](discriminator string) *DiscriminatedObjectShape[T] {
	discriminator = strings.TrimSpace(discriminator)
	if discriminator == "" {
		panic("httpapi/param: discriminator parameter is required")
	}
	return &DiscriminatedObjectShape[T]{discriminator: discriminator}
}

// DiscriminatedObjectShape defines a JSON object selected by a discriminator.
type DiscriminatedObjectShape[T any] struct {
	discriminator  string
	variants       []discriminatedObjectVariant[T]
	variantByValue map[string]discriminatedObjectVariant[T]
}

type discriminatedObjectVariant[T any] struct {
	value string
	shape Shape[T]
}

// Variant adds an accepted discriminator value and object shape.
func (shape *DiscriminatedObjectShape[T]) Variant(
	value string,
	variantShape Shape[T],
) *DiscriminatedObjectShape[T] {
	if shape == nil {
		panic("httpapi/param: discriminated object shape is required")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		panic("httpapi/param: discriminator variant value cannot be empty")
	}
	if variantShape == nil {
		panic("httpapi/param: discriminator variant shape is required")
	}
	if variantShape.wireType() != TypeObject {
		panic("httpapi/param: discriminator variant shape must be an object")
	}

	shape.variants = append(shape.variants, discriminatedObjectVariant[T]{
		value: value,
		shape: variantShape,
	})
	shape.variantByValue = nil
	return shape
}

func (shape *DiscriminatedObjectShape[T]) parseShape(
	ctx parseContext,
	path string,
	raw any,
) (T, *Error) {
	if shape == nil {
		panic("httpapi/param: nil discriminated object shape")
	}
	shape.normalize()

	object, ok := raw.(map[string]any)
	if !ok {
		return zeroValue[T](), typeMismatch(path, TypeObject)
	}

	discriminatorPath := joinPath(path, shape.discriminator)
	rawValue, ok := object[shape.discriminator]
	if !ok || rawValue == nil {
		return zeroValue[T](), missingError(discriminatorPath)
	}

	value, ok := rawValue.(string)
	if !ok {
		return zeroValue[T](), typeMismatch(discriminatorPath, TypeString)
	}

	variant, ok := shape.variantByValue[value]
	if !ok {
		return zeroValue[T](), valueNotAllowedError(discriminatorPath, value, shape.variantValues())
	}

	return variant.shape.parseShape(ctx, path, objectWithoutKey(object, shape.discriminator))
}

func (shape *DiscriminatedObjectShape[T]) describeShape() ShapeSpec {
	if shape == nil {
		panic("httpapi/param: nil discriminated object shape")
	}
	shape.normalize()

	variants := make([]DiscriminatorVariantSpec, 0, len(shape.variants))
	for _, variant := range shape.variants {
		variants = append(variants, DiscriminatorVariantSpec{
			Value: variant.value,
			Shape: variant.shape.describeShape(),
		})
	}

	return ShapeSpec{
		Type: TypeObject,
		Discriminator: &DiscriminatorSpec{
			Parameter: shape.discriminator,
			Variants:  variants,
		},
	}
}

func (shape *DiscriminatedObjectShape[T]) wireType() Type {
	return TypeObject
}

func (shape *DiscriminatedObjectShape[T]) normalize() {
	if len(shape.variants) == 0 {
		panic("httpapi/param: discriminator requires at least one variant")
	}
	if shape.variantByValue != nil {
		return
	}

	shape.variantByValue = make(map[string]discriminatedObjectVariant[T], len(shape.variants))
	for _, variant := range shape.variants {
		if _, ok := shape.variantByValue[variant.value]; ok {
			panic(fmt.Sprintf(
				"httpapi/param: duplicate discriminator variant %q",
				variant.value,
			))
		}
		shape.variantByValue[variant.value] = variant
	}
}

func (shape *DiscriminatedObjectShape[T]) variantValues() []string {
	values := make([]string, 0, len(shape.variants))
	for _, variant := range shape.variants {
		values = append(values, variant.value)
	}
	return values
}

func objectWithoutKey(object map[string]any, key string) map[string]any {
	out := make(map[string]any, len(object))
	for name, value := range object {
		if name == key {
			continue
		}
		out[name] = value
	}
	return out
}
