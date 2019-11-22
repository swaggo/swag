package swag

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"
)

// ErrFailedConvertPrimitiveType Failed to convert for swag to interpretable type
var ErrFailedConvertPrimitiveType = errors.New("swag property: failed convert primitive type")

type propertyName struct {
	SchemaType string
	ArrayType  string
	CrossPkg   string
}

type propertyNewFunc func(schemeType string, crossPkg string) propertyName

func newArrayProperty(schemeType string, crossPkg string) propertyName {
	return propertyName{
		SchemaType: "array",
		ArrayType:  schemeType,
		CrossPkg:   crossPkg,
	}
}

func newProperty(schemeType string, crossPkg string) propertyName {
	return propertyName{
		SchemaType: schemeType,
		ArrayType:  "string",
		CrossPkg:   crossPkg,
	}
}

func convertFromSpecificToPrimitive(typeName string) (string, error) {
	typeName = strings.ToUpper(typeName)
	switch typeName {
	case "TIME", "OBJECTID", "UUID":
		return "string", nil
	case "DECIMAL":
		return "number", nil
	}
	return "", ErrFailedConvertPrimitiveType
}

func parseFieldSelectorExpr(astTypeSelectorExpr *ast.SelectorExpr, parser *Parser, propertyNewFunc propertyNewFunc) propertyName {
	if primitiveType, err := convertFromSpecificToPrimitive(astTypeSelectorExpr.Sel.Name); err == nil {
		return propertyNewFunc(primitiveType, "")
	}

	if pkgName, ok := astTypeSelectorExpr.X.(*ast.Ident); ok {
		if typeDefinitions, ok := parser.TypeDefinitions[pkgName.Name][astTypeSelectorExpr.Sel.Name]; ok {
			if expr, ok := typeDefinitions.Type.(*ast.SelectorExpr); ok {
				if primitiveType, err := convertFromSpecificToPrimitive(expr.Sel.Name); err == nil {
					return propertyNewFunc(primitiveType, "")
				}
			}
			parser.ParseDefinition(pkgName.Name, astTypeSelectorExpr.Sel.Name, typeDefinitions)
			return propertyNewFunc(astTypeSelectorExpr.Sel.Name, pkgName.Name)
		}
		if actualPrimitiveType, isCustomType := parser.CustomPrimitiveTypes[astTypeSelectorExpr.Sel.Name]; isCustomType {
			return propertyName{SchemaType: actualPrimitiveType, ArrayType: actualPrimitiveType}
		}
	}
	return propertyName{SchemaType: "string", ArrayType: "string"}
}

// getPropertyName returns the string value for the given field if it exists
// allowedValues: array, boolean, integer, null, number, object, string
func getPropertyName(expr ast.Expr, parser *Parser) (propertyName, error) {
	if astTypeSelectorExpr, ok := expr.(*ast.SelectorExpr); ok {
		return parseFieldSelectorExpr(astTypeSelectorExpr, parser, newProperty), nil
	}

	// check if it is a custom type
	typeName := fmt.Sprintf("%v", expr)
	if actualPrimitiveType, isCustomType := parser.CustomPrimitiveTypes[typeName]; isCustomType {
		return propertyName{SchemaType: actualPrimitiveType, ArrayType: actualPrimitiveType}, nil
	}

	if astTypeIdent, ok := expr.(*ast.Ident); ok {
		name := astTypeIdent.Name
		schemeType := TransToValidSchemeType(name)
		return propertyName{SchemaType: schemeType, ArrayType: schemeType}, nil
	}

	if ptr, ok := expr.(*ast.StarExpr); ok {
		return getPropertyName(ptr.X, parser)
	}

	if astTypeArray, ok := expr.(*ast.ArrayType); ok { // if array
		if _, ok := astTypeArray.Elt.(*ast.StructType); ok {
			return propertyName{SchemaType: "array", ArrayType: "object"}, nil
		}
		return getArrayPropertyName(astTypeArray, parser), nil
	}

	if _, ok := expr.(*ast.MapType); ok { // if map
		return propertyName{SchemaType: "object", ArrayType: "object"}, nil
	}

	if _, ok := expr.(*ast.StructType); ok { // if struct
		return propertyName{SchemaType: "object", ArrayType: "object"}, nil
	}

	if _, ok := expr.(*ast.InterfaceType); ok { // if interface{}
		return propertyName{SchemaType: "object", ArrayType: "object"}, nil
	}
	return propertyName{}, errors.New("not supported" + fmt.Sprint(expr))
}

func getArrayPropertyName(astTypeArray *ast.ArrayType, parser *Parser) propertyName {
	if astTypeArrayExpr, ok := astTypeArray.Elt.(*ast.SelectorExpr); ok {
		return parseFieldSelectorExpr(astTypeArrayExpr, parser, newArrayProperty)
	}
	if astTypeArrayExpr, ok := astTypeArray.Elt.(*ast.StarExpr); ok {
		if astTypeArraySel, ok := astTypeArrayExpr.X.(*ast.SelectorExpr); ok {
			return parseFieldSelectorExpr(astTypeArraySel, parser, newArrayProperty)
		}
		if astTypeArrayIdent, ok := astTypeArrayExpr.X.(*ast.Ident); ok {
			name := TransToValidSchemeType(astTypeArrayIdent.Name)
			return propertyName{SchemaType: "array", ArrayType: name}
		}
	}
	itemTypeName := TransToValidSchemeType(fmt.Sprintf("%s", astTypeArray.Elt))
	if actualPrimitiveType, isCustomType := parser.CustomPrimitiveTypes[itemTypeName]; isCustomType {
		itemTypeName = actualPrimitiveType
	}
	return propertyName{SchemaType: "array", ArrayType: itemTypeName}
}
