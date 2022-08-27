//go:build !go1.18
// +build !go1.18

package swag

import (
	"fmt"
	"github.com/go-openapi/spec"
	"go/ast"
)

func typeSpecFullName(typeSpecDef *TypeSpecDef) string {
	return typeSpecDef.FullName()
}

func (pkgDefs *PackagesDefinitions) parametrizeStruct(file *ast.File, original *TypeSpecDef, fullGenericForm string, parseDependency bool) *TypeSpecDef {
	return original
}

func getGenericFieldType(file *ast.File, field ast.Expr) (string, error) {
	return "", fmt.Errorf("unknown field type %#v", field)
}

func (parser *Parser) parseGenericTypeExpr(file *ast.File, typeExpr ast.Expr) (*spec.Schema, error) {
	switch typeExpr.(type) {
	// suppress debug messages for these types
	case *ast.InterfaceType:
	case *ast.StructType:
	case *ast.Ident:
	case *ast.StarExpr:
	case *ast.SelectorExpr:
	case *ast.ArrayType:
	case *ast.MapType:
	case *ast.FuncType:
	default:
		parser.debug.Printf("Type definition of type '%T' is not supported yet. Using 'object' instead.\n", typeExpr)
	}

	return PrimitiveSchema(OBJECT), nil
}
