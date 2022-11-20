package swag

import (
	"go/ast"
	"go/token"
	"strconv"
)

// ConstVariable a model to record an const variable
type ConstVariable struct {
	Name    *ast.Ident
	Type    ast.Expr
	Value   interface{}
	Comment *ast.CommentGroup
}

// EvaluateValue evaluate the value
func (cv *ConstVariable) EvaluateValue(constTable map[string]*ConstVariable) interface{} {
	if expr, ok := cv.Value.(ast.Expr); ok {
		value, evalType := evaluateConstValue(cv.Name.Name, cv.Name.Obj.Data.(int), expr, constTable, make(map[string]struct{}))
		if cv.Type == nil && evalType != nil {
			cv.Type = evalType
		}
		if value != nil {
			cv.Value = value
		}
		return value
	}
	return cv.Value
}

func evaluateConstValue(name string, iota int, expr ast.Expr, constTable map[string]*ConstVariable, recursiveStack map[string]struct{}) (interface{}, ast.Expr) {
	if len(name) > 0 {
		if _, ok := recursiveStack[name]; ok {
			return nil, nil
		}
		recursiveStack[name] = struct{}{}
	}

	switch valueExpr := expr.(type) {
	case *ast.Ident:
		if valueExpr.Name == "iota" {
			return iota, nil
		}
		if constTable != nil {
			if cv, ok := constTable[valueExpr.Name]; ok {
				if expr, ok = cv.Value.(ast.Expr); ok {
					value, evalType := evaluateConstValue(valueExpr.Name, cv.Name.Obj.Data.(int), expr, constTable, recursiveStack)
					if cv.Type == nil {
						cv.Type = evalType
					}
					if value != nil {
						cv.Value = value
					}
					return value, evalType
				}
				return cv.Value, cv.Type
			}
		}
	case *ast.BasicLit:
		switch valueExpr.Kind {
		case token.INT:
			x, err := strconv.ParseInt(valueExpr.Value, 10, 64)
			if err != nil {
				return nil, nil
			}
			return int(x), nil
		case token.STRING, token.CHAR:
			return valueExpr.Value[1 : len(valueExpr.Value)-1], nil
		}
	case *ast.UnaryExpr:
		x, evalType := evaluateConstValue("", iota, valueExpr.X, constTable, recursiveStack)
		switch valueExpr.Op {
		case token.SUB:
			return -x.(int), evalType
		case token.XOR:
			return ^(x.(int)), evalType
		}
	case *ast.BinaryExpr:
		x, evalTypex := evaluateConstValue("", iota, valueExpr.X, constTable, recursiveStack)
		y, evalTypey := evaluateConstValue("", iota, valueExpr.Y, constTable, recursiveStack)
		evalType := evalTypex
		if evalType == nil {
			evalType = evalTypey
		}
		switch valueExpr.Op {
		case token.ADD:
			if ix, ok := x.(int); ok {
				return ix + y.(int), evalType
			} else if sx, ok := x.(string); ok {
				return sx + y.(string), evalType
			}
		case token.SUB:
			return x.(int) - y.(int), evalType
		case token.MUL:
			return x.(int) * y.(int), evalType
		case token.QUO:
			return x.(int) / y.(int), evalType
		case token.REM:
			return x.(int) % y.(int), evalType
		case token.AND:
			return x.(int) & y.(int), evalType
		case token.OR:
			return x.(int) | y.(int), evalType
		case token.XOR:
			return x.(int) ^ y.(int), evalType
		case token.SHL:
			return x.(int) << y.(int), evalType
		case token.SHR:
			return x.(int) >> y.(int), evalType
		}
	case *ast.ParenExpr:
		return evaluateConstValue("", iota, valueExpr.X, constTable, recursiveStack)
	case *ast.CallExpr:
		//data conversion
		if ident, ok := valueExpr.Fun.(*ast.Ident); ok && len(valueExpr.Args) == 1 && IsGolangPrimitiveType(ident.Name) {
			arg, _ := evaluateConstValue("", iota, valueExpr.Args[0], constTable, recursiveStack)
			return arg, nil
		}
	}
	return nil, nil
}
