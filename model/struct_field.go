package model

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/swaggo/swag/console"
)

type StructField struct {
	Name       string         `json:"name"`
	Type       types.Type     `json:"type"`
	TypeString string         `json:"type_string"` // For easier JSON serialization
	Tag        string         `json:"tag"`
	Fields     []*StructField `json:"fields"` // For nested structs
}

func (this *StructField) IsPublic() bool {
	_, ok := this.GetTags()["public"]
	return ok
}

func (this *StructField) GetTags() map[string]string {
	tags := strings.Split(this.Tag, " ")
	result := make(map[string]string)
	for _, tag := range tags {
		parts := strings.SplitN(tag, ":", 2)
		if len(parts) == 2 {
			key := strings.Trim(parts[0], "`")
			value := strings.Trim(parts[1], "`")
			result[key] = strings.Trim(value, "\"")
		}
	}
	return result
}

// ToSpecSchema converts a StructField to OpenAPI spec.Schema
// propName: extracted from json tag (first part before comma)
// schema: the OpenAPI schema for this field
// required: true if omitempty is absent from json tag
// nestedTypes: list of struct type names encountered for recursive definition generation
func (this *StructField) ToSpecSchema(public bool) (propName string, schema *spec.Schema, required bool, nestedTypes []string, err error) {
	// Filter field if public mode and field is not public
	if public && !this.IsPublic() {
		return "", nil, false, nil, nil
	}

	// Check for swaggerignore tag
	tags := this.GetTags()
	if swaggerIgnore, ok := tags["swaggerignore"]; ok && strings.EqualFold(swaggerIgnore, "true") {
		console.Printf("$Red{$Bold{Ignoring field %s due to swaggerignore tag}}\n", this.Name)
		return "", nil, false, nil, nil
	}

	// Extract property name from json tag
	jsonTag := tags["json"]
	if jsonTag == "" {
		jsonTag = tags["column"]
	}
	if jsonTag == "" {
		// Skip fields without json or column tags - these are likely unexported
		// embedded fields that shouldn't be in the API schema
		return "", nil, false, nil, nil
	}

	parts := strings.Split(jsonTag, ",")
	propName = parts[0]

	// Check for omitempty to determine required
	required = true
	for _, part := range parts[1:] {
		if strings.TrimSpace(part) == "omitempty" {
			required = false
			break
		}
	}

	// Skip if json tag is "-"
	if propName == "-" {
		return "", nil, false, nil, nil
	}

	// Detect StructField[T] pattern and extract type parameter T
	typeStr := this.TypeString
	// Only use Type.String() if TypeString is empty or not set
	// This preserves manually set TypeString values (like "account.Properties")
	// instead of overriding with full path from Type.String()
	if this.Type != nil && this.TypeString == "" {
		typeStr = this.Type.String()
	}

	var extractedType string
	if strings.Contains(typeStr, "fields.StructField[") {
		// Extract type parameter using bracket parsing
		extractedType, err = extractTypeParameter(typeStr)
		if err != nil {
			return "", nil, false, nil, fmt.Errorf("failed to extract type parameter from %s: %w", typeStr, err)
		}
	} else {
		extractedType = typeStr
	}

	// Build schema for the extracted type
	schema, nestedTypes, err = buildSchemaForType(extractedType, public)
	if err != nil {
		return "", nil, false, nil, fmt.Errorf("failed to build schema for type %s: %w", extractedType, err)
	}

	return propName, schema, required, nestedTypes, nil
}

// extractTypeParameter extracts the type parameter T from StructField[T]
// Handles nested brackets like StructField[map[string][]User]
func extractTypeParameter(typeStr string) (string, error) {
	// Find the opening bracket for StructField[
	idx := strings.Index(typeStr, "StructField[")
	if idx == -1 {
		return "", fmt.Errorf("StructField[ not found in %s", typeStr)
	}

	// Start after "StructField["
	start := idx + len("StructField[")
	bracketCount := 1
	end := start

	// Count brackets to find matching closing bracket
	for end < len(typeStr) && bracketCount > 0 {
		switch typeStr[end] {
		case '[':
			bracketCount++
		case ']':
			bracketCount--
		}
		if bracketCount > 0 {
			end++
		}
	}

	if bracketCount != 0 {
		return "", fmt.Errorf("mismatched brackets in %s", typeStr)
	}

	extracted := typeStr[start:end]

	// Remove leading * if it's a pointer
	extracted = strings.TrimPrefix(extracted, "*")

	return extracted, nil
}

// buildSchemaForType builds an OpenAPI schema for a Go type string
// Returns schema, list of nested struct type names, and error
func buildSchemaForType(typeStr string, public bool) (*spec.Schema, []string, error) {
	var nestedTypes []string

	// Remove pointer prefix
	isPointer := strings.HasPrefix(typeStr, "*")
	if isPointer {
		typeStr = strings.TrimPrefix(typeStr, "*")
	}

	// Check if this is a fields wrapper type (StringField, IntField, etc.)
	// These should be treated as primitives, not struct types
	if isFieldsWrapperType(typeStr) {
		return getPrimitiveSchemaForFieldType(typeStr)
	}

	// Handle primitive types
	if isPrimitiveType(typeStr) {
		schema := primitiveTypeToSchema(typeStr)
		return schema, nil, nil
	}

	// Handle arrays
	if strings.HasPrefix(typeStr, "[]") {
		elemType := strings.TrimPrefix(typeStr, "[]")
		elemSchema, elemNestedTypes, err := buildSchemaForType(elemType, public)
		if err != nil {
			return nil, nil, err
		}
		schema := spec.ArrayProperty(elemSchema)
		return schema, elemNestedTypes, nil
	}

	// Handle maps
	if strings.HasPrefix(typeStr, "map[") {
		// Extract value type
		bracketCount := 0
		valueStart := -1
		for i, ch := range typeStr {
			if ch == '[' {
				bracketCount++
			} else if ch == ']' {
				bracketCount--
				if bracketCount == 0 {
					valueStart = i + 1
					break
				}
			}
		}
		if valueStart == -1 {
			return nil, nil, fmt.Errorf("invalid map type: %s", typeStr)
		}
		valueType := typeStr[valueStart:]
		valueSchema, valueNestedTypes, err := buildSchemaForType(valueType, public)
		if err != nil {
			return nil, nil, err
		}
		schema := spec.MapProperty(valueSchema)
		return schema, valueNestedTypes, nil
	}

	// Handle struct types (including package-qualified names)
	// Filter out "any" and "interface{}" types - these should be treated as generic objects
	if typeStr == "any" || typeStr == "interface{}" {
		// Return a generic object schema, don't add to nestedTypes
		return &spec.Schema{}, nil, nil
	}

	// Keep the full type name (including package prefix if present)
	// e.g., "account.Properties" should remain "account.Properties"
	typeName := typeStr

	// Add Public suffix if in public mode
	refName := typeName
	if public {
		refName = typeName + "Public"
	}

	// Create reference schema using the full type name
	schema := spec.RefSchema("#/definitions/" + refName)
	nestedTypes = append(nestedTypes, typeName)

	return schema, nestedTypes, nil
}

// isPrimitiveType checks if a type string is a Go primitive type
func isPrimitiveType(typeStr string) bool {
	primitives := map[string]bool{
		"string": true, "bool": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"byte": true, "rune": true,
		"float32": true, "float64": true,
		"time.Time": true, "*time.Time": true,
	}
	return primitives[typeStr]
}

// isFieldsWrapperType checks if a type is a fields package wrapper type
// like fields.StringField, fields.IntField, fields.StructField[T], etc.
func isFieldsWrapperType(typeStr string) bool {
	// Check for various field wrapper patterns
	return strings.Contains(typeStr, "fields.")
}

// getPrimitiveSchemaForFieldType returns the appropriate schema for a fields wrapper type
func getPrimitiveSchemaForFieldType(typeStr string) (*spec.Schema, []string, error) {
	if strings.Contains(typeStr, "fields.StringField") || strings.Contains(typeStr, "fields.StringConstantField") {
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"string"}}}, nil, nil
	}
	if strings.Contains(typeStr, "fields.IntField") || strings.Contains(typeStr, "fields.IntConstantField") || strings.Contains(typeStr, "fields.DecimalField") {
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"integer"}}}, nil, nil
	}
	if strings.Contains(typeStr, "fields.UUIDField") {
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"string"}, Format: "uuid"}}, nil, nil
	}
	if strings.Contains(typeStr, "fields.BoolField") {
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"boolean"}}}, nil, nil
	}
	if strings.Contains(typeStr, "fields.FloatField") {
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"number"}}}, nil, nil
	}
	if strings.Contains(typeStr, "fields.TimeField") {
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"string"}, Format: "date-time"}}, nil, nil
	}
	// Default to string for unknown field types
	return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"string"}}}, nil, nil
}

// primitiveTypeToSchema converts a Go primitive type to OpenAPI schema
func primitiveTypeToSchema(typeStr string) *spec.Schema {
	switch typeStr {
	case "string":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"string"}}}
	case "bool":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"boolean"}}}
	case "int", "uint":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"integer"}}}
	case "int8", "uint8", "int16", "uint16", "int32", "uint32", "byte", "rune":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"integer"}, Format: "int32"}}
	case "int64", "uint64":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"integer"}, Format: "int64"}}
	case "float32":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"number"}, Format: "float"}}
	case "float64":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"number"}, Format: "double"}}
	case "time.Time", "*time.Time":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"string"}, Format: "date-time"}}
	default:
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{typeStr}}}
	}
}
