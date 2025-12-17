package model

import (
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
)

type StructBuilder struct {
	Fields []*StructField `json:"fields"` // For nested structs
}

func (this *StructBuilder) BuildStructs(name string, public bool, aliasName string, childStructs map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", name))
	fmt.Printf("\n\nBuilding struct %s (public=%v) with %d fields\n", name, public, len(this.Fields))
	for _, field := range this.Fields {
		if public && !field.IsPublic() {
			continue
		}
		sb.WriteString(field.BuildStructDef(public))

		fmt.Printf("Field %s: IsStruct=%v, IsPublic=%v, TypeString=%s, FieldsCount=%d\n",
			field.Name, field.IsStruct(), field.IsPublic(), field.TypeString, len(field.Fields))

		if field.IsStruct() && public && field.IsPublic() {
			// Strip package prefix from TypeString for the key
			typeName := field.TypeString
			if strings.Contains(typeName, ".") {
				parts := strings.Split(typeName, ".")
				typeName = parts[len(parts)-1]
			}
			fmt.Printf("Creating child struct for %s -> %sPublic\n", field.TypeString, typeName)
			childStructs[typeName+"Public"] = field.BuildStruct(childStructs, public, typeName)
		}
	}
	sb.WriteString(fmt.Sprintf("}//@name %s\n", aliasName))

	// fmt.Printf("%s, %v | Sub Structs: %+v\n", name, public, childStructs)

	return sb.String()
}

func (this *StructBuilder) BuildInterface(name string, public bool, childStructs map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("export interface %s extends BaseModel {\n", name))
	fmt.Printf("\n\nBuilding struct %s (public=%v) with %d fields\n", name, public, len(this.Fields))
	for _, field := range this.Fields {
		if public && !field.IsPublic() {
			continue
		}
		sb.WriteString(field.BuildInterfaceDef(public))

		fmt.Printf("Field %s: IsStruct=%v, IsPublic=%v, TypeString=%s, FieldsCount=%d\n",
			field.Name, field.IsStruct(), field.IsPublic(), field.TypeString, len(field.Fields))

		if field.IsStruct() && public && field.IsPublic() {
			// Strip package prefix from TypeString for the key
			typeName := field.TypeString
			if strings.Contains(typeName, ".") {
				parts := strings.Split(typeName, ".")
				typeName = parts[len(parts)-1]
			}
			fmt.Printf("Creating child struct for %s -> %sPublic\n", field.TypeString, typeName)
			childStructs[typeName+"Public"] = field.BuildInterface(childStructs, public, typeName)
		}
	}
	sb.WriteString("}\n")

	// fmt.Printf("%s, %v | Sub Structs: %+v\n", name, public, childStructs)

	return sb.String()
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

	fmt.Printf("[BuildSpecSchema] Building schema for '%s' (public=%v) with %d fields\n", typeName, public, len(this.Fields))

	for _, field := range this.Fields {
		fmt.Printf("[BuildSpecSchema] Processing field: Name=%s, Type=%v, TypeString=%s\n", field.Name, field.Type, field.TypeString)
		propName, propSchema, isRequired, nestedTypes, err := field.ToSpecSchema(public)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build schema for field %s: %w", field.Name, err)
		}

		// Skip if field was filtered (e.g., not public when public=true)
		if propName == "" || propSchema == nil {
			fmt.Printf("[BuildSpecSchema] Field %s skipped (propName=%s, propSchema=%v)\n", field.Name, propName, propSchema)
			continue
		}
		fmt.Printf("[BuildSpecSchema] Field %s -> property %s (required=%v, nestedTypes=%v)\n", field.Name, propName, isRequired, nestedTypes)

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
