//go:build go1.18
// +build go1.18

package swag

import (
	"errors"
	"fmt"
	"github.com/go-openapi/spec"
	"go/ast"
	"strings"
	"sync"
	"unicode"
)

var genericDefinitionsMutex = &sync.RWMutex{}
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

func (pkgDefs *PackagesDefinitions) parametrizeGenericType(file *ast.File, original *TypeSpecDef, fullGenericForm string, parseDependency bool) *TypeSpecDef {
	genericDefinitionsMutex.RLock()
	tSpec, ok := genericsDefinitions[original][fullGenericForm]
	genericDefinitionsMutex.RUnlock()
	if ok {
		return tSpec
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

		tdef := pkgDefs.FindTypeSpec(genericParam, file, parseDependency)
		if tdef != nil && !strings.Contains(genericParam, ".") {
			genericParam = fullTypeName(file.Name.Name, genericParam)
		}

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

	newType := pkgDefs.resolveGenericType(original.File, original.TypeSpec.Type, genericParamTypeDefs, parseDependency)

	genericDefinitionsMutex.Lock()
	defer genericDefinitionsMutex.Unlock()
	parametrizedTypeSpec.TypeSpec.Type = newType
	if genericsDefinitions[original] == nil {
		genericsDefinitions[original] = map[string]*TypeSpecDef{}
	}
	genericsDefinitions[original][fullGenericForm] = parametrizedTypeSpec
	return parametrizedTypeSpec
}

// splitStructName splits a generic struct name in his parts
func splitStructName(fullGenericForm string) (string, []string) {
	//remove all spaces character
	fullGenericForm = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, fullGenericForm)

	// split only at the first '[' and remove the last ']'
	if fullGenericForm[len(fullGenericForm)-1] != ']' {
		return "", nil
	}

	genericParams := strings.SplitN(fullGenericForm[:len(fullGenericForm)-1], "[", 2)
	if len(genericParams) == 1 {
		return "", nil
	}

	// generic type name
	genericTypeName := genericParams[0]

	depth := 0
	genericParams = strings.FieldsFunc(genericParams[1], func(r rune) bool {
		if r == '[' {
			depth++
		} else if r == ']' {
			depth--
		} else if r == ',' && depth == 0 {
			return true
		}
		return false
	})
	if depth != 0 {
		return "", nil
	}

	return genericTypeName, genericParams
}

func (pkgDefs *PackagesDefinitions) resolveGenericType(file *ast.File, expr ast.Expr, genericParamTypeDefs map[string]*genericTypeSpec, parseDependency bool) ast.Expr {
	switch astExpr := expr.(type) {
	case *ast.Ident:
		if genTypeSpec, ok := genericParamTypeDefs[astExpr.Name]; ok {
			retType := genTypeSpec.Type()
			for i := 0; i < genTypeSpec.ArrayDepth; i++ {
				retType = &ast.ArrayType{Elt: retType}
			}
			return retType
		}
	case *ast.ArrayType:
		return &ast.ArrayType{
			Elt:    pkgDefs.resolveGenericType(file, astExpr.Elt, genericParamTypeDefs, parseDependency),
			Len:    astExpr.Len,
			Lbrack: astExpr.Lbrack,
		}
	case *ast.StarExpr:
		return &ast.StarExpr{
			Star: astExpr.Star,
			X:    pkgDefs.resolveGenericType(file, astExpr.X, genericParamTypeDefs, parseDependency),
		}
	case *ast.IndexExpr, *ast.IndexListExpr:
		fullGenericName, _ := getGenericFieldType(file, expr, genericParamTypeDefs)
		typeDef := pkgDefs.findGenericTypeSpec(fullGenericName, file, parseDependency)
		if typeDef != nil {
			return typeDef.TypeSpec.Type
		}
	case *ast.StructType:
		newStructTypeDef := &ast.StructType{
			Struct:     astExpr.Struct,
			Incomplete: astExpr.Incomplete,
			Fields: &ast.FieldList{
				Opening: astExpr.Fields.Opening,
				Closing: astExpr.Fields.Closing,
			},
		}

		for _, field := range astExpr.Fields.List {
			newField := &ast.Field{
				Type:    field.Type,
				Doc:     field.Doc,
				Names:   field.Names,
				Tag:     field.Tag,
				Comment: field.Comment,
			}

			newField.Type = pkgDefs.resolveGenericType(file, field.Type, genericParamTypeDefs, parseDependency)

			newStructTypeDef.Fields.List = append(newStructTypeDef.Fields.List, newField)
		}
		return newStructTypeDef
	}
	return expr
}

func getExtendedGenericFieldType(file *ast.File, field ast.Expr, genericParamTypeDefs map[string]*genericTypeSpec) (string, error) {
	switch fieldType := field.(type) {
	case *ast.ArrayType:
		fieldName, err := getExtendedGenericFieldType(file, fieldType.Elt, genericParamTypeDefs)
		return "[]" + fieldName, err
	case *ast.StarExpr:
		return getExtendedGenericFieldType(file, fieldType.X, genericParamTypeDefs)
	case *ast.Ident:
		if genericParamTypeDefs != nil {
			if typeSpec, ok := genericParamTypeDefs[fieldType.Name]; ok {
				return typeSpec.Name, nil
			}
		}
		if fieldType.Obj == nil {
			return fieldType.Name, nil
		}

		tSpec := &TypeSpecDef{
			File:     file,
			TypeSpec: fieldType.Obj.Decl.(*ast.TypeSpec),
			PkgPath:  file.Name.Name,
		}
		return tSpec.FullName(), nil
	default:
		return getFieldType(file, field)
	}
}

func getGenericFieldType(file *ast.File, field ast.Expr, genericParamTypeDefs map[string]*genericTypeSpec) (string, error) {
	var fullName string
	var baseName string
	var err error
	switch fieldType := field.(type) {
	case *ast.IndexListExpr:
		baseName, err = getGenericTypeName(file, fieldType.X)
		if err != nil {
			return "", err
		}
		fullName = baseName + "["

		for _, index := range fieldType.Indices {
			fieldName, err := getExtendedGenericFieldType(file, index, genericParamTypeDefs)
			if err != nil {
				return "", err
			}

			fullName += fieldName + ","
		}

		fullName = strings.TrimRight(fullName, ",") + "]"
	case *ast.IndexExpr:
		baseName, err = getGenericTypeName(file, fieldType.X)
		if err != nil {
			return "", err
		}

		indexName, err := getExtendedGenericFieldType(file, fieldType.Index, genericParamTypeDefs)
		if err != nil {
			return "", err
		}

		fullName = fmt.Sprintf("%s[%s]", baseName, indexName)
	}

	if fullName == "" {
		return "", fmt.Errorf("unknown field type %#v", field)
	}

	var packageName string
	if !strings.Contains(baseName, ".") {
		if file.Name == nil {
			return "", errors.New("file name is nil")
		}
		packageName, _ = getFieldType(file, file.Name)
	}

	return strings.TrimLeft(fmt.Sprintf("%s.%s", packageName, fullName), "."), nil
}

func getGenericTypeName(file *ast.File, field ast.Expr) (string, error) {
	switch fieldType := field.(type) {
	case *ast.Ident:
		if fieldType.Obj == nil {
			return fieldType.Name, nil
		}

		tSpec := &TypeSpecDef{
			File:     file,
			TypeSpec: fieldType.Obj.Decl.(*ast.TypeSpec),
			PkgPath:  file.Name.Name,
		}
		return tSpec.FullName(), nil
	case *ast.ArrayType:
		tSpec := &TypeSpecDef{
			File:     file,
			TypeSpec: fieldType.Elt.(*ast.Ident).Obj.Decl.(*ast.TypeSpec),
			PkgPath:  file.Name.Name,
		}
		return tSpec.FullName(), nil
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", fieldType.X.(*ast.Ident).Name, fieldType.Sel.Name), nil
	}
	return "", fmt.Errorf("unknown type %#v", field)
}

func (parser *Parser) parseGenericTypeExpr(file *ast.File, typeExpr ast.Expr) (*spec.Schema, error) {
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
	case *ast.IndexExpr:
		name, err := getExtendedGenericFieldType(file, expr, nil)
		if err == nil {
			if schema, err := parser.getTypeSchema(name, file, false); err == nil {
				return schema, nil
			}
		}

		parser.debug.Printf("Type definition of type '%T' is not supported yet. Using 'object' instead. (%s)\n", typeExpr, err)
	default:
		parser.debug.Printf("Type definition of type '%T' is not supported yet. Using 'object' instead.\n", typeExpr)
	}

	return PrimitiveSchema(OBJECT), nil
}
