package swag

import "fmt"

// CheckSchemaType checks if typeName is not a name of primitive type
func CheckSchemaType(typeName string) error {
	if !IsPrimitiveType(typeName) {
		return fmt.Errorf("%s is not basic types", typeName)
	}
	return nil
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

// IsMineType determine whether the type name is a mine type
func IsMineType(typeName string) bool {
	switch typeName {
	case "json", "application/json", "xml", "text/xml", "plain", "text/plain", "html", "text/html",
		"mpfd", "multipart/form-data", "x-www-form-urlencoded", "application/x-www-form-urlencoded",
		"json-api", "application/vnd.api+json", "json-stream", "application/x-json-stream",
		"octet-stream", "application/octet-stream", "png", "image/png", "jpeg", "image/jpeg",
		"gif", "image/gif":
		return true
	default:
		return false
	}
}
