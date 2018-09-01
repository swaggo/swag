package swag

import "fmt"

// CheckSchemaType TODO: NEEDS COMMENT INFO
func CheckSchemaType(typeName string) {
	if !IsPrimitiveType(typeName) {
		panic(fmt.Errorf("%s is not basic types", typeName))
	}
}

// IsPrimitiveType determine whether the type name is a primitive type
func IsPrimitiveType(typeName string) bool {
	switch typeName {
	case "string", "number", "integer", "boolean", "array", "object":
		return true
	default:
		return false
	}
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
