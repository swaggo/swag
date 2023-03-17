//go:build !go1.18
// +build !go1.18

package swag

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go/ast"
	"testing"
)

type testLogger struct {
	Messages []string
}

func (t *testLogger) Printf(format string, v ...interface{}) {
	t.Messages = append(t.Messages, fmt.Sprintf(format, v...))
}

func TestParametrizeStruct(t *testing.T) {
	t.Parallel()

	pd := PackagesDefinitions{
		packages: make(map[string]*PackageDefinitions),
	}

	tSpec := &TypeSpecDef{
		TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{Name: "Field"},
			Type: &ast.StructType{Struct: 100, Fields: &ast.FieldList{Opening: 101, Closing: 102}},
		},
	}

	tr := pd.parametrizeGenericType(&ast.File{}, tSpec, "")
	assert.Equal(t, tr, tSpec)

	tr = pd.parametrizeGenericType(&ast.File{}, tSpec, "")
	assert.Equal(t, tr, tSpec)
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
}
