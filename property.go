package swag

import (
	"go/ast"
	"log"
)

// getPropertyNameAndType returns the (string, string) value for the given field if it exists, otherwise it panics.
// reference: https://swagger.io/docs/specification/data-models/data-types/
// reference: https://golang.org/ref/spec#Numeric_types
// allowedValues: array, boolean, integer, null, number, object, string
func getPropertyNameAndType(field *ast.Field) (string, string) {
	var name string
	if astTypeSelectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {

		// Support for time.Time as a structure field
		if "Time" == astTypeSelectorExpr.Sel.Name {
			return "string", "date-time"
		}

		// Support bson.ObjectId type
		if "ObjectId" == astTypeSelectorExpr.Sel.Name {
			return "string", ""
		}

		panic("not supported 'astSelectorExpr' yet.")

	} else if astTypeIdent, ok := field.Type.(*ast.Ident); ok {
		name = astTypeIdent.Name

		// When its the int type will transfer to integer which is goswagger supported type
		switch name {
		case "uint", "int", "uint8", "int8", "uint16", "int16", "byte":
			return "integer", ""
		case "uint32", "int32", "rune":
			return "integer", "int32"
		case "uint64", "int64":
			return "integer", "int64"
		case "float32":
			return "number", "float"
		case "float64":
			return "number", "double"
		}

	} else if _, ok := field.Type.(*ast.StarExpr); ok {
		panic("not supported astStarExpr yet.")
	} else if _, ok := field.Type.(*ast.MapType); ok { // if map
		//TODO: support map
		return "object", ""
	} else if _, ok := field.Type.(*ast.ArrayType); ok { // if array
		return "array", ""
	} else if _, ok := field.Type.(*ast.StructType); ok { // if struct
		//TODO: support nested struct
		return "object", ""
	} else {
		log.Fatalf("Something goes wrong: %#v", field.Type)
	}

	return name, ""
}
