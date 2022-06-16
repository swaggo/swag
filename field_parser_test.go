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
		schema.Type = []string{"string"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" example:""`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, "", schema.Example)

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

	t.Run("Default required tag", func(t *testing.T) {
		t.Parallel()

		got, err := newTagBaseFieldParser(
			&Parser{
				RequiredByDefault: true,
			},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("Optional tag", func(t *testing.T) {
		t.Parallel()

		got, err := newTagBaseFieldParser(
			&Parser{
				RequiredByDefault: true,
			},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" binding:"optional"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.False(t, got)

		got, err = newTagBaseFieldParser(
			&Parser{
				RequiredByDefault: true,
			},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"optional"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("Extensions tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"int"}
		schema.Extensions = map[string]interface{}{}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" extensions:"x-nullable,x-abc=def,!x-omitempty,x-example=[0, 9],x-example2={çãíœ, (bar=(abc, def)), [0,9]}"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, true, schema.Extensions["x-nullable"])
		assert.Equal(t, "def", schema.Extensions["x-abc"])
		assert.Equal(t, false, schema.Extensions["x-omitempty"])
		assert.Equal(t, "[0, 9]", schema.Extensions["x-example"])
		assert.Equal(t, "{çãíœ, (bar=(abc, def)), [0,9]}", schema.Extensions["x-example2"])
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

	t.Run("EnumVarNames tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"int"}
		schema.Extensions = map[string]interface{}{}
		schema.Enum = []interface{}{}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"0,1,2" x-enum-varnames:"Daily,Weekly,Monthly"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"Daily", "Weekly", "Monthly"}, schema.Extensions["x-enum-varnames"])

		schema = spec.Schema{}
		schema.Type = []string{"int"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"0,1,2,3" x-enum-varnames:"Daily,Weekly,Monthly"`,
			}},
		).ComplementSchema(&schema)
		assert.Error(t, err)

		// Test for an array of enums
		schema = spec.Schema{}
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"int"},
				},
			},
		}
		schema.Extensions = map[string]interface{}{}
		schema.Enum = []interface{}{}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"0,1,2" x-enum-varnames:"Daily,Weekly,Monthly"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"Daily", "Weekly", "Monthly"}, schema.Items.Schema.Extensions["x-enum-varnames"])
		assert.Equal(t, spec.Extensions{}, schema.Extensions)
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

	t.Run("Invalid tag", func(t *testing.T) {
		t.Parallel()

		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Names: []*ast.Ident{{Name: "BasicStruct"}}},
		).ComplementSchema(nil)
		assert.Error(t, err)
	})
}

func TestValidTags(t *testing.T) {
	t.Run("Required with max/min tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"string"}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,min=1"`,
			}},
		).ComplementSchema(&schema)
		max := int64(10)
		min := int64(1)
		assert.NoError(t, err)
		assert.Equal(t, &max, schema.MaxLength)
		assert.Equal(t, &min, schema.MinLength)

		schema = spec.Schema{}
		schema.Type = []string{"string"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,gte=1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, &max, schema.MaxLength)
		assert.Equal(t, &min, schema.MinLength)

		schema = spec.Schema{}
		schema.Type = []string{"integer"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,min=1"`,
			}},
		).ComplementSchema(&schema)
		maxFloat64 := float64(10)
		minFloat64 := float64(1)
		assert.NoError(t, err)
		assert.Equal(t, &maxFloat64, schema.Maximum)
		assert.Equal(t, &minFloat64, schema.Minimum)

		schema = spec.Schema{}
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				},
			},
		}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,min=1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, &max, schema.MaxItems)
		assert.Equal(t, &min, schema.MinItems)

		// wrong validate tag will be ignored.
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=ten,min=1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Empty(t, schema.MaxItems)
		assert.Equal(t, &min, schema.MinItems)
	})
	t.Run("Required with oneof tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"string"}

		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof='red book' 'green book'"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"red book", "green book"}, schema.Enum)

		schema = spec.Schema{}
		schema.Type = []string{"integer"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof=1 2 3"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{1, 2, 3}, schema.Enum)

		schema = spec.Schema{}
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				},
			},
		}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof=red green yellow"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"red", "green", "yellow"}, schema.Items.Schema.Enum)

		schema = spec.Schema{}
		schema.Type = []string{"string"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof='red green' blue 'c0x2Cc' 'd0x7Cd'"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"red green", "blue", "c,c", "d|d"}, schema.Enum)

		schema = spec.Schema{}
		schema.Type = []string{"string"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof='c0x9Ab' book"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"c0x9Ab", "book"}, schema.Enum)

		schema = spec.Schema{}
		schema.Type = []string{"string"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" binding:"oneof=foo bar" validate:"required,oneof=foo bar" enums:"a,b,c"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"a", "b", "c"}, schema.Enum)

		schema = spec.Schema{}
		schema.Type = []string{"string"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" binding:"oneof=aa bb" validate:"required,oneof=foo bar"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"aa", "bb"}, schema.Enum)
	})
	t.Run("Required with unique tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.Schema{}
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				},
			},
		}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,unique"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, true, schema.UniqueItems)
	})

	t.Run("All tag", func(t *testing.T) {
		t.Parallel()
		schema := spec.Schema{}
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				},
			},
		}
		err := newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,unique,max=10,min=1,oneof=a0x2Cc 'c0x7Cd book',omitempty,dive,max=1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, true, schema.UniqueItems)

		max := int64(10)
		min := int64(1)
		assert.Equal(t, &max, schema.MaxItems)
		assert.Equal(t, &min, schema.MinItems)
		assert.Equal(t, []interface{}{"a,c", "c|d book"}, schema.Items.Schema.Enum)

		schema = spec.Schema{}
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				},
			},
		}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof=,max=10=90,min=1"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Empty(t, schema.UniqueItems)
		assert.Empty(t, schema.MaxItems)
		assert.Equal(t, &min, schema.MinItems)

		schema = spec.Schema{}
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				},
			},
		}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,min=one"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Equal(t, &max, schema.MaxItems)
		assert.Empty(t, schema.MinItems)

		schema = spec.Schema{}
		schema.Type = []string{"integer"}
		err = newTagBaseFieldParser(
			&Parser{},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof=one two"`,
			}},
		).ComplementSchema(&schema)
		assert.NoError(t, err)
		assert.Empty(t, schema.Enum)
	})
}
