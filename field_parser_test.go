package swag

import (
	"go/ast"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

func TestDefaultFieldParser(t *testing.T) {
	t.Run("Example tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"string"}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" example:"one"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, "one", schema.Example)

		schema = spec.Schema{}
		schema.Type = []string{"float"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" example:"one"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)
	})

	t.Run("Format tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"string"}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" format:"csv"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, "csv", schema.Format)
	})

	t.Run("Required tag", func(t *testing.T) {
		t.Parallel()

		got, err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" binding:"required"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.Equal(t, true, got)

		got, err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.Equal(t, true, got)
	})

	t.Run("Extensions tag", func(t *testing.T) {

	})

	t.Run("Enums tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"string"}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"a,b,c"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"a", "b", "c"}, schema.Enum)

		schema = spec.Schema{}
		schema.Type = []string{"float"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"a,b,c"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)
	})

	t.Run("Default tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"string"}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" default:"pass"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, "pass", schema.Default)

		schema = spec.Schema{}
		schema.Type = []string{"float"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" default:"pass"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)
	})

	t.Run("Numeric value", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"integer"}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maximum:"1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		max := float64(1)
		assert.Equal(t, &max, schema.Maximum)

		schema = spec.Schema{}
		schema.Type = []string{"integer"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maximum:"one"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)

		schema = spec.Schema{}
		schema.Type = []string{"number"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maximum:"1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		max = float64(1)
		assert.Equal(t, &max, schema.Maximum)

		schema = spec.Schema{}
		schema.Type = []string{"number"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maximum:"one"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)

		schema = spec.Schema{}
		schema.Type = []string{"number"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" multipleOf:"1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		multipleOf := float64(1)
		assert.Equal(t, &multipleOf, schema.MultipleOf)

		schema = spec.Schema{}
		schema.Type = []string{"number"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" multipleOf:"one"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)

		schema = spec.Schema{}
		schema.Type = []string{"integer"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" minimum:"1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		min := float64(1)
		assert.Equal(t, &min, schema.Minimum)

		schema = spec.Schema{}
		schema.Type = []string{"integer"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" minimum:"one"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)
	})

	t.Run("String value", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"string"}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maxLength:"1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		max := int64(1)
		assert.Equal(t, &max, schema.MaxLength)

		schema = spec.Schema{}
		schema.Type = []string{"string"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maxLength:"one"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)

		schema = spec.Schema{}
		schema.Type = []string{"string"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" minLength:"1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		min := int64(1)
		assert.Equal(t, &min, schema.MinLength)

		schema = spec.Schema{}
		schema.Type = []string{"string"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" minLength:"one"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)
	})

	t.Run("Readonly tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"string"}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" readonly:"true"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, true, schema.ReadOnly)
	})
}
