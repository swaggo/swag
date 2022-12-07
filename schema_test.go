package swag

import (
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

func TestValidDataType(t *testing.T) {
	t.Parallel()

	assert.NoError(t, CheckSchemaType(STRING))
	assert.NoError(t, CheckSchemaType(NUMBER))
	assert.NoError(t, CheckSchemaType(INTEGER))
	assert.NoError(t, CheckSchemaType(BOOLEAN))
	assert.NoError(t, CheckSchemaType(ARRAY))
	assert.NoError(t, CheckSchemaType(OBJECT))

	assert.Error(t, CheckSchemaType("oops"))
}

func TestTransToValidSchemeType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, TransToValidSchemeType("uint"), INTEGER)
	assert.Equal(t, TransToValidSchemeType("uint32"), INTEGER)
	assert.Equal(t, TransToValidSchemeType("uint64"), INTEGER)
	assert.Equal(t, TransToValidSchemeType("float32"), NUMBER)
	assert.Equal(t, TransToValidSchemeType("bool"), BOOLEAN)
	assert.Equal(t, TransToValidSchemeType("string"), STRING)

	// should accept any type, due to user defined types
	other := "oops"
	assert.Equal(t, TransToValidSchemeType(other), other)
}

func TestTransToValidCollectionFormat(t *testing.T) {
	t.Parallel()

	assert.Equal(t, TransToValidCollectionFormat("csv"), "csv")
	assert.Equal(t, TransToValidCollectionFormat("multi"), "multi")
	assert.Equal(t, TransToValidCollectionFormat("pipes"), "pipes")
	assert.Equal(t, TransToValidCollectionFormat("tsv"), "tsv")
	assert.Equal(t, TransToValidSchemeType("string"), STRING)

	// should accept any type, due to user defined types
	assert.Equal(t, TransToValidCollectionFormat("oops"), "")
}

func TestIsGolangPrimitiveType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, IsGolangPrimitiveType("uint"), true)
	assert.Equal(t, IsGolangPrimitiveType("int"), true)
	assert.Equal(t, IsGolangPrimitiveType("uint8"), true)
	assert.Equal(t, IsGolangPrimitiveType("uint16"), true)
	assert.Equal(t, IsGolangPrimitiveType("int16"), true)
	assert.Equal(t, IsGolangPrimitiveType("byte"), true)
	assert.Equal(t, IsGolangPrimitiveType("uint32"), true)
	assert.Equal(t, IsGolangPrimitiveType("int32"), true)
	assert.Equal(t, IsGolangPrimitiveType("rune"), true)
	assert.Equal(t, IsGolangPrimitiveType("uint64"), true)
	assert.Equal(t, IsGolangPrimitiveType("int64"), true)
	assert.Equal(t, IsGolangPrimitiveType("float32"), true)
	assert.Equal(t, IsGolangPrimitiveType("float64"), true)
	assert.Equal(t, IsGolangPrimitiveType("bool"), true)
	assert.Equal(t, IsGolangPrimitiveType("string"), true)

	assert.Equal(t, IsGolangPrimitiveType("oops"), false)
}

func TestIsSimplePrimitiveType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, IsSimplePrimitiveType("string"), true)
	assert.Equal(t, IsSimplePrimitiveType("number"), true)
	assert.Equal(t, IsSimplePrimitiveType("integer"), true)
	assert.Equal(t, IsSimplePrimitiveType("boolean"), true)

	assert.Equal(t, IsSimplePrimitiveType("oops"), false)
}

func TestBuildCustomSchema(t *testing.T) {
	t.Parallel()

	var (
		schema *spec.Schema
		err    error
	)

	schema, err = BuildCustomSchema([]string{})
	assert.NoError(t, err)
	assert.Nil(t, schema)

	schema, err = BuildCustomSchema([]string{"primitive"})
	assert.Error(t, err)
	assert.Nil(t, schema)

	schema, err = BuildCustomSchema([]string{"primitive", "oops"})
	assert.Error(t, err)
	assert.Nil(t, schema)

	schema, err = BuildCustomSchema([]string{"primitive", "string"})
	assert.NoError(t, err)
	assert.Equal(t, schema.SchemaProps.Type, spec.StringOrArray{"string"})

	schema, err = BuildCustomSchema([]string{"array"})
	assert.Error(t, err)
	assert.Nil(t, schema)

	schema, err = BuildCustomSchema([]string{"array", "oops"})
	assert.Error(t, err)
	assert.Nil(t, schema)

	schema, err = BuildCustomSchema([]string{"array", "string"})
	assert.NoError(t, err)
	assert.Equal(t, schema.SchemaProps.Type, spec.StringOrArray{"array"})
	assert.Equal(t, schema.SchemaProps.Items.Schema.SchemaProps.Type, spec.StringOrArray{"string"})

	schema, err = BuildCustomSchema([]string{"object"})
	assert.NoError(t, err)
	assert.Equal(t, schema.SchemaProps.Type, spec.StringOrArray{"object"})

	schema, err = BuildCustomSchema([]string{"object", "oops"})
	assert.Error(t, err)
	assert.Nil(t, schema)

	schema, err = BuildCustomSchema([]string{"object", "string"})
	assert.NoError(t, err)
	assert.Equal(t, schema.SchemaProps.Type, spec.StringOrArray{"object"})
	assert.Equal(t, schema.SchemaProps.AdditionalProperties.Schema.Type, spec.StringOrArray{"string"})
}

func TestIsNumericType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, IsNumericType(INTEGER), true)
	assert.Equal(t, IsNumericType(NUMBER), true)

	assert.Equal(t, IsNumericType(STRING), false)
}

func TestIsInterfaceLike(t *testing.T) {
	t.Parallel()

	assert.Equal(t, IsInterfaceLike(ERROR), true)
	assert.Equal(t, IsInterfaceLike(ANY), true)

	assert.Equal(t, IsInterfaceLike(STRING), false)
}
