package swag

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
)

func getStructFields(refType string, pkgDefs *TypeSpecDef) ([]StructFieldInfo, error) {
	files, err := os.ReadDir(pkgDefs.PkgPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filename := filepath.Join(pkgDefs.PkgPath, file.Name())

		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		fields, err := findStructFields(astFile, refType)
		if err != nil {
			return nil, err
		}

		if len(fields) == 0 {
			continue
		}

		return fields, nil
	}

	return nil, nil
}

func parseNonJsonStructTag(tag string) (*StructTagValue, error) {
	paramPattern := regexp.MustCompile(`(?:param|query):"([^"]+)"`)
	validatePattern := regexp.MustCompile(`validate:"([^"]+)"`)

	paramMatch := paramPattern.FindStringSubmatch(tag)
	validateMatch := validatePattern.FindStringSubmatch(tag)

	result := &StructTagValue{}

	if len(paramMatch) > 1 {
		result.ParamValue = paramMatch[1]
	}

	if len(validateMatch) > 1 {
		result.Validate = validateMatch[1]
	}

	return result, nil
}

func findStructFields(file *ast.File, refType string) ([]StructFieldInfo, error) {
	structName := refType
	var fields []StructFieldInfo

	ast.Inspect(file, func(n ast.Node) bool {
		typeDecl, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if typeDecl.Name.Name != structName {
			return true
		}

		structType, ok := typeDecl.Type.(*ast.StructType)
		if !ok {
			return true
		}

		for _, field := range structType.Fields.List {
			fieldInfo := StructFieldInfo{}

			if len(field.Names) > 0 {
				fieldInfo.Name = field.Names[0].Name
			}

			fieldInfo.Type = typeToString(field.Type)

			if field.Tag != nil {
				fieldInfo.Tag = field.Tag.Value
			}

			if field.Doc != nil {
				for _, comment := range field.Doc.List {
					fieldInfo.Comments = append(fieldInfo.Comments, comment.Text)
				}
			}

			fields = append(fields, fieldInfo)
		}

		return false
	})

	return fields, nil
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "" + typeToString(t.X)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}
