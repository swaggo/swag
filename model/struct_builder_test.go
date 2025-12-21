package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSpecSchema_Simple(t *testing.T) {
	builder := &StructBuilder{
		Fields: []*StructField{
			{
				Name:       "FirstName",
				TypeString: "string",
				Tag:        `json:"first_name"`,
			},
			{
				Name:       "LastName",
				TypeString: "string",
				Tag:        `json:"last_name,omitempty"`,
			},
			{
				Name:       "Age",
				TypeString: "int",
				Tag:        `json:"age"`,
			},
		},
	}

	schema, nestedTypes, err := builder.BuildSpecSchema("User", false, nil)
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, 1, len(schema.Type))
	assert.Equal(t, "object", schema.Type[0])
	assert.Equal(t, 3, len(schema.Properties))

	// Check properties
	assert.Contains(t, schema.Properties, "first_name")
	assert.Contains(t, schema.Properties, "last_name")
	assert.Contains(t, schema.Properties, "age")

	// Check required fields
	assert.Equal(t, 2, len(schema.Required))
	assert.Contains(t, schema.Required, "first_name")
	assert.Contains(t, schema.Required, "age")
	assert.NotContains(t, schema.Required, "last_name")

	// No nested types
	assert.Equal(t, 0, len(nestedTypes))
}

func TestBuildSpecSchema_WithNestedStruct(t *testing.T) {
	builder := &StructBuilder{
		Fields: []*StructField{
			{
				Name:       "Name",
				TypeString: "string",
				Tag:        `json:"name"`,
			},
			{
				Name:       "Address",
				TypeString: "fields.StructField[*Address]",
				Tag:        `json:"address"`,
			},
		},
	}

	schema, nestedTypes, err := builder.BuildSpecSchema("User", false, nil)
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, 2, len(schema.Properties))

	// Check nested type
	assert.Equal(t, 1, len(nestedTypes))
	assert.Contains(t, nestedTypes, "Address")

	// Check address property is a reference
	addressProp := schema.Properties["address"]
	assert.Equal(t, "#/definitions/Address", addressProp.Ref.String())
}

func TestBuildSpecSchema_PublicMode(t *testing.T) {
	builder := &StructBuilder{
		Fields: []*StructField{
			{
				Name:       "PublicField",
				TypeString: "string",
				Tag:        `public:"view" json:"public_field"`,
			},
			{
				Name:       "PrivateField",
				TypeString: "string",
				Tag:        `json:"private_field"`,
			},
		},
	}

	schema, nestedTypes, err := builder.BuildSpecSchema("User", true, nil)
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	// Only public field should be included
	assert.Equal(t, 1, len(schema.Properties))
	assert.Contains(t, schema.Properties, "public_field")
	assert.NotContains(t, schema.Properties, "private_field")

	// No nested types
	assert.Equal(t, 0, len(nestedTypes))
}

func TestBuildSpecSchema_PublicModeWithNestedStruct(t *testing.T) {
	builder := &StructBuilder{
		Fields: []*StructField{
			{
				Name:       "Name",
				TypeString: "string",
				Tag:        `public:"view" json:"name"`,
			},
			{
				Name:       "Profile",
				TypeString: "fields.StructField[*Profile]",
				Tag:        `public:"view" json:"profile"`,
			},
			{
				Name:       "InternalData",
				TypeString: "fields.StructField[*InternalData]",
				Tag:        `json:"internal_data"`,
			},
		},
	}

	schema, nestedTypes, err := builder.BuildSpecSchema("User", true, nil)
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	// Only public fields should be included
	assert.Equal(t, 2, len(schema.Properties))
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "profile")
	assert.NotContains(t, schema.Properties, "internal_data")

	// Only public nested type should be included
	assert.Equal(t, 1, len(nestedTypes))
	assert.Contains(t, nestedTypes, "Profile")
	assert.NotContains(t, nestedTypes, "InternalData")

	// Check profile property has Public suffix
	profileProp := schema.Properties["profile"]
	assert.Equal(t, "#/definitions/ProfilePublic", profileProp.Ref.String())
}

func TestBuildSpecSchema_MultipleNestedStructs(t *testing.T) {
	builder := &StructBuilder{
		Fields: []*StructField{
			{
				Name:       "User",
				TypeString: "fields.StructField[*User]",
				Tag:        `json:"user"`,
			},
			{
				Name:       "Address",
				TypeString: "fields.StructField[*Address]",
				Tag:        `json:"address"`,
			},
			{
				Name:       "SecondaryAddress",
				TypeString: "fields.StructField[*Address]",
				Tag:        `json:"secondary_address,omitempty"`,
			},
		},
	}

	schema, nestedTypes, err := builder.BuildSpecSchema("Contact", false, nil)
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	// Should deduplicate Address
	assert.Equal(t, 2, len(nestedTypes))
	assert.Contains(t, nestedTypes, "User")
	assert.Contains(t, nestedTypes, "Address")

	// Check required fields
	assert.Equal(t, 2, len(schema.Required))
	assert.Contains(t, schema.Required, "user")
	assert.Contains(t, schema.Required, "address")
	assert.NotContains(t, schema.Required, "secondary_address")
}

func TestBuildSpecSchema_ArrayOfStructs(t *testing.T) {
	builder := &StructBuilder{
		Fields: []*StructField{
			{
				Name:       "Name",
				TypeString: "string",
				Tag:        `json:"name"`,
			},
			{
				Name:       "Items",
				TypeString: "fields.StructField[[]Item]",
				Tag:        `json:"items"`,
			},
		},
	}

	schema, nestedTypes, err := builder.BuildSpecSchema("Order", false, nil)
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, 2, len(schema.Properties))

	// Check nested type
	assert.Equal(t, 1, len(nestedTypes))
	assert.Contains(t, nestedTypes, "Item")

	// Check items property is an array
	itemsProp := schema.Properties["items"]
	assert.Equal(t, 1, len(itemsProp.Type))
	assert.Equal(t, "array", itemsProp.Type[0])
	assert.NotNil(t, itemsProp.Items)
	assert.NotNil(t, itemsProp.Items.Schema)
	assert.Equal(t, "#/definitions/Item", itemsProp.Items.Schema.Ref.String())
}

func TestBuildSpecSchema_EmptyBuilder(t *testing.T) {
	builder := &StructBuilder{
		Fields: []*StructField{},
	}

	schema, nestedTypes, err := builder.BuildSpecSchema("Empty", false, nil)
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, 1, len(schema.Type))
	assert.Equal(t, "object", schema.Type[0])
	assert.Equal(t, 0, len(schema.Properties))
	assert.Equal(t, 0, len(schema.Required))
	assert.Equal(t, 0, len(nestedTypes))
}
