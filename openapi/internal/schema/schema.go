package schema

import (
	"github.com/zebodotdev/httpapi/openapi/spec"
	parampkg "github.com/zebodotdev/httpapi/param"
	responsepkg "github.com/zebodotdev/httpapi/response"
)

// FromParamShape translates a provider-neutral param shape into an OpenAPI
// schema.
func FromParamShape(shape parampkg.ShapeSpec) spec.Schema {
	schema := paramTypeSchema(shape.Type)
	switch shape.Type {
	case parampkg.TypeObject:
		schema.Properties = paramProperties(shape.Parameters)
		schema.Required = paramRequired(shape.Parameters)
	case parampkg.TypeArray:
		schema.Items = paramItemSchema(shape.Item)
	}
	return schema
}

// FromResponseShape translates a provider-neutral response shape into an
// OpenAPI schema.
func FromResponseShape(shape responsepkg.ShapeSpec) spec.Schema {
	schema := responseTypeSchema(shape.Type)
	if shape.Format != "" {
		schema.Format = shape.Format
	}
	switch shape.Type {
	case responsepkg.TypeObject:
		schema.Properties = responseProperties(shape.Attributes)
		schema.Required = responseRequired(shape.Attributes)
	case responsepkg.TypeArray:
		schema.Items = responseItemSchema(shape.Item)
	}
	return schema
}

func paramTypeSchema(typ parampkg.Type) spec.Schema {
	switch typ {
	case parampkg.TypeString:
		return spec.Schema{Type: "string"}
	case parampkg.TypeInt:
		return spec.Schema{Type: "integer", Format: "int32"}
	case parampkg.TypeInt64:
		return spec.Schema{Type: "integer", Format: "int64"}
	case parampkg.TypeFloat64:
		return spec.Schema{Type: "number", Format: "double"}
	case parampkg.TypeBool:
		return spec.Schema{Type: "boolean"}
	case parampkg.TypeObject:
		return spec.Schema{Type: "object"}
	case parampkg.TypeArray:
		return spec.Schema{Type: "array"}
	default:
		return spec.Schema{}
	}
}

func responseTypeSchema(typ responsepkg.Type) spec.Schema {
	switch typ {
	case responsepkg.TypeString:
		return spec.Schema{Type: "string"}
	case responsepkg.TypeInt:
		return spec.Schema{Type: "integer", Format: "int32"}
	case responsepkg.TypeFloat64:
		return spec.Schema{Type: "number", Format: "double"}
	case responsepkg.TypeBool:
		return spec.Schema{Type: "boolean"}
	case responsepkg.TypeObject:
		return spec.Schema{Type: "object"}
	case responsepkg.TypeArray:
		return spec.Schema{Type: "array"}
	default:
		return spec.Schema{}
	}
}

func paramProperties(parameters []parampkg.ParameterSpec) map[string]spec.Schema {
	if len(parameters) == 0 {
		return nil
	}

	properties := make(map[string]spec.Schema, len(parameters))
	for _, parameter := range parameters {
		properties[parameter.Name] = FromParamShape(parameter.Shape)
	}
	return properties
}

func paramRequired(parameters []parampkg.ParameterSpec) []string {
	required := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		if parameter.Required {
			required = append(required, parameter.Name)
		}
	}
	if len(required) == 0 {
		return nil
	}
	return required
}

func paramItemSchema(shape *parampkg.ShapeSpec) *spec.Schema {
	if shape == nil {
		schema := spec.Schema{}
		return &schema
	}
	schema := FromParamShape(*shape)
	return &schema
}

func responseProperties(attributes []responsepkg.AttributeSpec) map[string]spec.Schema {
	if len(attributes) == 0 {
		return nil
	}

	properties := make(map[string]spec.Schema, len(attributes))
	for _, attribute := range attributes {
		properties[attribute.Name] = FromResponseShape(attribute.Shape)
	}
	return properties
}

func responseRequired(attributes []responsepkg.AttributeSpec) []string {
	required := make([]string, 0, len(attributes))
	for _, attribute := range attributes {
		if attribute.Required {
			required = append(required, attribute.Name)
		}
	}
	if len(required) == 0 {
		return nil
	}
	return required
}

func responseItemSchema(shape *responsepkg.ShapeSpec) *spec.Schema {
	if shape == nil {
		schema := spec.Schema{}
		return &schema
	}
	schema := FromResponseShape(*shape)
	return &schema
}
