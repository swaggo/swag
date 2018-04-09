package swag

import (
	"fmt"
	"go/ast"
)

// getPropertyName returns the string value for the given field if it exists, otherwise it panics.
// allowedValues: array, boolean, integer, null, number, object, string
func getPropertyName(field *ast.Field) (name string, fieldType string) {
	if astTypeSelectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {

		// Support for time.Time as a structure field
		if "Time" == astTypeSelectorExpr.Sel.Name {
			return "string", "string"
		}

		// Support bson.ObjectId type
		if "ObjectId" == astTypeSelectorExpr.Sel.Name {
			return "string", "string"
		}

		panic("not supported 'astSelectorExpr' yet.")

	} else if astTypeIdent, ok := field.Type.(*ast.Ident); ok {
		name = astTypeIdent.Name

		// When its the int type will transfer to integer which is goswagger supported type
		schemeType := TransToValidSchemeType(name)
		return schemeType, schemeType

	} else if ptr, ok := field.Type.(*ast.StarExpr); ok {
		if astTypeSelectorExpr, ok := ptr.X.(*ast.SelectorExpr); ok {

			// Support for time.Time as a structure field
			if "Time" == astTypeSelectorExpr.Sel.Name {
				return "string", "string"
			}

			// Support bson.ObjectId type
			if "ObjectId" == astTypeSelectorExpr.Sel.Name {
				return "string", "string"
			}

			panic("not supported 'astSelectorExpr' yet.")

		} else if astTypeIdent, ok := ptr.X.(*ast.Ident); ok {
			name = astTypeIdent.Name
			schemeType := TransToValidSchemeType(name)
			return schemeType, schemeType
		}
	} else if _, ok := field.Type.(*ast.MapType); ok { // if map
		//TODO: support map
		return "object", "object"
	} else if astTypeArray, ok := field.Type.(*ast.ArrayType); ok { // if array
		str := fmt.Sprintf("%s", astTypeArray.Elt)
		return "array", str
	} else if _, ok := field.Type.(*ast.StructType); ok { // if struct
		return "object", "object"
	} else if _, ok := field.Type.(*ast.InterfaceType); ok { // if interface{}
		return "object", "object"
	}

	return name, fieldType
}
