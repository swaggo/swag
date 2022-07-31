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

func (pkgDefs *PackagesDefinitions) parametrizeStruct(original *TypeSpecDef, fullGenericForm string) *TypeSpecDef {
	if spec, ok := genericsDefinitions[original][fullGenericForm]; ok {
		return spec
	}

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

		tdef := pkgDefs.FindTypeSpec(genericParam, original.File, true)
		if tdef == nil {
			return nil
		}

		genericParamTypeDefs[original.TypeSpec.TypeParams.List[i].Names[0].Name] = &genericTypeSpec{
			ArrayDepth: arrayDepth,
			TypeSpec:   tdef,
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

	var typeName = []string{TypeDocName(genericTypeName, parametrizedTypeSpec.TypeSpec)}

	for _, def := range original.TypeSpec.TypeParams.List {
		if specDef, ok := genericParamTypeDefs[def.Names[0].Name]; ok {
			var prefix = ""
			if specDef.ArrayDepth > 0 {
				prefix = "array_"
				if specDef.ArrayDepth > 1 {
					prefix = fmt.Sprintf("array%d_", specDef.ArrayDepth)
				}
			}
			typeName = append(typeName, prefix+strings.Replace(TypeDocName(specDef.TypeSpec.FullName(), specDef.TypeSpec.TypeSpec), "-", "_", -1))
		}
	}

	ident.Name = strings.Join(typeName, "-")
	ident.Name = strings.Replace(ident.Name, ".", "_", -1)

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
			return genTypeSpec.TypeSpec.TypeSpec.Type
		}
	} else if asArray, ok := expr.(*ast.ArrayType); ok {
		return &ast.ArrayType{Elt: resolveType(asArray.Elt, field, genericParamTypeDefs), Len: asArray.Len, Lbrack: asArray.Lbrack}
	}

	return field.Type
}
