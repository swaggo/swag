//go:build go1.18
// +build go1.18

package swag

import (
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
		return s.TypeSpec.TypeSpec.Type
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
			var isArray = len(genericParam) > 2 && genericParam[:2] == "[]"
			if isArray {
				genericParam = genericParam[2:]
				arrayDepth++
			} else {
				break
			}
		}

		tdef := pkgDefs.FindTypeSpec(genericParam, original.File, parseDependency)
		if tdef == nil {
			genericParamTypeDefs[original.TypeSpec.TypeParams.List[i].Names[0].Name] = &genericTypeSpec{
				ArrayDepth: arrayDepth,
				TypeSpec:   nil,
				Name:       genericParam,
			}
		} else {
			genericParamTypeDefs[original.TypeSpec.TypeParams.List[i].Names[0].Name] = &genericTypeSpec{
				ArrayDepth: arrayDepth,
				TypeSpec:   tdef,
			}
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
	ident.Name = string(IgnoreNameOverridePrefix) + strings.Replace(strings.Replace(ident.Name, ".", "_", -1), "_", ".", 1)

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
	if asIdent, ok := expr.(*ast.Ident); ok {
		if genTypeSpec, ok := genericParamTypeDefs[asIdent.Name]; ok {
			if genTypeSpec.ArrayDepth > 0 {
				genTypeSpec.ArrayDepth--
				return &ast.ArrayType{Elt: resolveType(expr, field, genericParamTypeDefs)}
			}
			return genTypeSpec.Type()
		}
	} else if asArray, ok := expr.(*ast.ArrayType); ok {
		return &ast.ArrayType{Elt: resolveType(asArray.Elt, field, genericParamTypeDefs), Len: asArray.Len, Lbrack: asArray.Lbrack}
	}

	return field.Type
}

func getGenericFieldType(file *ast.File, field ast.Expr) (string, error) {
	switch fieldType := field.(type) {
	case *ast.IndexListExpr:
		spec := &TypeSpecDef{
			File:     file,
			TypeSpec: getGenericTypeSpec(fieldType.X),
			PkgPath:  file.Name.Name,
		}
		fullName := spec.FullName() + "["

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

			fullName += fieldName + ", "
		}

		return strings.TrimRight(fullName, ", ") + "]", nil
	}

	return "", fmt.Errorf("unknown field type %#v", field)
}

func getGenericTypeSpec(field ast.Expr) *ast.TypeSpec {
	switch indexType := field.(type) {
	case *ast.Ident:
		return indexType.Obj.Decl.(*ast.TypeSpec)
	case *ast.ArrayType:
		return indexType.Elt.(*ast.Ident).Obj.Decl.(*ast.TypeSpec)
	}
	return nil
}
