package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sv-tools/openapi/spec"
)

func TestBuildCustomSchemaV3(t *testing.T) {
	t.Parallel()

	t.Run("array of primitive puts the element schema in items", func(t *testing.T) {
		t.Parallel()
		schema, err := BuildCustomSchemaV3([]string{"array", "integer"})
		assert.NoError(t, err)
		assert.Equal(t, &spec.SingleOrArray[string]{ARRAY}, schema.Spec.Type)
		assert.Nil(t, schema.Spec.AdditionalProperties)
		if assert.NotNil(t, schema.Spec.Items) && assert.NotNil(t, schema.Spec.Items.Schema) {
			assert.Equal(t, &spec.SingleOrArray[string]{INTEGER}, schema.Spec.Items.Schema.Spec.Type)
		}
	})

	t.Run("array without item type errors", func(t *testing.T) {
		t.Parallel()
		_, err := BuildCustomSchemaV3([]string{"array"})
		assert.Error(t, err)
	})

	t.Run("object of primitive keeps additionalProperties", func(t *testing.T) {
		t.Parallel()
		schema, err := BuildCustomSchemaV3([]string{"object", "string"})
		assert.NoError(t, err)
		assert.Equal(t, &spec.SingleOrArray[string]{OBJECT}, schema.Spec.Type)
		assert.NotNil(t, schema.Spec.AdditionalProperties)
		assert.Nil(t, schema.Spec.Items)
	})

	t.Run("primitive", func(t *testing.T) {
		t.Parallel()
		schema, err := BuildCustomSchemaV3([]string{"string"})
		assert.NoError(t, err)
		assert.Equal(t, &spec.SingleOrArray[string]{STRING}, schema.Spec.Type)
	})
}
