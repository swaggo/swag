package swag

import (
	"errors"
	"fmt"
	"github.com/go-openapi/spec"
	"go/ast"
	"strings"
)

// ErrFailedConvertPrimitiveType Failed to convert for swag to interpretable type
var ErrFailedConvertPrimitiveType = errors.New("swag property: failed convert primitive type")

// convertFromSpecificToPrimitive convert specific type to primitive
func convertFromSpecificToPrimitive(typeName string) (string, error) {
	name := typeName
	if strings.ContainsRune(name, '.') {
		name = strings.Split(name, ".")[1]
	}
	switch strings.ToUpper(name) {
	case "TIME", "OBJECTID", "UUID":
		return "string", nil
	case "DECIMAL":
		return "number", nil
	}
	return typeName, ErrFailedConvertPrimitiveType
}

// CheckSchemaType checks if typeName is not a name of primitive type
func CheckSchemaType(typeName string) error {
	if !IsPrimitiveType(typeName) {
		return fmt.Errorf("%s is not basic types", typeName)
	}
	return nil
}

// IsSimplePrimitiveType determine whether the type name is a simple primitive type
func IsSimplePrimitiveType(typeName string) bool {
	switch typeName {
	case "string", "number", "integer", "boolean":
		return true
	default:
		return false
	}
}

// IsPrimitiveType determine whether the type name is a primitive type
func IsPrimitiveType(typeName string) bool {
	switch typeName {
	case "string", "number", "integer", "boolean", "array", "object", "func":
		return true
	default:
		return false
	}
}

// IsNumericType determines whether the swagger type name is a numeric type
func IsNumericType(typeName string) bool {
	return typeName == "integer" || typeName == "number"
}

// TransToValidSchemeType indicates type will transfer golang basic type to swagger supported type.
func TransToValidSchemeType(typeName string) string {
	switch typeName {
	case "uint", "int", "uint8", "int8", "uint16", "int16", "byte":
		return "integer"
	case "uint32", "int32", "rune":
		return "integer"
	case "uint64", "int64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	default:
		return typeName // to support user defined types
	}
}

// IsGolangPrimitiveType determine whether the type name is a golang primitive type
func IsGolangPrimitiveType(typeName string) bool {
	switch typeName {
	case "uint",
		"int",
		"uint8",
		"int8",
		"uint16",
		"int16",
		"byte",
		"uint32",
		"int32",
		"rune",
		"uint64",
		"int64",
		"float32",
		"float64",
		"bool",
		"string":
		return true
	default:
		return false
	}
}

// TransToValidCollectionFormat determine valid collection format
func TransToValidCollectionFormat(format string) string {
	switch format {
	case "csv", "multi", "pipes", "tsv", "ssv":
		return format
	default:
		return ""
	}
}

// TypeDocName get alias from comment '// @name ', otherwise the original type name to display in doc
func TypeDocName(pkgName string, spec *ast.TypeSpec) string {
	if spec != nil {
		if spec.Comment != nil {
			for _, comment := range spec.Comment.List {
				text := strings.TrimSpace(comment.Text)
				text = strings.TrimLeft(text, "//")
				text = strings.TrimSpace(text)
				texts := strings.Split(text, " ")
				if len(texts) > 1 && strings.ToLower(texts[0]) == "@name" {
					return texts[1]
				}
			}
		}
		if spec.Name != nil {
			return fullTypeName(strings.Split(pkgName, ".")[0], spec.Name.Name)
		}
	}

	return pkgName
}

func PrimitiveSchema(typeName string) *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{Type: spec.StringOrArray{typeName}},
	}
}

func RefSchema(typeName string) *spec.Schema {
	return spec.RefSchema("#/definitions/" + typeName)
}

// BuildCustomSchema build custom schema specified by tag swaggertype
func BuildCustomSchema(types []string) (*spec.Schema, error) {
	if len(types) == 0 {
		return nil, nil
	}

	switch types[0] {
	case "primitive":
		if len(types) == 1 {
			return nil, errors.New("need primitive type after primitive")
		}
		return BuildCustomSchema(types[1:])
	case "array":
		if len(types) == 1 {
			return nil, errors.New("need array item type after array")
		}
		schema, err := BuildCustomSchema(types[1:])
		if err != nil {
			return nil, err
		}
		return spec.ArrayProperty(schema), nil
	default:
		err := CheckSchemaType(types[0])
		if err != nil {
			return nil, err
		}
		return PrimitiveSchema(types[0]), nil
	}
}
