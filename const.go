package swag

import (
	"go/ast"
)

// ConstVariable a model to record a const variable
type ConstVariable struct {
	Name    *ast.Ident
	Type    ast.Expr
	Value   interface{}
	Comment *ast.CommentGroup
	File    *ast.File
	Pkg     *PackageDefinitions
}
