package schema

import (
	"reflect"
	"testing"

	"github.com/zebodotdev/httpapi/param"
	"github.com/zebodotdev/httpapi/response"
)

func TestFromParamShapeIncludesStringEnum(t *testing.T) {
	shape := param.ShapeSpec{
		Type: param.TypeString,
		Enum: []string{"product", "fee"},
	}

	got := FromParamShape(shape)
	if got.Type != "string" {
		t.Fatalf("type = %q, want string", got.Type)
	}
	if !reflect.DeepEqual(got.Enum, []string{"product", "fee"}) {
		t.Fatalf("enum = %#v", got.Enum)
	}

	shape.Enum[0] = "mutated"
	if got.Enum[0] != "product" {
		t.Fatalf("enum leaked input mutation: %#v", got.Enum)
	}
}

func TestFromParamShapeIncludesDiscriminatorBranches(t *testing.T) {
	got := FromParamShape(param.ShapeSpec{
		Type: param.TypeObject,
		Discriminator: &param.DiscriminatorSpec{
			Parameter: "type",
			Variants: []param.DiscriminatorVariantSpec{
				{
					Value: "product",
					Shape: param.ShapeSpec{
						Type: param.TypeObject,
						Parameters: []param.ParameterSpec{{
							Name:     "name",
							Required: true,
							Shape:    param.ShapeSpec{Type: param.TypeString},
						}},
					},
				},
				{
					Value: "fee",
					Shape: param.ShapeSpec{
						Type: param.TypeObject,
						Parameters: []param.ParameterSpec{{
							Name:     "amount",
							Required: true,
							Shape:    param.ShapeSpec{Type: param.TypeInt},
						}},
					},
				},
			},
		},
	})

	if got.Discriminator == nil || got.Discriminator.PropertyName != "type" {
		t.Fatalf("discriminator = %#v", got.Discriminator)
	}
	if len(got.OneOf) != 2 {
		t.Fatalf("oneOf = %#v", got.OneOf)
	}
	product := got.OneOf[0]
	if product.Properties["type"].Type != "string" ||
		!reflect.DeepEqual(product.Properties["type"].Enum, []string{"product"}) {
		t.Fatalf("product discriminator property = %#v", product.Properties["type"])
	}
	if product.Properties["name"].Type != "string" {
		t.Fatalf("product name property = %#v", product.Properties["name"])
	}
	if !reflect.DeepEqual(product.Required, []string{"name", "type"}) {
		t.Fatalf("product required = %#v", product.Required)
	}
	fee := got.OneOf[1]
	if fee.Properties["amount"].Type != "integer" {
		t.Fatalf("fee amount property = %#v", fee.Properties["amount"])
	}
}

func TestFromParamShapeSwagger2DowngradesDiscriminator(t *testing.T) {
	got := FromParamShapeSwagger2(param.ShapeSpec{
		Type: param.TypeObject,
		Discriminator: &param.DiscriminatorSpec{
			Parameter: "type",
			Variants: []param.DiscriminatorVariantSpec{
				{
					Value: "product",
					Shape: param.ShapeSpec{
						Type: param.TypeObject,
						Parameters: []param.ParameterSpec{{
							Name:     "name",
							Required: true,
							Shape:    param.ShapeSpec{Type: param.TypeString},
						}},
					},
				},
				{
					Value: "fee",
					Shape: param.ShapeSpec{
						Type: param.TypeObject,
						Parameters: []param.ParameterSpec{{
							Name:     "amount",
							Required: true,
							Shape:    param.ShapeSpec{Type: param.TypeInt},
						}},
					},
				},
			},
		},
	})

	if got.Discriminator != nil || len(got.OneOf) != 0 {
		t.Fatalf("swagger schema contains OpenAPI 3 discriminator data: %#v", got)
	}
	if got.Properties["type"].Type != "string" ||
		!reflect.DeepEqual(got.Properties["type"].Enum, []string{"product", "fee"}) {
		t.Fatalf("type property = %#v", got.Properties["type"])
	}
	if got.Properties["name"].Type != "string" {
		t.Fatalf("name property = %#v", got.Properties["name"])
	}
	if got.Properties["amount"].Type != "integer" {
		t.Fatalf("amount property = %#v", got.Properties["amount"])
	}
	if !reflect.DeepEqual(got.Required, []string{"type"}) {
		t.Fatalf("required = %#v", got.Required)
	}
}

func TestFromResponseShapeIncludesMapAdditionalProperties(t *testing.T) {
	got := FromResponseShape(response.ShapeSpec{
		Type: response.TypeObject,
		MapValue: &response.ShapeSpec{
			Type: response.TypeObject,
			Attributes: []response.AttributeSpec{
				{
					Name:     "status",
					Required: true,
					Shape:    response.ShapeSpec{Type: response.TypeString},
				},
				{
					Name:  "count",
					Shape: response.ShapeSpec{Type: response.TypeInt},
				},
			},
		},
	})

	if got.Type != "object" {
		t.Fatalf("type = %q, want object", got.Type)
	}
	if got.AdditionalProperties == nil {
		t.Fatal("additionalProperties is nil")
	}
	if got.Properties != nil || got.Required != nil {
		t.Fatalf("map schema has fixed object properties: %#v", got)
	}

	value := got.AdditionalProperties
	if value.Type != "object" {
		t.Fatalf("additionalProperties type = %q, want object", value.Type)
	}
	if value.Properties["status"].Type != "string" {
		t.Fatalf("status schema = %#v", value.Properties["status"])
	}
	if value.Properties["count"].Type != "integer" {
		t.Fatalf("count schema = %#v", value.Properties["count"])
	}
	if !reflect.DeepEqual(value.Required, []string{"status"}) {
		t.Fatalf("required = %#v, want status", value.Required)
	}
}
