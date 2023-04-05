package swag

import (
	"go/ast"

	"github.com/sv-tools/openapi/spec"
)

func (p *Parser) parseGenericTypeExprV3(file *ast.File, typeExpr ast.Expr) (*spec.RefOrSpec[spec.Schema], error) {
	switch expr := typeExpr.(type) {
	// suppress debug messages for these types
	case *ast.InterfaceType:
	case *ast.StructType:
	case *ast.Ident:
	case *ast.StarExpr:
	case *ast.SelectorExpr:
	case *ast.ArrayType:
	case *ast.MapType:
	case *ast.FuncType:
	case *ast.IndexExpr, *ast.IndexListExpr:
		name, err := getExtendedGenericFieldType(file, expr, nil)
		if err == nil {
			if schema, err := p.getTypeSchemaV3(name, file, false); err == nil {
				return schema, nil
			}
		}

		p.debug.Printf("Type definition of type '%T' is not supported yet. Using 'object' instead. (%s)\n", typeExpr, err)
	default:
		p.debug.Printf("Type definition of type '%T' is not supported yet. Using 'object' instead.\n", typeExpr)
	}

	return PrimitiveSchemaV3(OBJECT), nil
}
