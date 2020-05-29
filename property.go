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
		if aliasedNames, ok := parser.ImportAliases[pkgName.Name]; ok {
			for aliasedName := range aliasedNames {
				if typeDefinitions, ok := parser.TypeDefinitions[aliasedName][astTypeSelectorExpr.Sel.Name]; ok {
					if expr, ok := typeDefinitions.Type.(*ast.SelectorExpr); ok {
						if primitiveType, err := convertFromSpecificToPrimitive(expr.Sel.Name); err == nil {
							return propertyNewFunc(primitiveType, "")
						}
					}
					parser.ParseDefinition(aliasedName, astTypeSelectorExpr.Sel.Name, typeDefinitions)
					return propertyNewFunc(astTypeSelectorExpr.Sel.Name, aliasedName)
				}
			}
		}
		name := fmt.Sprintf("%s.%v", pkgName, astTypeSelectorExpr.Sel.Name)
		if actualPrimitiveType, isCustomType := parser.CustomPrimitiveTypes[name]; isCustomType {
			return propertyName{SchemaType: actualPrimitiveType, ArrayType: actualPrimitiveType}
		}
	}
	return propertyName{SchemaType: "string", ArrayType: "string"}
}

// getPropertyName returns the string value for the given field if it exists
// allowedValues: array, boolean, integer, null, number, object, string
func getPropertyName(pkgName string, expr ast.Expr, parser *Parser) (propertyName, error) {
	switch tp := expr.(type) {
	case *ast.SelectorExpr:
		return parseFieldSelectorExpr(tp, parser, newProperty), nil
	case *ast.StarExpr:
		return getPropertyName(pkgName, tp.X, parser)
	case *ast.ArrayType:
		return getArrayPropertyName(pkgName, tp.Elt, parser), nil
	case *ast.MapType, *ast.StructType, *ast.InterfaceType:
		return propertyName{SchemaType: "object", ArrayType: "object"}, nil
	case *ast.FuncType:
		return propertyName{SchemaType: "func", ArrayType: ""}, nil
	case *ast.Ident:
		name := tp.Name
		// check if it is a custom type
		if actualPrimitiveType, isCustomType := parser.CustomPrimitiveTypes[fullTypeName(pkgName, name)]; isCustomType {
			return propertyName{SchemaType: actualPrimitiveType, ArrayType: actualPrimitiveType}, nil
		}

		name = TransToValidSchemeType(name)
		return propertyName{SchemaType: name, ArrayType: name}, nil
	default:
		return propertyName{}, errors.New("not supported" + fmt.Sprint(expr))
	}
}

func getArrayPropertyName(pkgName string, astTypeArrayElt ast.Expr, parser *Parser) propertyName {
	switch elt := astTypeArrayElt.(type) {
	case *ast.StructType, *ast.MapType, *ast.InterfaceType:
		return propertyName{SchemaType: "array", ArrayType: "object"}
	case *ast.ArrayType:
		return propertyName{SchemaType: "array", ArrayType: "array"}
	case *ast.StarExpr:
		return getArrayPropertyName(pkgName, elt.X, parser)
	case *ast.SelectorExpr:
		return parseFieldSelectorExpr(elt, parser, newArrayProperty)
	case *ast.Ident:
		name := elt.Name
		if actualPrimitiveType, isCustomType := parser.CustomPrimitiveTypes[fullTypeName(pkgName, name)]; isCustomType {
			name = actualPrimitiveType
		} else {
			name = TransToValidSchemeType(elt.Name)
		}
		return propertyName{SchemaType: "array", ArrayType: name}
	default:
		name := fmt.Sprintf("%s", astTypeArrayElt)
		if actualPrimitiveType, isCustomType := parser.CustomPrimitiveTypes[fullTypeName(pkgName, name)]; isCustomType {
			name = actualPrimitiveType
		} else {
			name = TransToValidSchemeType(name)
		}
		return propertyName{SchemaType: "array", ArrayType: name}
	}
}
