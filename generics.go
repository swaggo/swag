//go:build go1.18
// +build go1.18

package swag

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"
)

var genericsDefinitions = map[*TypeSpecDef]map[string]*TypeSpecDef{}

type genericTypeSpec struct {
	ArrayDepth int
	TypeSpec   *TypeSpecDef
	Name       string
}

func (s *genericTypeSpec) Type() ast.Expr {
	if s.TypeSpec != nil {
		return &ast.SelectorExpr{
			X:   &ast.Ident{Name: ""},
			Sel: &ast.Ident{Name: s.Name},
		}
	}

	return &ast.Ident{Name: s.Name}
}

func (s *genericTypeSpec) TypeDocName() string {
	if s.TypeSpec != nil {
		return strings.Replace(TypeDocName(s.TypeSpec.FullName(), s.TypeSpec.TypeSpec), "-", "_", -1)
	}

	return s.Name
}

func typeSpecFullName(typeSpecDef *TypeSpecDef) string {
	fullName := typeSpecDef.FullName()

	if typeSpecDef.TypeSpec.TypeParams != nil {
		fullName = fullName + "["
		for i, typeParam := range typeSpecDef.TypeSpec.TypeParams.List {
			if i > 0 {
				fullName = fullName + "-"
			}

			fullName = fullName + typeParam.Names[0].Name
		}
		fullName = fullName + "]"
	}

	return fullName
}

func (pkgDefs *PackagesDefinitions) parametrizeStruct(original *TypeSpecDef, fullGenericForm string, parseDependency bool) *TypeSpecDef {
	if spec, ok := genericsDefinitions[original][fullGenericForm]; ok {
		return spec
	}

	pkgName := strings.Split(fullGenericForm, ".")[0]
	genericTypeName, genericParams := splitStructName(fullGenericForm)
	if genericParams == nil {
		return nil
	}

	genericParamTypeDefs := map[string]*genericTypeSpec{}
	if len(genericParams) != len(original.TypeSpec.TypeParams.List) {
		return nil
	}

	for i, genericParam := range genericParams {
		arrayDepth := 0
		for {
			if len(genericParam) <= 2 || genericParam[:2] != "[]" {
				break
			}
			genericParam = genericParam[2:]
			arrayDepth++
		}

		tdef := pkgDefs.FindTypeSpec(genericParam, original.File, parseDependency)
		genericParamTypeDefs[original.TypeSpec.TypeParams.List[i].Names[0].Name] = &genericTypeSpec{
			ArrayDepth: arrayDepth,
			TypeSpec:   tdef,
			Name:       genericParam,
		}
	}

	parametrizedTypeSpec := &TypeSpecDef{
		File:    original.File,
		PkgPath: original.PkgPath,
		TypeSpec: &ast.TypeSpec{
			Doc:     original.TypeSpec.Doc,
			Comment: original.TypeSpec.Comment,
			Assign:  original.TypeSpec.Assign,
		},
	}

	ident := &ast.Ident{
		NamePos: original.TypeSpec.Name.NamePos,
		Obj:     original.TypeSpec.Name.Obj,
	}

	if strings.Contains(genericTypeName, ".") {
		genericTypeName = strings.Split(genericTypeName, ".")[1]
	}

	var typeName = []string{TypeDocName(fullTypeName(pkgName, genericTypeName), parametrizedTypeSpec.TypeSpec)}

	for _, def := range original.TypeSpec.TypeParams.List {
		if specDef, ok := genericParamTypeDefs[def.Names[0].Name]; ok {
			var prefix = ""
			if specDef.ArrayDepth > 0 {
				prefix = "array_"
				if specDef.ArrayDepth > 1 {
					prefix = fmt.Sprintf("array%d_", specDef.ArrayDepth)
				}
			}
			typeName = append(typeName, prefix+specDef.TypeDocName())
		}
	}

	ident.Name = strings.Join(typeName, "-")
	ident.Name = strings.Replace(ident.Name, ".", "_", -1)
	pkgNamePrefix := pkgName + "_"
	if strings.HasPrefix(ident.Name, pkgNamePrefix) {
		ident.Name = fullTypeName(pkgName, ident.Name[len(pkgNamePrefix):])
	}
	ident.Name = string(IgnoreNameOverridePrefix) + ident.Name

	parametrizedTypeSpec.TypeSpec.Name = ident
	origStructType := original.TypeSpec.Type.(*ast.StructType)

	newStructTypeDef := &ast.StructType{
		Struct:     origStructType.Struct,
		Incomplete: origStructType.Incomplete,
		Fields: &ast.FieldList{
			Opening: origStructType.Fields.Opening,
			Closing: origStructType.Fields.Closing,
		},
	}

	for _, field := range origStructType.Fields.List {
		newField := &ast.Field{
			Doc:     field.Doc,
			Names:   field.Names,
			Tag:     field.Tag,
			Comment: field.Comment,
		}

		newField.Type = resolveType(field.Type, field, genericParamTypeDefs)

		newStructTypeDef.Fields.List = append(newStructTypeDef.Fields.List, newField)
	}

	parametrizedTypeSpec.TypeSpec.Type = newStructTypeDef
	if genericsDefinitions[original] == nil {
		genericsDefinitions[original] = map[string]*TypeSpecDef{}
	}
	genericsDefinitions[original][fullGenericForm] = parametrizedTypeSpec
	return parametrizedTypeSpec
}

// splitStructName splits a generic struct name in his parts
func splitStructName(fullGenericForm string) (string, []string) {
	// split only at the first '[' and remove the last ']'
	if fullGenericForm[len(fullGenericForm)-1] != ']' {
		return "", nil
	}

	genericParams := strings.SplitN(strings.TrimSpace(fullGenericForm)[:len(fullGenericForm)-1], "[", 2)
	if len(genericParams) == 1 {
		return "", nil
	}

	// generic type name
	genericTypeName := genericParams[0]

	// generic params
	insideBrackets := 0
	lastParam := ""
	params := strings.Split(genericParams[1], ",")
	genericParams = []string{}
	for _, p := range params {
		numOpened := strings.Count(p, "[")
		numClosed := strings.Count(p, "]")
		if numOpened == numClosed && insideBrackets == 0 {
			genericParams = append(genericParams, strings.TrimSpace(p))
			continue
		}

		insideBrackets += numOpened - numClosed
		lastParam += p + ","

		if insideBrackets == 0 {
			genericParams = append(genericParams, strings.TrimSpace(strings.TrimRight(lastParam, ",")))
			lastParam = ""
		}
	}

	return genericTypeName, genericParams
}

func resolveType(expr ast.Expr, field *ast.Field, genericParamTypeDefs map[string]*genericTypeSpec) ast.Expr {
	switch astExpr := expr.(type) {
	case *ast.Ident:
		if genTypeSpec, ok := genericParamTypeDefs[astExpr.Name]; ok {
			if genTypeSpec.ArrayDepth > 0 {
				genTypeSpec.ArrayDepth--
				return &ast.ArrayType{Elt: resolveType(expr, field, genericParamTypeDefs)}
			}
			return genTypeSpec.Type()
		}
	case *ast.ArrayType:
		return &ast.ArrayType{
			Elt:    resolveType(astExpr.Elt, field, genericParamTypeDefs),
			Len:    astExpr.Len,
			Lbrack: astExpr.Lbrack,
		}
	}

	return field.Type
}

func getGenericFieldType(file *ast.File, field ast.Expr) (string, error) {
	switch fieldType := field.(type) {
	case *ast.IndexListExpr:
		fullName, err := getGenericTypeName(file, fieldType.X)
		if err != nil {
			return "", err
		}
		fullName += "["

		for _, index := range fieldType.Indices {
			var fieldName string
			var err error

			switch item := index.(type) {
			case *ast.ArrayType:
				fieldName, err = getFieldType(file, item.Elt)
				fieldName = "[]" + fieldName
			default:
				fieldName, err = getFieldType(file, index)
			}

			if err != nil {
				return "", err
			}

			fullName += fieldName + ","
		}

		return strings.TrimRight(fullName, ",") + "]", nil
	case *ast.IndexExpr:
		x, err := getFieldType(file, fieldType.X)
		if err != nil {
			return "", err
		}

		i, err := getFieldType(file, fieldType.Index)
		if err != nil {
			return "", err
		}

		packageName := ""
		if !strings.Contains(x, ".") {
			if file.Name == nil {
				return "", errors.New("file name is nil")
			}
			packageName, _ = getFieldType(file, file.Name)
		}

		return strings.TrimLeft(fmt.Sprintf("%s.%s[%s]", packageName, x, i), "."), nil
	}

	return "", fmt.Errorf("unknown field type %#v", field)
}

func getGenericTypeName(file *ast.File, field ast.Expr) (string, error) {
	switch indexType := field.(type) {
	case *ast.Ident:
		spec := &TypeSpecDef{
			File:     file,
			TypeSpec: indexType.Obj.Decl.(*ast.TypeSpec),
			PkgPath:  file.Name.Name,
		}
		return spec.FullName(), nil
	case *ast.ArrayType:
		spec := &TypeSpecDef{
			File:     file,
			TypeSpec: indexType.Elt.(*ast.Ident).Obj.Decl.(*ast.TypeSpec),
			PkgPath:  file.Name.Name,
		}
		return spec.FullName(), nil
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", indexType.X.(*ast.Ident).Name, indexType.Sel.Name), nil
	}
	return "", fmt.Errorf("unknown type %#v", field)
}
