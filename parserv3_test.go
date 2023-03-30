package swag

import (
	"go/ast"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverridesGetTypeSchemaV3(t *testing.T) {
	t.Parallel()

	overrides := map[string]string{
		"sql.NullString": "string",
	}

	p := New(SetOverrides(overrides))

	t.Run("Override sql.NullString by string", func(t *testing.T) {
		t.Parallel()

		s, err := p.getTypeSchemaV3("sql.NullString", nil, false)
		if assert.NoError(t, err) {
			assert.Truef(t, s.Spec.Type[0] == "string", "type sql.NullString should be overridden by string")
		}
	})

	t.Run("Missing Override for sql.NullInt64", func(t *testing.T) {
		t.Parallel()

		_, err := p.getTypeSchemaV3("sql.NullInt64", nil, false)
		if assert.Error(t, err) {
			assert.Equal(t, "cannot find type definition: sql.NullInt64", err.Error())
		}
	})
}

func TestParserParseDefinitionV3(t *testing.T) {
	p := New()

	// Parsing existing type
	definition := &TypeSpecDef{
		PkgPath: "github.com/swagger/swag",
		File: &ast.File{
			Name: &ast.Ident{
				Name: "swag",
			},
		},
		TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{
				Name: "Test",
			},
		},
	}

	expected := &SchemaV3{}
	p.parsedSchemasV3[definition] = expected

	schema, err := p.ParseDefinitionV3(definition)
	assert.NoError(t, err)
	assert.Equal(t, expected, schema)

	// Parsing *ast.FuncType
	definition = &TypeSpecDef{
		PkgPath: "github.com/swagger/swag/model",
		File: &ast.File{
			Name: &ast.Ident{
				Name: "model",
			},
		},
		TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{
				Name: "Test",
			},
			Type: &ast.FuncType{},
		},
	}
	_, err = p.ParseDefinitionV3(definition)
	assert.Error(t, err)

	// Parsing *ast.FuncType with parent spec
	definition = &TypeSpecDef{
		PkgPath: "github.com/swagger/swag/model",
		File: &ast.File{
			Name: &ast.Ident{
				Name: "model",
			},
		},
		TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{
				Name: "Test",
			},
			Type: &ast.FuncType{},
		},
		ParentSpec: &ast.FuncDecl{
			Name: ast.NewIdent("TestFuncDecl"),
		},
	}
	_, err = p.ParseDefinitionV3(definition)
	assert.Error(t, err)
	assert.Equal(t, "model.TestFuncDecl.Test", definition.TypeName())
}

func TestParserParseGeneralApiInfoV3(t *testing.T) {
	t.Parallel()

	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)

	p := New(SetOpenAPIVersion(true))

	err := p.ParseGeneralAPIInfo("testdata/v3/main.go")
	assert.NoError(t, err)

	assert.Equal(t, "This is a sample server Petstore server.\nIt has a lot of beautiful features.", p.openAPI.Info.Spec.Description)
	assert.Equal(t, "Swagger Example API", p.openAPI.Info.Spec.Title)
	assert.Equal(t, "http://swagger.io/terms/", p.openAPI.Info.Spec.TermsOfService)
	assert.Equal(t, "API Support", p.openAPI.Info.Spec.Contact.Spec.Name)
	assert.Equal(t, "http://www.swagger.io/support", p.openAPI.Info.Spec.Contact.Spec.URL)
	assert.Equal(t, "support@swagger.io", p.openAPI.Info.Spec.Contact.Spec.Email)
	assert.Equal(t, "Apache 2.0", p.openAPI.Info.Spec.License.Spec.Name)
	assert.Equal(t, "http://www.apache.org/licenses/LICENSE-2.0.html", p.openAPI.Info.Spec.License.Spec.URL)
	assert.Equal(t, "1.0", p.openAPI.Info.Spec.Version)

	xLogo := map[string]interface{}(map[string]interface{}{"altText": "Petstore logo", "backgroundColor": "#FFFFFF", "url": "https://redocly.github.io/redoc/petstore-logo.png"})
	assert.Equal(t, xLogo, p.openAPI.Info.Extensions["x-logo"])
	assert.Equal(t, "marks values", p.openAPI.Info.Extensions["x-google-marks"])

	endpoints := interface{}([]interface{}{map[string]interface{}{"allowCors": true, "name": "name.endpoints.environment.cloud.goog"}})
	assert.Equal(t, endpoints, p.openAPI.Info.Extensions["x-google-endpoints"])

	assert.Equal(t, "OpenAPI", p.openAPI.ExternalDocs.Spec.Description)
	assert.Equal(t, "https://swagger.io/resources/open-api", p.openAPI.ExternalDocs.Spec.URL)

	assert.Equal(t, 6, len(p.openAPI.Components.Spec.SecuritySchemes))
}
