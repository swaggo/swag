package swag

import (
	"fmt"
	"go/ast"
	"strings"
)

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
