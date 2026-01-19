package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSpecSchema_PrimitiveTypes(t *testing.T) {
	tests := []struct {
		name         string
		field        *StructField
		public       bool
		wantPropName string
		wantType     []string
		wantFormat   string
		wantRequired bool
		wantNested   int
	}{
		{
			name: "string field with json tag",
			field: &StructField{
				Name:       "FirstName",
				TypeString: "string",
				Tag:        `json:"first_name"`,
			},
			public:       false,
			wantPropName: "first_name",
			wantType:     []string{"string"},
			wantRequired: true,
			wantNested:   0,
		},
		{
			name: "int field with omitempty",
			field: &StructField{
				Name:       "Age",
				TypeString: "int",
				Tag:        `json:"age,omitempty"`,
			},
			public:       false,
			wantPropName: "age",
			wantType:     []string{"integer"},
			wantRequired: false,
			wantNested:   0,
		},
		{
			name: "int64 field",
			field: &StructField{
				Name:       "ID",
				TypeString: "int64",
				Tag:        `json:"id"`,
			},
			public:       false,
			wantPropName: "id",
			wantType:     []string{"integer"},
			wantFormat:   "int64",
			wantRequired: true,
			wantNested:   0,
		},
		{
			name: "bool field",
			field: &StructField{
				Name:       "Active",
				TypeString: "bool",
				Tag:        `json:"active"`,
			},
			public:       false,
			wantPropName: "active",
			wantType:     []string{"boolean"},
			wantRequired: true,
			wantNested:   0,
		},
		{
			name: "float64 field",
			field: &StructField{
				Name:       "Price",
				TypeString: "float64",
				Tag:        `json:"price"`,
			},
			public:       false,
			wantPropName: "price",
			wantType:     []string{"number"},
			wantFormat:   "double",
			wantRequired: true,
			wantNested:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			propName, schema, required, nestedTypes, err := tt.field.ToSpecSchema(tt.public, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantPropName, propName)
			assert.Equal(t, tt.wantRequired, required)
			assert.Equal(t, tt.wantNested, len(nestedTypes))
			if schema != nil {
				assert.Equal(t, len(tt.wantType), len(schema.Type))
				if len(tt.wantType) > 0 {
					assert.Equal(t, tt.wantType[0], schema.Type[0])
				}
				if tt.wantFormat != "" {
					assert.Equal(t, tt.wantFormat, schema.Format)
				}
			}
		})
	}
}

func TestToSpecSchema_StructField_Simple(t *testing.T) {
	field := &StructField{
		Name:       "Properties",
		TypeString: "fields.StructField[*Properties]",
		Tag:        `json:"properties"`,
	}

	propName, schema, required, nestedTypes, err := field.ToSpecSchema(false, nil)
	assert.NoError(t, err)
	assert.Equal(t, "properties", propName)
	assert.True(t, required)
	assert.Equal(t, 1, len(nestedTypes))
	assert.Equal(t, "Properties", nestedTypes[0])
	assert.NotNil(t, schema)
	assert.Equal(t, "#/definitions/Properties", schema.Ref.String())
}

func TestToSpecSchema_StructField_Public(t *testing.T) {
	field := &StructField{
		Name:       "User",
		TypeString: "fields.StructField[*User]",
		Tag:        `public:"view" json:"user"`,
	}

	propName, schema, required, nestedTypes, err := field.ToSpecSchema(true, nil)
	assert.NoError(t, err)
	assert.Equal(t, "user", propName)
	assert.True(t, required)
	assert.Equal(t, 1, len(nestedTypes))
	assert.Equal(t, "User", nestedTypes[0])
	assert.NotNil(t, schema)
	assert.Equal(t, "#/definitions/UserPublic", schema.Ref.String())
}

func TestToSpecSchema_StructField_NotPublic(t *testing.T) {
	field := &StructField{
		Name:       "InternalData",
		TypeString: "fields.StructField[*InternalData]",
		Tag:        `json:"internal_data"`,
	}

	// When public=true but field has no public tag, should return nil
	propName, schema, required, nestedTypes, err := field.ToSpecSchema(true, nil)
	assert.NoError(t, err)
	assert.Equal(t, "", propName)
	assert.Nil(t, schema)
	assert.False(t, required)
	assert.Nil(t, nestedTypes)
}

func TestExtractTypeParameter(t *testing.T) {
	tests := []struct {
		name    string
		typeStr string
		want    string
		wantErr bool
	}{
		{
			name:    "simple type",
			typeStr: "fields.StructField[User]",
			want:    "User",
		},
		{
			name:    "pointer type",
			typeStr: "fields.StructField[*User]",
			want:    "User",
		},
		{
			name:    "package qualified type",
			typeStr: "fields.StructField[*billing_plan.FeatureSet]",
			want:    "billing_plan.FeatureSet",
		},
		{
			name:    "array type",
			typeStr: "fields.StructField[[]User]",
			want:    "[]User",
		},
		{
			name:    "map type",
			typeStr: "fields.StructField[map[string]User]",
			want:    "map[string]User",
		},
		{
			name:    "complex nested type",
			typeStr: "fields.StructField[map[string][]User]",
			want:    "map[string][]User",
		},
		{
			name:    "invalid - no bracket",
			typeStr: "fields.StructField",
			wantErr: true,
		},
		{
			name:    "invalid - mismatched brackets",
			typeStr: "fields.StructField[map[string]User",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTypeParameter(tt.typeStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuildSchemaForType(t *testing.T) {
	tests := []struct {
		name       string
		typeStr    string
		public     bool
		wantType   []string
		wantRef    string
		wantNested []string
	}{
		{
			name:     "string",
			typeStr:  "string",
			public:   false,
			wantType: []string{"string"},
		},
		{
			name:     "int",
			typeStr:  "int",
			public:   false,
			wantType: []string{"integer"},
		},
		{
			name:       "struct without public",
			typeStr:    "User",
			public:     false,
			wantRef:    "#/definitions/User",
			wantNested: []string{"User"},
		},
		{
			name:       "struct with public",
			typeStr:    "User",
			public:     true,
			wantRef:    "#/definitions/UserPublic",
			wantNested: []string{"User"},
		},
		{
			name:       "package qualified struct",
			typeStr:    "billing_plan.FeatureSet",
			public:     true,
			wantRef:    "#/definitions/billing_plan.FeatureSetPublic",
			wantNested: []string{"billing_plan.FeatureSet"},
		},
		{
			name:     "array of strings",
			typeStr:  "[]string",
			public:   false,
			wantType: []string{"array"},
		},
		{
			name:       "array of structs",
			typeStr:    "[]User",
			public:     false,
			wantType:   []string{"array"},
			wantNested: []string{"User"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, nestedTypes, err := buildSchemaForType(tt.typeStr, tt.public, "", nil)
			assert.NoError(t, err)
			assert.NotNil(t, schema)

			if tt.wantRef != "" {
				assert.Equal(t, tt.wantRef, schema.Ref.String())
			} else if len(tt.wantType) > 0 {
				assert.Equal(t, len(tt.wantType), len(schema.Type))
				if len(tt.wantType) > 0 {
					assert.Equal(t, tt.wantType[0], schema.Type[0])
				}
			}

			if tt.wantNested != nil {
				assert.Equal(t, tt.wantNested, nestedTypes)
			}
		})
	}
}
