package swag

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sv-tools/openapi/spec"
)

func TestDefaultFieldParserV3(t *testing.T) {
	t.Run("Example tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" example:"one"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, "one", schema.Spec.Example)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" example:""`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, "", schema.Spec.Example)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{"float"}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" example:"one"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)
	})

	t.Run("Format tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" format:"csv"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, "csv", schema.Spec.Format)
	})

	t.Run("Required tag", func(t *testing.T) {
		t.Parallel()

		got, err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" binding:"required"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.Equal(t, true, got)

		got, err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.Equal(t, true, got)
	})

	t.Run("Default required tag", func(t *testing.T) {
		t.Parallel()

		got, err := newTagBaseFieldParserV3(
			&Parser{
				RequiredByDefault: true,
			},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("Optional tag", func(t *testing.T) {
		t.Parallel()

		got, err := newTagBaseFieldParserV3(
			&Parser{
				RequiredByDefault: true,
			},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" binding:"optional"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.False(t, got)

		got, err = newTagBaseFieldParserV3(
			&Parser{
				RequiredByDefault: true,
			},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"optional"`,
			}},
		).IsRequired()
		assert.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("Skipped tag", func(t *testing.T) {
		t.Parallel()

		got, err := newTagBaseFieldParserV3(
			&Parser{
				RequiredByDefault: true,
			},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"-"`,
			}},
		).FieldName()
		assert.NoError(t, err)
		assert.Empty(t, got)

		got, err = newTagBaseFieldParserV3(
			&Parser{
				RequiredByDefault: true,
			},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `form:"-"`,
			}},
		).FieldName()
		assert.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("Extensions tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		schema.Spec.Extensions = map[string]interface{}{}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" extensions:"x-nullable,x-abc=def,!x-omitempty,x-example=[0, 9],x-example2={çãíœ, (bar=(abc, def)), [0,9]}"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, true, schema.Spec.Extensions["x-nullable"])
		assert.Equal(t, "def", schema.Spec.Extensions["x-abc"])
		assert.Equal(t, false, schema.Spec.Extensions["x-omitempty"])
		assert.Equal(t, "[0, 9]", schema.Spec.Extensions["x-example"])
		assert.Equal(t, "{çãíœ, (bar=(abc, def)), [0,9]}", schema.Spec.Extensions["x-example2"])
	})

	t.Run("Enums tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"a,b,c"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"a", "b", "c"}, schema.Spec.Enum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{"float"}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"a,b,c"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)
	})

	t.Run("Enums tag twice", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()

		typeArray := spec.NewSingleOrArray("string")
		schema.Spec.Type = &typeArray

		parser := &Parser{}
		fieldParser := newTagBaseFieldParserV3(
			parser,
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"a,b,c"`,
			}},
		)
		err := fieldParser.ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"a", "b", "c"}, schema.Spec.Enum)

		fieldParser2 := newTagBaseFieldParserV3(
			parser,
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"d,e,f"`,
			}},
		)
		fieldParser2.ComplementSchema(schema)
		assert.Equal(t, []interface{}{"a", "b", "c", "d", "e", "f"}, schema.Spec.Enum)

	})

	t.Run("EnumVarNames tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		schema.Spec.Extensions = map[string]interface{}{}
		schema.Spec.Enum = []interface{}{}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"0,1,2" x-enum-varnames:"Daily,Weekly,Monthly"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"Daily", "Weekly", "Monthly"}, schema.Spec.Extensions["x-enum-varnames"])

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"0,1,2,3" x-enum-varnames:"Daily,Weekly,Monthly"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)

		// Test for an array of enums
		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{ARRAY}
		schema.Spec.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
		schema.Spec.Items.Schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}

		schema.Spec.Extensions = map[string]interface{}{}
		schema.Spec.Enum = []interface{}{}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" enums:"0,1,2" x-enum-varnames:"Daily,Weekly,Monthly"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"Daily", "Weekly", "Monthly"}, schema.Spec.Items.Schema.Spec.Extensions["x-enum-varnames"])
		assert.Equal(t, map[string]any{}, schema.Spec.Extensions)
	})

	t.Run("Default tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" default:"pass"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, "pass", schema.Spec.Default)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{"float"}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" default:"pass"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)
	})

	t.Run("Numeric value", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maximum:"1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		max := int(1)
		assert.Equal(t, &max, schema.Spec.Maximum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maximum:"one"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{NUMBER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maximum:"1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		max = int(1)
		assert.Equal(t, &max, schema.Spec.Maximum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{NUMBER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maximum:"one"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{NUMBER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" multipleOf:"1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		multipleOf := int(1)
		assert.Equal(t, &multipleOf, schema.Spec.MultipleOf)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{NUMBER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" multipleOf:"one"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" minimum:"1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		min := int(1)
		assert.Equal(t, &min, schema.Spec.Minimum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" minimum:"one"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)
	})

	t.Run("String value", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maxLength:"1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		max := int(1)
		assert.Equal(t, &max, schema.Spec.MaxLength)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" maxLength:"one"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" minLength:"1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		min := int(1)
		assert.Equal(t, &min, schema.Spec.MinLength)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" minLength:"one"`,
			}},
		).ComplementSchema(schema)
		assert.Error(t, err)
	})

	t.Run("Readonly tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" readonly:"true"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, true, schema.Spec.ReadOnly)
	})

	t.Run("OneOf tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{ANY}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" oneOf:"string,float64"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Len(t, schema.Spec.OneOf, 2)
		assert.Equal(t, &spec.SingleOrArray[string]{STRING}, schema.Spec.OneOf[0].Spec.Type)
		assert.Equal(t, &spec.SingleOrArray[string]{NUMBER}, schema.Spec.OneOf[1].Spec.Type)
	})
}

func TestValidTagsV3(t *testing.T) {
	t.Run("Required with max/min tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,min=1"`,
			}},
		).ComplementSchema(schema)
		max := int(10)
		min := int(1)
		assert.NoError(t, err)
		assert.Equal(t, &max, schema.Spec.MaxLength)
		assert.Equal(t, &min, schema.Spec.MinLength)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,gte=1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, &max, schema.Spec.MaxLength)
		assert.Equal(t, &min, schema.Spec.MinLength)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,min=1"`,
			}},
		).ComplementSchema(schema)
		maxFloat64 := int(10)
		minFloat64 := int(1)
		assert.NoError(t, err)
		assert.Equal(t, &maxFloat64, schema.Spec.Maximum)
		assert.Equal(t, &minFloat64, schema.Spec.Minimum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{ARRAY}

		schema.Spec.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
		schema.Spec.Items.Schema.Spec.Type = &spec.SingleOrArray[string]{STRING}

		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,min=1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, &max, schema.Spec.MaxItems)
		assert.Equal(t, &min, schema.Spec.MinItems)

		// wrong validate tag will be ignored.
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=ten,min=1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Empty(t, schema.Spec.MaxItems)
		assert.Equal(t, &min, schema.Spec.MinItems)
	})
	t.Run("Required with oneof tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}

		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof='red book' 'green book'"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"red book", "green book"}, schema.Spec.Enum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof=1 2 3"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{1, 2, 3}, schema.Spec.Enum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{ARRAY}

		schema.Spec.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
		schema.Spec.Items.Schema.Spec.Type = &spec.SingleOrArray[string]{STRING}

		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof=red green yellow"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"red", "green", "yellow"}, schema.Spec.Items.Schema.Spec.Enum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof='red green' blue 'c0x2Cc' 'd0x7Cd'"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"red green", "blue", "c,c", "d|d"}, schema.Spec.Enum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof='c0x9Ab' book"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"c0x9Ab", "book"}, schema.Spec.Enum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" binding:"oneof=foo bar" validate:"required,oneof=foo bar" enums:"a,b,c"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"a", "b", "c"}, schema.Spec.Enum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" binding:"oneof=aa bb" validate:"required,oneof=foo bar"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"aa", "bb"}, schema.Spec.Enum)
	})
	t.Run("Required with unique tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{ARRAY}
		schema.Spec.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
		schema.Spec.Items.Schema.Spec.Type = &spec.SingleOrArray[string]{STRING}

		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,unique"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.True(t, *schema.Spec.UniqueItems)
	})

	t.Run("All tag", func(t *testing.T) {
		t.Parallel()
		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{ARRAY}
		schema.Spec.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
		schema.Spec.Items.Schema.Spec.Type = &spec.SingleOrArray[string]{STRING}

		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,unique,max=10,min=1,oneof=a0x2Cc 'c0x7Cd book',omitempty,dive,max=1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.True(t, *schema.Spec.UniqueItems)

		max := int(10)
		min := int(1)
		assert.Equal(t, &max, schema.Spec.MaxItems)
		assert.Equal(t, &min, schema.Spec.MinItems)
		assert.Equal(t, []interface{}{"a,c", "c|d book"}, schema.Spec.Items.Schema.Spec.Enum)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{ARRAY}
		schema.Spec.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
		schema.Spec.Items.Schema.Spec.Type = &spec.SingleOrArray[string]{STRING}

		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof=,max=10=90,min=1"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Empty(t, schema.Spec.UniqueItems)
		assert.Empty(t, schema.Spec.MaxItems)
		assert.Equal(t, &min, schema.Spec.MinItems)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{ARRAY}
		schema.Spec.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
		schema.Spec.Items.Schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,max=10,min=one"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, &max, schema.Spec.MaxItems)
		assert.Empty(t, schema.Spec.MinItems)

		schema = spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{INTEGER}
		err = newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" validate:"required,oneof=one two"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Empty(t, schema.Spec.Enum)
	})

	t.Run("Pattern tag", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &spec.SingleOrArray[string]{ARRAY}
		schema.Spec.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
		schema.Spec.Items.Schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" pattern:"^[a-zA-Z0-9_]*$"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, "^[a-zA-Z0-9_]*$", schema.Spec.Items.Schema.Spec.Pattern)
	})

	t.Run("Pattern tag array", func(t *testing.T) {
		t.Parallel()

		schema := spec.NewSchemaSpec()
		schema.Spec.Type = &typeString
		err := newTagBaseFieldParserV3(
			&Parser{},
			&ast.File{Name: &ast.Ident{Name: "test"}},
			&ast.Field{Tag: &ast.BasicLit{
				Value: `json:"test" pattern:"^[a-zA-Z0-9_]*$"`,
			}},
		).ComplementSchema(schema)
		assert.NoError(t, err)
		assert.Equal(t, "^[a-zA-Z0-9_]*$", schema.Spec.Pattern)
	})
}
