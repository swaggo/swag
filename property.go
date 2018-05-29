package swag

import (
	"fmt"
	"go/ast"
	"strings"
)

type propertyName struct {
	SchemaType string
	ArrayType  string
}

func newSchemaPropertyName(schemaType string) propertyName {
	return propertyName{
		SchemaType: schemaType,
		ArrayType:  "",
	}
}

func newArraySchemaPropertyName(arrayType string) propertyName {
	return propertyName{
		SchemaType: "array",
		ArrayType:  arrayType,
	}
}

func parseFieldSelectorExpr(astTypeSelectorExpr *ast.SelectorExpr) propertyName {
	// TODO: In the future, add functions and make them solve for each user
	// Support for time.Time as a structure field
	if "Time" == astTypeSelectorExpr.Sel.Name {
		return newSchemaPropertyName("string")
	}

	// Support bson.ObjectId type
	if "ObjectId" == astTypeSelectorExpr.Sel.Name {
		return newSchemaPropertyName("string")
	}

	// Supprt UUID
	if "UUID" == strings.ToUpper(astTypeSelectorExpr.Sel.Name) {
		return newSchemaPropertyName("string")
	}

	// Supprt shopspring/decimal
	if "Decimal" == astTypeSelectorExpr.Sel.Name {
		return newSchemaPropertyName("number")
	}

	fmt.Printf("%s is not supported. but it will be set with string temporary. Please report any problems.", astTypeSelectorExpr.Sel.Name)
	return newSchemaPropertyName("string")
}

// getPropertyName returns the string value for the given field if it exists, otherwise it panics.
// allowedValues: array, boolean, integer, null, number, object, string
func getPropertyName(field *ast.Field) propertyName {
	if astTypeSelectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {
		return parseFieldSelectorExpr(astTypeSelectorExpr)
	}
	if astTypeIdent, ok := field.Type.(*ast.Ident); ok {
		name := astTypeIdent.Name
		schemeType := TransToValidSchemeType(name)
		return newSchemaPropertyName(schemeType)
	}
	if ptr, ok := field.Type.(*ast.StarExpr); ok {
		if astTypeSelectorExpr, ok := ptr.X.(*ast.SelectorExpr); ok {
			return parseFieldSelectorExpr(astTypeSelectorExpr)
		}
		if astTypeIdent, ok := ptr.X.(*ast.Ident); ok {
			name := astTypeIdent.Name
			schemeType := TransToValidSchemeType(name)
			return newSchemaPropertyName(schemeType)
		}
		if astTypeArray, ok := ptr.X.(*ast.ArrayType); ok { // if array
			if astTypeArrayIdent := astTypeArray.Elt.(*ast.Ident); ok {
				arrayType := astTypeArrayIdent.Name
				return newArraySchemaPropertyName(arrayType)
			}
		}
	}
	if astTypeArray, ok := field.Type.(*ast.ArrayType); ok { // if array
		if astTypeArrayExpr, ok := astTypeArray.Elt.(*ast.StarExpr); ok {
			if astTypeArrayIdent := astTypeArrayExpr.X.(*ast.Ident); ok {
				arrayType := astTypeArrayIdent.Name
				return newArraySchemaPropertyName(arrayType)
			}
		}
		arrayType := fmt.Sprintf("%s", astTypeArray.Elt)
		return newArraySchemaPropertyName(arrayType)
	}
	if _, ok := field.Type.(*ast.MapType); ok { // if map
		//TODO: support map
		return newSchemaPropertyName("object")
	}
	if _, ok := field.Type.(*ast.StructType); ok { // if struct
		return newSchemaPropertyName("object")
	}
	if _, ok := field.Type.(*ast.InterfaceType); ok { // if interface{}
		return newSchemaPropertyName("object")
	}
	panic("not supported" + fmt.Sprint(field.Type))
}
