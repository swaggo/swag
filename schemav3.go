package swag

import "github.com/sv-tools/openapi/spec"

// PrimitiveSchemaV3 build a primitive schema.
func PrimitiveSchemaV3(refType string) *spec.Schema {
	return &spec.Schema{
		JsonSchema: spec.JsonSchema{
			JsonSchemaCore: spec.JsonSchemaCore{
				Type: spec.SingleOrArray[string]{
					refType,
				},
			},
		},
	}
}
