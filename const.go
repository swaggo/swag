package swag

import (
	"go/ast"
)

// ConstVariable a model to record an consts variable
type ConstVariable struct {
	Name    *ast.Ident
	Type    ast.Expr
	Value   interface{}
	Comment *ast.CommentGroup
	File    *ast.File
}
