//go:build go1.18
// +build go1.18

package swag

import (
	"encoding/json"
	"go/ast"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGenericsBasic(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/generics_basic"
	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	p := New()
	p.Overrides = map[string]string{
		"types.Field[string]":              "string",
		"types.DoubleField[string,string]": "string",
	}

	err = p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(b))
}

func TestParseGenericsArrays(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/generics_arrays"
	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	p := New()
	err = p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(b))
}

func TestParseGenericsNested(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/generics_nested"
	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	p := New()
	err = p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(b))
}

func TestParseGenericsProperty(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/generics_property"
	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	p := New()
	err = p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(b))
}

func TestParseGenericsNames(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/generics_names"
	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	p := New()
	err = p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(b))
}

func TestParametrizeStruct(t *testing.T) {
	pd := PackagesDefinitions{
		packages: make(map[string]*PackageDefinitions),
	}
	// valid
	typeSpec := pd.parametrizeStruct(&TypeSpecDef{
		TypeSpec: &ast.TypeSpec{
			Name:       &ast.Ident{Name: "Field"},
			TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
			Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
		}}, "test.Field[string, []string]", false)
	assert.Equal(t, "$test.Field-string-array_string", typeSpec.Name())

	// definition contains one type params, but two type params are provided
	typeSpec = pd.parametrizeStruct(&TypeSpecDef{
		TypeSpec: &ast.TypeSpec{
			Name:       &ast.Ident{Name: "Field"},
			TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}}},
			Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
		}}, "test.Field[string, string]", false)
	assert.Nil(t, typeSpec)

	// definition contains two type params, but only one is used
	typeSpec = pd.parametrizeStruct(&TypeSpecDef{
		TypeSpec: &ast.TypeSpec{
			Name:       &ast.Ident{Name: "Field"},
			TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
			Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
		}}, "test.Field[string]", false)
	assert.Nil(t, typeSpec)

	// name is not a valid type name
	typeSpec = pd.parametrizeStruct(&TypeSpecDef{
		TypeSpec: &ast.TypeSpec{
			Name:       &ast.Ident{Name: "Field"},
			TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
			Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
		}}, "test.Field[string", false)
	assert.Nil(t, typeSpec)

	typeSpec = pd.parametrizeStruct(&TypeSpecDef{
		TypeSpec: &ast.TypeSpec{
			Name:       &ast.Ident{Name: "Field"},
			TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
			Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
		}}, "test.Field[string, [string]", false)
	assert.Nil(t, typeSpec)

	typeSpec = pd.parametrizeStruct(&TypeSpecDef{
		TypeSpec: &ast.TypeSpec{
			Name:       &ast.Ident{Name: "Field"},
			TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
			Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
		}}, "test.Field[string, ]string]", false)
	assert.Nil(t, typeSpec)
}

func TestSplitStructNames(t *testing.T) {
	t.Parallel()

	field, params := splitStructName("test.Field")
	assert.Empty(t, field)
	assert.Nil(t, params)

	field, params = splitStructName("test.Field]")
	assert.Empty(t, field)
	assert.Nil(t, params)

	field, params = splitStructName("test.Field[string")
	assert.Empty(t, field)
	assert.Nil(t, params)

	field, params = splitStructName("test.Field[string]")
	assert.Equal(t, "test.Field", field)
	assert.Equal(t, []string{"string"}, params)

	field, params = splitStructName("test.Field[string, []string]")
	assert.Equal(t, "test.Field", field)
	assert.Equal(t, []string{"string", "[]string"}, params)
}

func TestGetGenericFieldType(t *testing.T) {
	field, err := getFieldType(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexListExpr{
			X:       &ast.Ident{Name: "types", Obj: &ast.Object{Decl: &ast.TypeSpec{Name: &ast.Ident{Name: "Field"}}}},
			Indices: []ast.Expr{&ast.Ident{Name: "string"}},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, "test.Field[string]", field)

	field, err = getFieldType(
		&ast.File{Name: &ast.Ident{}},
		&ast.IndexListExpr{
			X:       &ast.Ident{Name: "types", Obj: &ast.Object{Decl: &ast.TypeSpec{Name: &ast.Ident{Name: "Field"}}}},
			Indices: []ast.Expr{&ast.Ident{Name: "string"}},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, "Field[string]", field)

	field, err = getFieldType(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexListExpr{
			X:       &ast.Ident{Name: "types", Obj: &ast.Object{Decl: &ast.TypeSpec{Name: &ast.Ident{Name: "Field"}}}},
			Indices: []ast.Expr{&ast.Ident{Name: "string"}, &ast.Ident{Name: "int"}},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, "test.Field[string,int]", field)

	field, err = getFieldType(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexListExpr{
			X:       &ast.Ident{Name: "types", Obj: &ast.Object{Decl: &ast.TypeSpec{Name: &ast.Ident{Name: "Field"}}}},
			Indices: []ast.Expr{&ast.Ident{Name: "string"}, &ast.ArrayType{Elt: &ast.Ident{Name: "int"}}},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, "test.Field[string,[]int]", field)

	field, err = getFieldType(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexListExpr{
			X:       &ast.BadExpr{},
			Indices: []ast.Expr{&ast.Ident{Name: "string"}, &ast.Ident{Name: "int"}},
		},
	)
	assert.Error(t, err)

	field, err = getFieldType(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexListExpr{
			X:       &ast.Ident{Name: "types", Obj: &ast.Object{Decl: &ast.TypeSpec{Name: &ast.Ident{Name: "Field"}}}},
			Indices: []ast.Expr{&ast.Ident{Name: "string"}, &ast.ArrayType{Elt: &ast.BadExpr{}}},
		},
	)
	assert.Error(t, err)

	field, err = getFieldType(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexExpr{X: &ast.Ident{Name: "Field"}, Index: &ast.Ident{Name: "string"}},
	)
	assert.NoError(t, err)
	assert.Equal(t, "test.Field[string]", field)

	field, err = getFieldType(
		&ast.File{Name: nil},
		&ast.IndexExpr{X: &ast.Ident{Name: "Field"}, Index: &ast.Ident{Name: "string"}},
	)
	assert.Error(t, err)

	field, err = getFieldType(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexExpr{X: &ast.BadExpr{}, Index: &ast.Ident{Name: "string"}},
	)
	assert.Error(t, err)

	field, err = getFieldType(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexExpr{X: &ast.Ident{Name: "Field"}, Index: &ast.BadExpr{}},
	)
	assert.Error(t, err)

	field, err = getFieldType(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "field"}, Sel: &ast.Ident{Name: "Name"}}, Index: &ast.Ident{Name: "string"}},
	)
	assert.NoError(t, err)
	assert.Equal(t, "field.Name[string]", field)
}

func TestGetGenericTypeName(t *testing.T) {
	field, err := getGenericTypeName(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.Ident{Name: "types", Obj: &ast.Object{Decl: &ast.TypeSpec{Name: &ast.Ident{Name: "Field"}}}},
	)
	assert.NoError(t, err)
	assert.Equal(t, "test.Field", field)

	field, err = getGenericTypeName(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.ArrayType{Elt: &ast.Ident{Name: "types", Obj: &ast.Object{Decl: &ast.TypeSpec{Name: &ast.Ident{Name: "Field"}}}}},
	)
	assert.NoError(t, err)
	assert.Equal(t, "test.Field", field)

	field, err = getGenericTypeName(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.SelectorExpr{X: &ast.Ident{Name: "field"}, Sel: &ast.Ident{Name: "Name"}},
	)
	assert.NoError(t, err)
	assert.Equal(t, "field.Name", field)

	_, err = getGenericTypeName(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.BadExpr{},
	)
	assert.Error(t, err)
}
