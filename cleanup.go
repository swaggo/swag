package swag

import (
	"github.com/go-openapi/spec"
)

// RemoveUnusedDefinitions removes schema definitions that are not referenced anywhere in the Swagger spec.
// This helps keep the generated documentation clean by eliminating schemas that were generated but never used.
func RemoveUnusedDefinitions(swagger *spec.Swagger) {
	if swagger == nil || swagger.Definitions == nil {
		return
	}

	// Collect all $ref references from the entire swagger spec
	used := make(map[string]bool)
	collectRefs(swagger, used)

	// Iteratively find transitive dependencies in definitions
	// Keep checking until no new dependencies are found
	changed := true
	for changed {
		changed = false
		for name := range swagger.Definitions {
			if used[name] {
				// Check this definition for nested refs
				schema := swagger.Definitions[name]
				beforeCount := len(used)
				collectSchemaRefs(&schema, used)
				if len(used) > beforeCount {
					changed = true
				}
			}
		}
	}

	// Remove definitions that are not referenced
	for name := range swagger.Definitions {
		if !used[name] {
			delete(swagger.Definitions, name)
		}
	}
}

// collectRefs recursively collects all $ref references in the swagger spec
func collectRefs(v interface{}, used map[string]bool) {
	switch val := v.(type) {
	case spec.Schema:
		collectSchemaRefs(&val, used)
	case *spec.Schema:
		collectSchemaRefs(val, used)
	case spec.Response:
		if val.Schema != nil {
			collectSchemaRefs(val.Schema, used)
		}
		for _, header := range val.Headers {
			if header.Items != nil {
				collectItemsRefs(header.Items, used)
			}
		}
	case *spec.Response:
		if val.Schema != nil {
			collectSchemaRefs(val.Schema, used)
		}
		for _, header := range val.Headers {
			if header.Items != nil {
				collectItemsRefs(header.Items, used)
			}
		}
	case spec.Parameter:
		if val.Schema != nil {
			collectSchemaRefs(val.Schema, used)
		}
		if val.Items != nil {
			collectItemsRefs(val.Items, used)
		}
	case *spec.Parameter:
		if val.Schema != nil {
			collectSchemaRefs(val.Schema, used)
		}
		if val.Items != nil {
			collectItemsRefs(val.Items, used)
		}
	case spec.Operation:
		for _, param := range val.Parameters {
			collectRefs(param, used)
		}
		if val.Responses != nil {
			for _, resp := range val.Responses.StatusCodeResponses {
				collectRefs(resp, used)
			}
			if val.Responses.Default != nil {
				collectRefs(*val.Responses.Default, used)
			}
		}
	case *spec.Operation:
		for _, param := range val.Parameters {
			collectRefs(param, used)
		}
		if val.Responses != nil {
			for _, resp := range val.Responses.StatusCodeResponses {
				collectRefs(resp, used)
			}
			if val.Responses.Default != nil {
				collectRefs(*val.Responses.Default, used)
			}
		}
	case spec.PathItem:
		if val.Get != nil {
			collectRefs(*val.Get, used)
		}
		if val.Put != nil {
			collectRefs(*val.Put, used)
		}
		if val.Post != nil {
			collectRefs(*val.Post, used)
		}
		if val.Delete != nil {
			collectRefs(*val.Delete, used)
		}
		if val.Options != nil {
			collectRefs(*val.Options, used)
		}
		if val.Head != nil {
			collectRefs(*val.Head, used)
		}
		if val.Patch != nil {
			collectRefs(*val.Patch, used)
		}
		for _, param := range val.Parameters {
			collectRefs(param, used)
		}
	case *spec.Swagger:
		// Collect from paths
		for _, pathItem := range val.Paths.Paths {
			collectRefs(pathItem, used)
		}
		// Collect from parameters
		for _, param := range val.Parameters {
			collectRefs(param, used)
		}
		// Collect from responses
		for _, resp := range val.Responses {
			collectRefs(resp, used)
		}
	case map[string]spec.Schema:
		for _, schema := range val {
			collectSchemaRefs(&schema, used)
		}
	}
}

// collectSchemaRefs collects $ref references from a schema
func collectSchemaRefs(schema *spec.Schema, used map[string]bool) {
	if schema == nil {
		return
	}

	// Check direct $ref
	if schema.Ref.String() != "" {
		refName := getRefName(schema.Ref.String())
		if refName != "" {
			used[refName] = true
		}
	}

	// Check items
	if schema.Items != nil {
		if schema.Items.Schema != nil {
			collectSchemaRefs(schema.Items.Schema, used)
		}
		for _, itemSchema := range schema.Items.Schemas {
			collectSchemaRefs(&itemSchema, used)
		}
	}

	// Check properties
	for _, prop := range schema.Properties {
		collectSchemaRefs(&prop, used)
	}

	// Check additional properties
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
		collectSchemaRefs(schema.AdditionalProperties.Schema, used)
	}

	// Check allOf, oneOf, anyOf
	for _, s := range schema.AllOf {
		collectSchemaRefs(&s, used)
	}
	for _, s := range schema.OneOf {
		collectSchemaRefs(&s, used)
	}
	for _, s := range schema.AnyOf {
		collectSchemaRefs(&s, used)
	}

	// Check not
	if schema.Not != nil {
		collectSchemaRefs(schema.Not, used)
	}

	// Check definitions within schema
	for _, def := range schema.Definitions {
		collectSchemaRefs(&def, used)
	}
}

// collectItemsRefs collects $ref references from items
func collectItemsRefs(items *spec.Items, used map[string]bool) {
	if items == nil {
		return
	}

	if items.Ref.String() != "" {
		refName := getRefName(items.Ref.String())
		if refName != "" {
			used[refName] = true
		}
	}

	if items.Items != nil {
		collectItemsRefs(items.Items, used)
	}
}

// getRefName extracts the definition name from a $ref string like "#/definitions/ModelName"
func getRefName(ref string) string {
	// Expected format: "#/definitions/ModelName"
	const prefix = "#/definitions/"
	if len(ref) > len(prefix) && ref[:len(prefix)] == prefix {
		return ref[len(prefix):]
	}
	return ""
}
