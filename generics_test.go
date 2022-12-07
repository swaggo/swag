//go:build go1.18
// +build go1.18

package swag

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testLogger struct {
	Messages []string
}

func (t *testLogger) Printf(format string, v ...interface{}) {
	t.Messages = append(t.Messages, fmt.Sprintf(format, v...))
}

func TestParseGenericsBasic(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/generics_basic"
	expected, err := os.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	p := New()
	p.Overrides = map[string]string{
		"types.Field[string]":               "string",
		"types.DoubleField[string,string]":  "[]string",
		"types.TrippleField[string,string]": "[][]string",
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
	expected, err := os.ReadFile(filepath.Join(searchDir, "expected.json"))
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
	expected, err := os.ReadFile(filepath.Join(searchDir, "expected.json"))
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
	expected, err := os.ReadFile(filepath.Join(searchDir, "expected.json"))
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
	expected, err := os.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	p := New()
	err = p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(b))
}

func TestParseGenericsPackageAlias(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/generics_package_alias/internal"
	expected, err := os.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	p := New(SetParseDependency(true))
	err = p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(b))
}

func TestParametrizeStruct(t *testing.T) {
	pd := PackagesDefinitions{
		packages:          make(map[string]*PackageDefinitions),
		uniqueDefinitions: make(map[string]*TypeSpecDef),
	}
	// valid
	typeSpec := pd.parametrizeGenericType(
		&ast.File{Name: &ast.Ident{Name: "test2"}},
		&TypeSpecDef{
			File: &ast.File{Name: &ast.Ident{Name: "test"}},
			TypeSpec: &ast.TypeSpec{
				Name:       &ast.Ident{Name: "Field"},
				TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
				Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
			}}, "test.Field[string, []string]")
	assert.NotNil(t, typeSpec)
	assert.Equal(t, "$test.Field-string-array_string", typeSpec.Name())
	assert.Equal(t, "test.Field-string-array_string", typeSpec.TypeName())

	// definition contains one type params, but two type params are provided
	typeSpec = pd.parametrizeGenericType(
		&ast.File{Name: &ast.Ident{Name: "test2"}},
		&TypeSpecDef{
			TypeSpec: &ast.TypeSpec{
				Name:       &ast.Ident{Name: "Field"},
				TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}}},
				Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
			}}, "test.Field[string, string]")
	assert.Nil(t, typeSpec)

	// definition contains two type params, but only one is used
	typeSpec = pd.parametrizeGenericType(
		&ast.File{Name: &ast.Ident{Name: "test2"}},
		&TypeSpecDef{
			TypeSpec: &ast.TypeSpec{
				Name:       &ast.Ident{Name: "Field"},
				TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
				Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
			}}, "test.Field[string]")
	assert.Nil(t, typeSpec)

	// name is not a valid type name
	typeSpec = pd.parametrizeGenericType(
		&ast.File{Name: &ast.Ident{Name: "test2"}},
		&TypeSpecDef{
			TypeSpec: &ast.TypeSpec{
				Name:       &ast.Ident{Name: "Field"},
				TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
				Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
			}}, "test.Field[string")
	assert.Nil(t, typeSpec)

	typeSpec = pd.parametrizeGenericType(
		&ast.File{Name: &ast.Ident{Name: "test2"}},
		&TypeSpecDef{
			TypeSpec: &ast.TypeSpec{
				Name:       &ast.Ident{Name: "Field"},
				TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
				Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
			}}, "test.Field[string, [string]")
	assert.Nil(t, typeSpec)

	typeSpec = pd.parametrizeGenericType(
		&ast.File{Name: &ast.Ident{Name: "test2"}},
		&TypeSpecDef{
			TypeSpec: &ast.TypeSpec{
				Name:       &ast.Ident{Name: "Field"},
				TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}, {Names: []*ast.Ident{{Name: "T2"}}}}},
				Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
			}}, "test.Field[string, ]string]")
	assert.Nil(t, typeSpec)
}

func TestSplitGenericsTypeNames(t *testing.T) {
	t.Parallel()

	field, params := splitGenericsTypeName("test.Field")
	assert.Empty(t, field)
	assert.Nil(t, params)

	field, params = splitGenericsTypeName("test.Field]")
	assert.Empty(t, field)
	assert.Nil(t, params)

	field, params = splitGenericsTypeName("test.Field[string")
	assert.Empty(t, field)
	assert.Nil(t, params)

	field, params = splitGenericsTypeName("test.Field[string] ")
	assert.Equal(t, "test.Field", field)
	assert.Equal(t, []string{"string"}, params)

	field, params = splitGenericsTypeName("test.Field[string, []string]")
	assert.Equal(t, "test.Field", field)
	assert.Equal(t, []string{"string", "[]string"}, params)

	field, params = splitGenericsTypeName("test.Field[test.Field[ string, []string] ]")
	assert.Equal(t, "test.Field", field)
	assert.Equal(t, []string{"test.Field[string,[]string]"}, params)
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

func TestParseGenericTypeExpr(t *testing.T) {
	t.Parallel()

	parser := New()
	logger := &testLogger{}
	SetDebugger(logger)(parser)

	_, _ = parser.parseGenericTypeExpr(&ast.File{}, &ast.InterfaceType{})
	assert.Empty(t, logger.Messages)
	_, _ = parser.parseGenericTypeExpr(&ast.File{}, &ast.StructType{})
	assert.Empty(t, logger.Messages)
	_, _ = parser.parseGenericTypeExpr(&ast.File{}, &ast.Ident{})
	assert.Empty(t, logger.Messages)
	_, _ = parser.parseGenericTypeExpr(&ast.File{}, &ast.StarExpr{})
	assert.Empty(t, logger.Messages)
	_, _ = parser.parseGenericTypeExpr(&ast.File{}, &ast.SelectorExpr{})
	assert.Empty(t, logger.Messages)
	_, _ = parser.parseGenericTypeExpr(&ast.File{}, &ast.ArrayType{})
	assert.Empty(t, logger.Messages)
	_, _ = parser.parseGenericTypeExpr(&ast.File{}, &ast.MapType{})
	assert.Empty(t, logger.Messages)
	_, _ = parser.parseGenericTypeExpr(&ast.File{}, &ast.FuncType{})
	assert.Empty(t, logger.Messages)
	_, _ = parser.parseGenericTypeExpr(&ast.File{}, &ast.BadExpr{})
	assert.NotEmpty(t, logger.Messages)
	assert.Len(t, logger.Messages, 1)

	parser.packages.uniqueDefinitions["field.Name[string]"] = &TypeSpecDef{
		File: &ast.File{Name: &ast.Ident{Name: "test"}},
		TypeSpec: &ast.TypeSpec{
			Name:       &ast.Ident{Name: "Field"},
			TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}}},
			Type:       &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
		},
	}
	spec, err := parser.parseTypeExpr(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "field"}, Sel: &ast.Ident{Name: "Name"}}, Index: &ast.Ident{Name: "string"}},
		false,
	)
	assert.NotNil(t, spec)
	assert.NoError(t, err)

	logger.Messages = []string{}
	spec, err = parser.parseTypeExpr(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.IndexExpr{X: &ast.BadExpr{}, Index: &ast.Ident{Name: "string"}},
		false,
	)
	assert.NotNil(t, spec)
	assert.Equal(t, "object", spec.SchemaProps.Type[0])
	assert.NotEmpty(t, logger.Messages)
	assert.Len(t, logger.Messages, 1)

	logger.Messages = []string{}
	spec, err = parser.parseTypeExpr(
		&ast.File{Name: &ast.Ident{Name: "test"}},
		&ast.BadExpr{},
		false,
	)
	assert.NotNil(t, spec)
	assert.Equal(t, "object", spec.SchemaProps.Type[0])
	assert.NotEmpty(t, logger.Messages)
	assert.Len(t, logger.Messages, 1)
}
