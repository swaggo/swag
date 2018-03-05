package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidDataType(t *testing.T) {
	assert.NotPanics(t, func() {
		CheckSchemaType("string")
	})
	assert.NotPanics(t, func() {
		CheckSchemaType("number")
	})
	assert.NotPanics(t, func() {
		CheckSchemaType("integer")
	})
	assert.NotPanics(t, func() {
		CheckSchemaType("boolean")
	})
	assert.NotPanics(t, func() {
		CheckSchemaType("array")
	})
	assert.NotPanics(t, func() {
		CheckSchemaType("object")
	})

	assert.Panics(t, func() {
		CheckSchemaType("oops")
	})
}

func TestTransToValidSchemeType(t *testing.T) {
	assert.Equal(t, TransToValidSchemeType("uint"), "integer")
	assert.Equal(t, TransToValidSchemeType("uint32"), "integer")
	assert.Equal(t, TransToValidSchemeType("uint64"), "integer")
	assert.Equal(t, TransToValidSchemeType("float32"), "number")
	assert.Equal(t, TransToValidSchemeType("bool"), "boolean")
	assert.Equal(t, TransToValidSchemeType("string"), "string")

	// should accept any type, due to user defined types
	assert.NotPanics(t, func() {
		TransToValidSchemeType("oops")
	})
}
