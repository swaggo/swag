package swag

import (
	"errors"

	"github.com/sv-tools/openapi/spec"
)

// PrimitiveSchemaV3 build a primitive schema.
func PrimitiveSchemaV3(refType string) *spec.RefOrSpec[spec.Schema] {
	result := spec.NewSchemaSpec()
	result.Spec.Type = &spec.SingleOrArray[string]{refType}

	return result
}

// IsComplexSchemaV3 whether a schema is complex and should be a ref schema
func IsComplexSchemaV3(schema *SchemaV3) bool {
	// a enum type should be complex
	if len(schema.Enum) > 0 {
		return true
	}

	// a schema without type (i.e. `any`) cannot be complex
	if schema.Type == nil {
		return false
	}

	// a deep array type is complex, how to determine deep? here more than 2 ,for example: [][]object,[][][]int
	if len(*schema.Type) > 2 {
		return true
	}

	//Object included, such as Object or []Object
	for _, st := range *schema.Type {
		if st == OBJECT {
			return true
		}
	}
	return false
}

// RefSchemaV3 build a reference schema.
func RefSchemaV3(refType string) *spec.RefOrSpec[spec.Schema] {
	return spec.NewRefOrSpec[spec.Schema](spec.NewRef("#/components/schemas/"+refType), nil)
}

// BuildCustomSchemaV3 build custom schema specified by tag swaggertype.
func BuildCustomSchemaV3(types []string) (*spec.RefOrSpec[spec.Schema], error) {
	if len(types) == 0 {
		return nil, nil
	}

	switch types[0] {
	case PRIMITIVE:
		if len(types) == 1 {
			return nil, errors.New("need primitive type after primitive")
		}

		return BuildCustomSchemaV3(types[1:])
	case ARRAY:
		if len(types) == 1 {
			return nil, errors.New("need array item type after array")
		}

		schema, err := BuildCustomSchemaV3(types[1:])
		if err != nil {
			return nil, err
		}

		// TODO: check if this is correct
		result := spec.NewSchemaSpec()
		result.Spec.Type = &spec.SingleOrArray[string]{ARRAY}
		result.Spec.AdditionalProperties = spec.NewBoolOrSchema(true, schema)

		return result, nil
	case OBJECT:
		if len(types) == 1 {
			return PrimitiveSchemaV3(types[0]), nil
		}

		schema, err := BuildCustomSchemaV3(types[1:])
		if err != nil {
			return nil, err
		}

		result := spec.NewSchemaSpec()
		result.Spec.AdditionalProperties = spec.NewBoolOrSchema(true, schema)
		result.Spec.Type = &spec.SingleOrArray[string]{OBJECT}

		return result, nil
	default:
		err := CheckSchemaType(types[0])
		if err != nil {
			return nil, err
		}

		return PrimitiveSchemaV3(types[0]), nil
	}
}

// TransToValidCollectionFormatV3 determine valid collection format.
func TransToValidCollectionFormatV3(format, in string) string {
	switch in {
	case "query":
		switch format {
		case "form", "spaceDelimited", "pipeDelimited", "deepObject":
			return format
		case "ssv":
			return "spaceDelimited"
		case "pipes":
			return "pipe"
		case "multi":
			return "form"
		case "csv":
			return "form"
		default:
			return ""
		}
	case "path":
		switch format {
		case "matrix", "label", "simple":
			return format
		case "csv":
			return "simple"
		default:
			return ""
		}
	case "header":
		switch format {
		case "form", "simple":
			return format
		case "csv":
			return "simple"
		default:
			return ""
		}
	case "cookie":
		switch format {
		case "form":
			return format
		}
	}

	return ""
}
