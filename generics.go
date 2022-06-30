//go:build go1.18
// +build go1.18

package swag

import (
	"go/ast"
	"strings"
)

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

func (pkgDefs *PackagesDefinitions) parametrizeStruct(original *TypeSpecDef, fullGenericForm string) *TypeSpecDef {
	genericParams := strings.Split(strings.TrimRight(fullGenericForm, "]"), "[")
	if len(genericParams) == 1 {
		return nil
	}

	genericParams = strings.Split(genericParams[1], ",")
	for i, p := range genericParams {
		genericParams[i] = strings.TrimSpace(p)
	}
	genericParamTypeDefs := map[string]*TypeSpecDef{}

	if len(genericParams) != len(original.TypeSpec.TypeParams.List) {
		return nil
	}

	for i, genericParam := range genericParams {
		tdef, ok := pkgDefs.uniqueDefinitions[genericParam]
		if !ok {
			return nil
		}

		genericParamTypeDefs[original.TypeSpec.TypeParams.List[i].Names[0].Name] = tdef
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

	genNameParts := strings.Split(fullGenericForm, "[")
	if strings.Contains(genNameParts[0], ".") {
		genNameParts[0] = strings.Split(genNameParts[0], ".")[1]
	}

	ident.Name = genNameParts[0] + "-" + strings.Replace(strings.Join(genericParams, "-"), ".", "_", -1)
	ident.Name = strings.Replace(strings.Replace(ident.Name, "\t", "", -1), " ", "", -1)

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

	return parametrizedTypeSpec
}

func resolveType(expr ast.Expr, field *ast.Field, genericParamTypeDefs map[string]*TypeSpecDef) ast.Expr {
	if asIdent, ok := expr.(*ast.Ident); ok {
		if genTypeSpec, ok := genericParamTypeDefs[asIdent.Name]; ok {
			return genTypeSpec.TypeSpec.Type
		}
	} else if asArray, ok := expr.(*ast.ArrayType); ok {
		return &ast.ArrayType{Elt: resolveType(asArray.Elt, field, genericParamTypeDefs), Len: asArray.Len, Lbrack: asArray.Lbrack}
	}

	return field.Type
}
