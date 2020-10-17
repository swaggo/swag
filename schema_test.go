package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidDataType(t *testing.T) {
	assert.NoError(t, CheckSchemaType(STRING))
	assert.NoError(t, CheckSchemaType(NUMBER))
	assert.NoError(t, CheckSchemaType(INTEGER))
	assert.NoError(t, CheckSchemaType(BOOLEAN))
	assert.NoError(t, CheckSchemaType(ARRAY))
	assert.NoError(t, CheckSchemaType(OBJECT))

	assert.Error(t, CheckSchemaType("oops"))
}

func TestTransToValidSchemeType(t *testing.T) {
	assert.Equal(t, TransToValidSchemeType("uint"), INTEGER)
	assert.Equal(t, TransToValidSchemeType("uint32"), INTEGER)
	assert.Equal(t, TransToValidSchemeType("uint64"), INTEGER)
	assert.Equal(t, TransToValidSchemeType("float32"), NUMBER)
	assert.Equal(t, TransToValidSchemeType("bool"), BOOLEAN)
	assert.Equal(t, TransToValidSchemeType("string"), STRING)

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
	assert.Equal(t, IsNumericType(INTEGER), true)
	assert.Equal(t, IsNumericType(NUMBER), true)

	assert.Equal(t, IsNumericType(STRING), false)
}
