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
}
