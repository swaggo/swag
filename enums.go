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

type EnumValue struct {
	key     string
	Value   interface{}
	Comment string
}

func (parser *Parser) parseConstEnumsFromFile(astFile *ast.File) {
	enums := make(map[string]map[string]interface{})
	for _, astDeclaration := range astFile.Decls {
		if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.CONST {
			var lastType, lastValueExpr ast.Expr
			for i, astSpec := range generalDeclaration.Specs {
				if valueSpec, ok := astSpec.(*ast.ValueSpec); ok {
					if valueSpec.Type != nil {
						lastType = valueSpec.Type
					}
					if len(valueSpec.Values) == 1 {
						lastValueExpr = valueSpec.Values[0]
					} else if len(valueSpec.Values) > 1 {
						lastValueExpr = nil
					}
					if ident, ok := lastType.(*ast.Ident); ok && !IsGolangPrimitiveType(ident.Name) {
						if enums[ident.Name] == nil {
							enums[ident.Name] = make(map[string]interface{})
						}

						if len(valueSpec.Values) == 0 {
							for j := 0; j < len(valueSpec.Names); j++ {
								enums[ident.Name][valueSpec.Names[j].Name] = evaluateEnumValue(i, "", lastValueExpr)
							}
						} else if len(valueSpec.Values) == len(valueSpec.Names) {
							for j := 0; j < len(valueSpec.Names); j++ {
								enums[ident.Name][valueSpec.Names[j].Name] = evaluateEnumValue(i, "", valueSpec.Values[j])
							}
						}
					}
				}
			}
		}
	}
}

func evaluateEnumValue(iota int, valueType string, expr ast.Expr) interface{} {
	if expr == nil {
		switch valueType {
		case "string":
			return ""
		case "int":
			return iota
		case "int32":
			return int32(iota)
		case "int64":
			return int64(iota)
		case "uint":
			return uint(iota)
		case "uint32":
			return uint32(iota)
		case "uint64":
			return uint64(iota)
		default:
			return iota
		}
	}
	return evaluateEnumValueFromExpr(iota, expr)
}

func evaluateEnumValueFromExpr(iota int, expr ast.Expr) interface{} {
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
		x := evaluateEnumValueFromExpr(iota, valueExpr.X)
		switch valueExpr.Op {
		case token.SUB:
			return -x.(int)
		case token.XOR:
			return ^(x.(int))
		}
	case *ast.BinaryExpr:
		x := evaluateEnumValueFromExpr(iota, valueExpr.X)
		y := evaluateEnumValueFromExpr(iota, valueExpr.Y)
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
	}
	return nil
}
