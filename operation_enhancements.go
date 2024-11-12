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

func (operation *Operation) getStructFields(refType string, astFile *ast.File) ([]StructFieldInfo, error) {
	pkgDefs := operation.parser.packages.FindTypeSpec(refType, astFile)
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

		fields, err := operation.findStructFields(astFile, refType)
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

type TagValue struct {
	Param    string
	Validate string
}

func (operation *Operation) parseStructTag(tag string) (*TagValue, error) {
	paramPattern := regexp.MustCompile(`(?:param|query):"([^"]+)"`)
	validatePattern := regexp.MustCompile(`validate:"([^"]+)"`)

	paramMatch := paramPattern.FindStringSubmatch(tag)
	validateMatch := validatePattern.FindStringSubmatch(tag)

	result := &TagValue{}

	if len(paramMatch) > 1 {
		result.Param = paramMatch[1]
	}

	if len(validateMatch) > 1 {
		result.Validate = validateMatch[1]
	}

	return result, nil
}

type StructFieldInfo struct {
	Name     string
	Type     string
	Tag      string
	Comments []string
}

func (operation *Operation) findStructFields(file *ast.File, refType string) ([]StructFieldInfo, error) {
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

			fieldInfo.Type = operation.typeExprToString(field.Type)

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

func (operation *Operation) typeExprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}
