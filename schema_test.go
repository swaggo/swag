package swag

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
