//go:build !go1.18
// +build !go1.18

package swag

import (
	"fmt"
	"go/ast"
)

func typeSpecFullName(typeSpecDef *TypeSpecDef) string {
	return typeSpecDef.FullName()
}

func (pkgDefs *PackagesDefinitions) parametrizeStruct(original *TypeSpecDef, fullGenericForm string, parseDependency bool) *TypeSpecDef {
	return original
}

func getGenericFieldType(file *ast.File, field ast.Expr) (string, error) {
	return "", fmt.Errorf("unknown field type %#v", field)
}
