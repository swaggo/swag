package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidDataType(t *testing.T) {
	assert.NoError(t, CheckSchemaType("string"))
	assert.NoError(t, CheckSchemaType("number"))
	assert.NoError(t, CheckSchemaType("integer"))
	assert.NoError(t, CheckSchemaType("boolean"))
	assert.NoError(t, CheckSchemaType("array"))
	assert.NoError(t, CheckSchemaType("object"))

	assert.Error(t, CheckSchemaType("oops"))
}

func TestTransToValidSchemeType(t *testing.T) {
	assert.Equal(t, TransToValidSchemeType("uint"), "integer")
	assert.Equal(t, TransToValidSchemeType("uint32"), "integer")
	assert.Equal(t, TransToValidSchemeType("uint64"), "integer")
	assert.Equal(t, TransToValidSchemeType("float32"), "number")
	assert.Equal(t, TransToValidSchemeType("bool"), "boolean")
	assert.Equal(t, TransToValidSchemeType("string"), "string")

	// should accept any type, due to user defined types
	TransToValidSchemeType("oops")
}

func TestIsGolangPrimitiveType(t *testing.T) {

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

func TestIsNumericType(t *testing.T) {
	assert.Equal(t, IsNumericType("integer"), true)
	assert.Equal(t, IsNumericType("number"), true)

	assert.Equal(t, IsNumericType("string"), false)
}
