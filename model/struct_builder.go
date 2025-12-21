package model

import (
	"fmt"

	"github.com/go-openapi/spec"
)

type StructBuilder struct {
	Fields []*StructField `json:"fields"` // For nested structs
}

// BuildSpecSchema builds an OpenAPI spec.Schema for the struct
// Returns the schema, a list of nested struct type names, and any error
func (this *StructBuilder) BuildSpecSchema(typeName string, public bool) (*spec.Schema, []string, error) {
	schema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       []string{"object"},
			Properties: make(map[string]spec.Schema),
		},
	}

	var required []string
	nestedStructs := make(map[string]bool) // Use map to deduplicate

	for _, field := range this.Fields {
		propName, propSchema, isRequired, nestedTypes, err := field.ToSpecSchema(public)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build schema for field %s: %w", field.Name, err)
		}

		// Skip if field was filtered (e.g., not public when public=true)
		if propName == "" || propSchema == nil {
			continue
		}

		// Add property to schema
		schema.Properties[propName] = *propSchema

		// Add to required list if needed
		if isRequired {
			required = append(required, propName)
		}

		// Collect nested struct types
		for _, nestedType := range nestedTypes {
			nestedStructs[nestedType] = true
		}
	}

	// Set required fields
	if len(required) > 0 {
		schema.Required = required
	}

	// Convert nested structs map to slice
	var nestedList []string
	for typeName := range nestedStructs {
		nestedList = append(nestedList, typeName)
	}

	return schema, nestedList, nil
}
