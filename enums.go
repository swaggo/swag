package swag

import (
	"go/ast"
	"go/token"
	"strconv"
)

const (
	enumVarNamesExtension = "x-enum-varnames"
	enumCommentsExtension = "x-enum-comments"
)

// EnumValue a model to record an enum const variable
type EnumValue struct {
	key     string
	Value   interface{}
	Comment string
}

func evaluateEnumValue(iota int, expr ast.Expr) interface{} {
	switch valueExpr := expr.(type) {
	case *ast.Ident:
		if valueExpr.Name == "iota" {
			return iota
		}
	case *ast.BasicLit:
		switch valueExpr.Kind {
		case token.INT:
			x, err := strconv.ParseInt(valueExpr.Value, 10, 64)
			if err != nil {
				return nil
			}
			return int(x)
		case token.STRING:
			return valueExpr.Value[1 : len(valueExpr.Value)-1]
		}
	case *ast.UnaryExpr:
		x := evaluateEnumValue(iota, valueExpr.X)
		switch valueExpr.Op {
		case token.SUB:
			return -x.(int)
		case token.XOR:
			return ^(x.(int))
		}
	case *ast.BinaryExpr:
		x := evaluateEnumValue(iota, valueExpr.X)
		y := evaluateEnumValue(iota, valueExpr.Y)
		switch valueExpr.Op {
		case token.ADD:
			if ix, ok := x.(int); ok {
				return ix + y.(int)
			} else if sx, ok := x.(string); ok {
				return sx + y.(string)
			}
		case token.SUB:
			return x.(int) - y.(int)
		case token.MUL:
			return x.(int) * y.(int)
		case token.QUO:
			return x.(int) / y.(int)
		case token.REM:
			return x.(int) % y.(int)
		case token.AND:
			return x.(int) & y.(int)
		case token.OR:
			return x.(int) | y.(int)
		case token.XOR:
			return x.(int) & y.(int)
		case token.SHL:
			return x.(int) << y.(int)
		case token.SHR:
			return x.(int) >> y.(int)
		}
	case *ast.ParenExpr:
		return evaluateEnumValue(iota, valueExpr.X)
	}
	return nil
}
