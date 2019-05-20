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
