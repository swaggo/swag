package swag

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"

	"github.com/go-openapi/spec"
)

const (
	// ARRAY represent a array value.
	ARRAY = "array"
	// OBJECT represent a object value.
	OBJECT = "object"
	// PRIMITIVE represent a primitive value.
	PRIMITIVE = "primitive"
	// BOOLEAN represent a boolean value.
	BOOLEAN = "boolean"
	// INTEGER represent a integer value.
	INTEGER = "integer"
	// NUMBER represent a number value.
	NUMBER = "number"
	// STRING represent a string value.
	STRING = "string"
	// FUNC represent a function value.
	FUNC = "func"
	// ERROR represent a error value.
	ERROR = "error"
	// INTERFACE represent a interface value.
	INTERFACE = "interface{}"
	// ANY represent a any value.
	ANY = "any"
	// NIL represent a empty value.
	NIL = "nil"

	// IgnoreNameOverridePrefix Prepend to model to avoid renaming based on comment.
	IgnoreNameOverridePrefix = '$'
)

// CheckSchemaType checks if typeName is not a name of primitive type.
func CheckSchemaType(typeName string) error {
	if !IsPrimitiveType(typeName) {
		return fmt.Errorf("%s is not basic types", typeName)
	}

	return nil
}

// IsSimplePrimitiveType determine whether the type name is a simple primitive type.
func IsSimplePrimitiveType(typeName string) bool {
	switch typeName {
	case STRING, NUMBER, INTEGER, BOOLEAN:
		return true
	}

	return false
}

// IsPrimitiveType determine whether the type name is a primitive type.
func IsPrimitiveType(typeName string) bool {
	switch typeName {
	case STRING, NUMBER, INTEGER, BOOLEAN, ARRAY, OBJECT, FUNC:
		return true
	}

	return false
}

// IsInterfaceLike determines whether the swagger type name is an go named interface type like error type.
func IsInterfaceLike(typeName string) bool {
	return typeName == ERROR || typeName == ANY
}

// IsNumericType determines whether the swagger type name is a numeric type.
func IsNumericType(typeName string) bool {
	return typeName == INTEGER || typeName == NUMBER
}

// TransToValidSchemeType indicates type will transfer golang basic type to swagger supported type.
func TransToValidSchemeType(typeName string) string {
	switch typeName {
	case "uint", "int", "uint8", "int8", "uint16", "int16", "byte":
		return INTEGER
	case "uint32", "int32", "rune":
		return INTEGER
	case "uint64", "int64":
		return INTEGER
	case "float32", "float64":
		return NUMBER
	case "bool":
		return BOOLEAN
	case "string":
		return STRING
	}

	return typeName
}

// IsGolangPrimitiveType determine whether the type name is a golang primitive type.
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
	}

	return false
}

// TransToValidCollectionFormat determine valid collection format.
func TransToValidCollectionFormat(format string) string {
	switch format {
	case "csv", "multi", "pipes", "tsv", "ssv":
		return format
	}

	return ""
}

// TypeDocName get alias from comment '// @name ', otherwise the original type name to display in doc.
func TypeDocName(pkgName string, spec *ast.TypeSpec) string {
	if spec != nil && !ignoreNameOverride(pkgName) {
		if spec.Comment != nil {
			for _, comment := range spec.Comment.List {
				texts := strings.Split(strings.TrimSpace(strings.TrimLeft(comment.Text, "/")), " ")
				if len(texts) > 1 && strings.ToLower(texts[0]) == "@name" {
					return texts[1]
				}
			}
		}

		if spec.Name != nil {
			return fullTypeName(strings.Split(pkgName, ".")[0], spec.Name.Name)
		}
	}

	if ignoreNameOverride(pkgName) {
		return pkgName[1:]
	}

	return pkgName
}

func ignoreNameOverride(name string) bool {
	return len(name) != 0 && name[0] == IgnoreNameOverridePrefix
}

// TypeDocNameFuncScoped get alias from comment '// @name ', otherwise the original type name to display in doc.
func TypeDocNameFuncScoped(pkgName string, spec *ast.TypeSpec, fnName string) string {
	if spec != nil && !ignoreNameOverride(pkgName) {
		if spec.Comment != nil {
			for _, comment := range spec.Comment.List {
				texts := strings.Split(strings.TrimSpace(strings.TrimLeft(comment.Text, "/")), " ")
				if len(texts) > 1 && strings.ToLower(texts[0]) == "@name" {
					return texts[1]
				}
			}
		}

		if spec.Name != nil {
			return fullTypeNameFunctionScoped(strings.Split(pkgName, ".")[0], fnName, spec.Name.Name)
		}
	}

	if ignoreNameOverride(pkgName) {
		return pkgName[1:]
	}

	return pkgName
}

// RefSchema build a reference schema.
func RefSchema(refType string) *spec.Schema {
	return spec.RefSchema("#/definitions/" + refType)
}

// PrimitiveSchema build a primitive schema.
func PrimitiveSchema(refType string) *spec.Schema {
	return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{refType}}}
}

// BuildCustomSchema build custom schema specified by tag swaggertype.
func BuildCustomSchema(types []string) (*spec.Schema, error) {
	if len(types) == 0 {
		return nil, nil
	}

	switch types[0] {
	case PRIMITIVE:
		if len(types) == 1 {
			return nil, errors.New("need primitive type after primitive")
		}

		return BuildCustomSchema(types[1:])
	case ARRAY:
		if len(types) == 1 {
			return nil, errors.New("need array item type after array")
		}

		schema, err := BuildCustomSchema(types[1:])
		if err != nil {
			return nil, err
		}

		return spec.ArrayProperty(schema), nil
	case OBJECT:
		if len(types) == 1 {
			return PrimitiveSchema(types[0]), nil
		}

		schema, err := BuildCustomSchema(types[1:])
		if err != nil {
			return nil, err
		}

		return spec.MapProperty(schema), nil
	default:
		err := CheckSchemaType(types[0])
		if err != nil {
			return nil, err
		}

		return PrimitiveSchema(types[0]), nil
	}
}
