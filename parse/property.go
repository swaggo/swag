package parse

import (
	"go/ast"
	"log"
)

func getPropertyName(field *ast.Field) string {
	var name string
	if _, ok := field.Type.(*ast.SelectorExpr); ok {
		panic("not supported 'astSelectorExpr' yet.")
	} else if astTypeIdent, ok := field.Type.(*ast.Ident); ok {
		name = astTypeIdent.Name
	} else if _, ok := field.Type.(*ast.StarExpr); ok {
		panic("not supported astStarExpr yet.")
	} else {
		log.Fatalf("Something goes wrong: %#v", field.Type)
	}

	return name
}
