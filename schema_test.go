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

func TestIsMineType(t *testing.T) {
	assert.Equal(t, IsMineType("json"), true)
	assert.Equal(t, IsMineType("application/json"), true)
	assert.Equal(t, IsMineType("xml"), true)
	assert.Equal(t, IsMineType("text/xml"), true)
	assert.Equal(t, IsMineType("plain"), true)
	assert.Equal(t, IsMineType("text/plain"), true)
	assert.Equal(t, IsMineType("html"), true)
	assert.Equal(t, IsMineType("text/html"), true)
	assert.Equal(t, IsMineType("mpfd"), true)
	assert.Equal(t, IsMineType("multipart/form-data"), true)
	assert.Equal(t, IsMineType("x-www-form-urlencoded"), true)
	assert.Equal(t, IsMineType("application/x-www-form-urlencoded"), true)
	assert.Equal(t, IsMineType("json-api"), true)
	assert.Equal(t, IsMineType("application/vnd.api+json"), true)
	assert.Equal(t, IsMineType("json-stream"), true)
	assert.Equal(t, IsMineType("application/x-json-stream"), true)
	assert.Equal(t, IsMineType("octet-stream"), true)
	assert.Equal(t, IsMineType("application/octet-stream"), true)
	assert.Equal(t, IsMineType("png"), true)
	assert.Equal(t, IsMineType("image/png"), true)
	assert.Equal(t, IsMineType("jpeg"), true)
	assert.Equal(t, IsMineType("image/jpeg"), true)
	assert.Equal(t, IsMineType("gif"), true)
	assert.Equal(t, IsMineType("image/gif"), true)
	assert.Equal(t, IsMineType("avatar"), false)
}
