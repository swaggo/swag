package swag

import (
	"encoding/json"
	"go/ast"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sv-tools/openapi/spec"
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
			assert.Truef(t, (*s.Spec.Type)[0] == "string", "type sql.NullString should be overridden by string")
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

	p := New(GenerateOpenAPI3Doc(true))

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

	assert.Equal(t, 8, len(p.openAPI.Components.Spec.SecuritySchemes))

	security := p.openAPI.Components.Spec.SecuritySchemes
	if v, ok := security["basic"]; ok && v != nil && v.Spec != nil && v.Spec.Spec != nil {
		assert.Equal(t, "basic", v.Spec.Spec.Scheme)
		assert.Equal(t, "http", v.Spec.Spec.Type)
	}
	if v, ok := security["ApiKeyAuth"]; ok && v != nil && v.Spec != nil && v.Spec.Spec != nil {
		assert.Equal(t, "apiKey", v.Spec.Spec.Type)
		assert.Equal(t, "Authorization", v.Spec.Spec.Name)
		assert.Equal(t, "header", v.Spec.Spec.In)
		assert.Equal(t, "some description", v.Spec.Spec.Description)
	}
	if v, ok := security["OAuth2Application"]; ok && v != nil && v.Spec != nil && v.Spec.Spec != nil {
		assert.Equal(t, "oauth2", v.Spec.Spec.Type)
		assert.Equal(t, "header", v.Spec.Spec.In)
		assert.Equal(t, "https://example.com/oauth/token", v.Spec.Spec.Flows.Spec.ClientCredentials.Spec.TokenURL)
		assert.Equal(t, 2, len(v.Spec.Spec.Flows.Spec.ClientCredentials.Spec.Scopes))
	}
	if v, ok := security["OAuth2Implicit"]; ok && v != nil && v.Spec != nil && v.Spec.Spec != nil {
		assert.Equal(t, "oauth2", v.Spec.Spec.Type)
		assert.Equal(t, "header", v.Spec.Spec.In)
		assert.Equal(t, "https://example.com/oauth/authorize", v.Spec.Spec.Flows.Spec.Implicit.Spec.AuthorizationURL)
		assert.Equal(t, "some_audience.google.com", v.Spec.Spec.Flows.Extensions["x-google-audiences"])
	}
	if v, ok := security["OAuth2Password"]; ok && v != nil && v.Spec != nil && v.Spec.Spec != nil {
		assert.Equal(t, "oauth2", v.Spec.Spec.Type)
		assert.Equal(t, "header", v.Spec.Spec.In)
		assert.Equal(t, "https://example.com/oauth/token", v.Spec.Spec.Flows.Spec.Password.Spec.TokenURL)
	}
	if v, ok := security["OAuth2AccessCode"]; ok && v != nil && v.Spec != nil && v.Spec.Spec != nil {
		assert.Equal(t, "oauth2", v.Spec.Spec.Type)
		assert.Equal(t, "header", v.Spec.Spec.In)
		assert.Equal(t, "https://example.com/oauth/token", v.Spec.Spec.Flows.Spec.AuthorizationCode.Spec.TokenURL)
	}
	if v, ok := security["BearerAuth1"]; ok && v != nil && v.Spec != nil && v.Spec.Spec != nil {
		assert.Equal(t, "bearer", v.Spec.Spec.Scheme)
		assert.Equal(t, "http", v.Spec.Spec.Type)
		assert.Equal(t, "JWT", v.Spec.Spec.BearerFormat)
		assert.Equal(t, "First bearer token", v.Spec.Spec.Description)
	}
	if v, ok := security["BearerAuth2"]; ok && v != nil && v.Spec != nil && v.Spec.Spec != nil {
		assert.Equal(t, "bearer", v.Spec.Spec.Scheme)
		assert.Equal(t, "http", v.Spec.Spec.Type)
		assert.Equal(t, "CustomToken", v.Spec.Spec.BearerFormat)
		assert.Equal(t, "Second bearer token", v.Spec.Spec.Description)
	}
}

func TestParser_ParseGeneralApiInfoExtensionsV3(t *testing.T) {
	// should return an error because extension value is not a valid json
	t.Run("Test invalid extension value", func(t *testing.T) {
		t.Parallel()

		expected := "could not parse extension comment: annotation @x-google-endpoints need a valid json value. error: invalid character ':' after array element"
		gopath := os.Getenv("GOPATH")
		assert.NotNil(t, gopath)

		p := New(GenerateOpenAPI3Doc(true))

		err := p.ParseGeneralAPIInfo("testdata/v3/extensionsFail1.go")
		if assert.Error(t, err) {
			assert.Equal(t, expected, err.Error())
		}
	})

	// should return an error because extension don't have a value
	t.Run("Test missing extension value", func(t *testing.T) {
		t.Parallel()

		expected := "could not parse extension comment: annotation @x-google-endpoints need a value"
		gopath := os.Getenv("GOPATH")
		assert.NotNil(t, gopath)

		p := New(GenerateOpenAPI3Doc(true))

		err := p.ParseGeneralAPIInfo("testdata/v3/extensionsFail2.go")
		if assert.Error(t, err) {
			assert.Equal(t, expected, err.Error())
		}
	})
}

func TestParserParseGeneralApiInfoWithOpsInSameFileV3(t *testing.T) {
	t.Parallel()

	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)

	p := New(GenerateOpenAPI3Doc(true))

	err := p.ParseGeneralAPIInfo("testdata/single_file_api/main.go")
	assert.NoError(t, err)

	assert.Equal(t, "This is a sample server Petstore server.\nIt has a lot of beautiful features.", p.openAPI.Info.Spec.Description)
	assert.Equal(t, "Swagger Example API", p.openAPI.Info.Spec.Title)
	assert.Equal(t, "http://swagger.io/terms/", p.openAPI.Info.Spec.TermsOfService)
}

func TestParserParseGeneralAPIInfoMarkdownV3(t *testing.T) {
	t.Parallel()

	p := New(SetMarkdownFileDirectory("testdata"), GenerateOpenAPI3Doc(true))
	mainAPIFile := "testdata/markdown.go"
	err := p.ParseGeneralAPIInfo(mainAPIFile)
	assert.NoError(t, err)

	assert.Equal(t, "Swagger Example API Markdown Description", p.openAPI.Info.Spec.Description)
	assert.Equal(t, "users", p.openAPI.Tags[0].Spec.Name)
	assert.Equal(t, "Users Tag Markdown Description", p.openAPI.Tags[0].Spec.Description)

	p = New(GenerateOpenAPI3Doc(true))

	err = p.ParseGeneralAPIInfo(mainAPIFile)
	assert.Error(t, err)
}

func TestParserParseGeneralApiInfoFailedV3(t *testing.T) {
	t.Parallel()

	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)
	p := New(GenerateOpenAPI3Doc(true))
	assert.Error(t, p.ParseGeneralAPIInfo("testdata/noexist.go"))
}

func TestParserParseGeneralAPIInfoCollectionFormatV3(t *testing.T) {
	t.Parallel()

	parser := New(GenerateOpenAPI3Doc(true))
	assert.NoError(t, parser.parseGeneralAPIInfoV3([]string{
		"@query.collection.format csv",
	}))
	assert.Equal(t, parser.collectionFormatInQuery, "csv")

	assert.NoError(t, parser.parseGeneralAPIInfoV3([]string{
		"@query.collection.format tsv",
	}))
	assert.Equal(t, parser.collectionFormatInQuery, "tsv")
}

func TestParserParseGeneralAPITagGroupsV3(t *testing.T) {
	t.Parallel()

	parser := New(GenerateOpenAPI3Doc(true))
	assert.NoError(t, parser.parseGeneralAPIInfoV3([]string{
		"@x-tagGroups [{\"name\":\"General\",\"tags\":[\"lanes\",\"video-recommendations\"]}]",
	}))

	expected := []interface{}{map[string]interface{}{"name": "General", "tags": []interface{}{"lanes", "video-recommendations"}}}
	assert.Equal(t, expected, parser.openAPI.Info.Extensions["x-tagGroups"])
}

func TestParserParseGeneralAPITagDocsV3(t *testing.T) {
	t.Parallel()

	parser := New(GenerateOpenAPI3Doc(true))
	assert.Error(t, parser.parseGeneralAPIInfoV3([]string{
		"@tag.name Test",
		"@tag.docs.description Best example documentation"}))

	parser = New(GenerateOpenAPI3Doc(true))
	err := parser.parseGeneralAPIInfoV3([]string{
		"@tag.name test",
		"@tag.description A test Tag",
		"@tag.docs.url https://example.com",
		"@tag.docs.description Best example documentation"})
	assert.NoError(t, err)

	assert.Equal(t, "test", parser.openAPI.Tags[0].Spec.Name)
	assert.Equal(t, "A test Tag", parser.openAPI.Tags[0].Spec.Description)
	assert.Equal(t, "https://example.com", parser.openAPI.Tags[0].Spec.ExternalDocs.Spec.URL)
	assert.Equal(t, "Best example documentation", parser.openAPI.Tags[0].Spec.ExternalDocs.Spec.Description)
}

func TestGetAllGoFileInfoV3(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/pet"

	p := New(GenerateOpenAPI3Doc(true))
	err := p.getAllGoFileInfo("testdata", searchDir)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(p.packages.files))
}

func TestParser_ParseTypeV3(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/v3/simple/"

	p := New(GenerateOpenAPI3Doc(true))
	err := p.getAllGoFileInfo("testdata", searchDir)
	assert.NoError(t, err)

	_, err = p.packages.ParseTypes()

	assert.NoError(t, err)
	assert.NotNil(t, p.packages.uniqueDefinitions["api.Pet3"])
	assert.NotNil(t, p.packages.uniqueDefinitions["web.Pet"])
	assert.NotNil(t, p.packages.uniqueDefinitions["web.Pet2"])
}

func TestParsePet(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/v3/pet"

	p := New(GenerateOpenAPI3Doc(true))
	p.PropNamingStrategy = PascalCase

	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	schemas := p.openAPI.Components.Spec.Schemas
	assert.NotNil(t, schemas)

	tagSchema := schemas["web.Tag"].Spec
	assert.Equal(t, 2, len(tagSchema.Properties))
	assert.Equal(t, &typeInteger, tagSchema.Properties["id"].Spec.Type)
	assert.Equal(t, &typeString, tagSchema.Properties["name"].Spec.Type)

	petSchema := schemas["web.Pet"].Spec
	assert.NotNil(t, petSchema)
	assert.Equal(t, 8, len(petSchema.Properties))
	assert.Equal(t, &typeInteger, petSchema.Properties["id"].Spec.Type)
	assert.Equal(t, &typeString, petSchema.Properties["name"].Spec.Type)

}

func TestParseSimpleApiV3(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/v3/simple"
	p := New(GenerateOpenAPI3Doc(true))
	p.PropNamingStrategy = PascalCase

	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	paths := p.openAPI.Paths.Spec.Paths

	path := paths["/testapi/get-string-by-int/{some_id}"].Spec.Spec.Get.Spec
	assert.Equal(t, "get string by ID", path.Description)
	assert.Equal(t, "Add a new pet to the store", path.Summary)
	assert.Equal(t, "get-string-by-int", path.OperationID)

	response := path.Responses.Spec.Response["200"]
	assert.Equal(t, "ok", response.Spec.Spec.Description)

	path = paths["/FormData"].Spec.Spec.Post.Spec
	assert.NotNil(t, path)
	assert.NotNil(t, path.RequestBody)
	//TODO add asserts

	t.Run("Test parse struct oneOf", func(t *testing.T) {
		t.Parallel()

		assert.Contains(t, p.openAPI.Components.Spec.Schemas, "web.OneOfTest")
		schema := p.openAPI.Components.Spec.Schemas["web.OneOfTest"].Spec
		expected := `{
    "properties": {
        "big_int": {
            "oneOf": [
                {
                    "type": "string"
                },
                {
                    "type": "integer"
                }
            ]
        },
        "pet_detail": {
            "oneOf": [
                {
                    "$ref": "#/components/schemas/web.Cat"
                },
                {
                    "$ref": "#/components/schemas/web.Dog"
                }
            ]
        }
    },
    "type": "object"
}`
		out, err := json.MarshalIndent(schema, "", "    ")
		assert.NoError(t, err)
		assert.Equal(t, expected, string(out))

		assert.Contains(t, p.openAPI.Components.Spec.Schemas, "web.Cat")
		schema = p.openAPI.Components.Spec.Schemas["web.Cat"].Spec
		expected = `{
    "properties": {
        "age": {
            "type": "integer"
        },
        "hunts": {
            "type": "boolean"
        }
    },
    "type": "object"
}`
		out, err = json.MarshalIndent(schema, "", "    ")
		assert.NoError(t, err)
		assert.Equal(t, expected, string(out))

		assert.Contains(t, p.openAPI.Components.Spec.Schemas, "web.Dog")
		schema = p.openAPI.Components.Spec.Schemas["web.Dog"].Spec
		expected = `{
    "properties": {
        "bark": {
            "type": "boolean"
        },
        "breed": {
            "enum": [
                "Dingo",
                "Husky",
                "Retriever",
                "Shepherd"
            ],
            "type": "string"
        }
    },
    "type": "object"
}`
		out, err = json.MarshalIndent(schema, "", "    ")
		assert.NoError(t, err)
		assert.Equal(t, expected, string(out))
	})

	t.Run("Test parse response oneOf", func(t *testing.T) {
		t.Parallel()

		assert.Contains(t, paths, "/pets/{id}")
		path := paths["/pets/{id}"]
		assert.Contains(t, path.Spec.Spec.Get.Spec.Responses.Spec.Response, "200")
		response = path.Spec.Spec.Get.Spec.Responses.Spec.Response["200"]
		assert.Equal(t, "Return Cat or Dog", response.Spec.Spec.Description)
		mediaType := response.Spec.Spec.Content["application/json"]
		rootSchema := mediaType.Spec.Schema.Spec
		assert.Equal(t, []*spec.RefOrSpec[spec.Schema]{
			{Ref: &spec.Ref{Ref: "#/components/schemas/web.Cat"}},
			{Ref: &spec.Ref{Ref: "#/components/schemas/web.Dog"}},
		}, rootSchema.OneOf)

	})
}

func TestParserParseServers(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/v3/servers"
	p := New(GenerateOpenAPI3Doc(true))
	p.PropNamingStrategy = PascalCase

	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	servers := p.openAPI.Servers
	require.NotNil(t, servers)

	assert.Equal(t, 2, len(servers))
	assert.Equal(t, "{scheme}://{host}:{port}", servers[0].Spec.URL)
	assert.Equal(t, "Test Petstore server.", servers[0].Spec.Description)

	assert.Equal(t, "https", servers[0].Spec.Variables["scheme"].Spec.Default)
	assert.Equal(t, []string{"http", "https"}, servers[0].Spec.Variables["scheme"].Spec.Enum)
	assert.Equal(t, "test.petstore.com", servers[0].Spec.Variables["host"].Spec.Default)
	assert.Equal(t, "443", servers[0].Spec.Variables["port"].Spec.Default)

	assert.Equal(t, "https://petstore.com/v3", servers[1].Spec.URL)
	assert.Equal(t, "Production Petstore server.", servers[1].Spec.Description)

}

func TestParserParseGeneralAPIInfoGlobalSecurityV3(t *testing.T) {
	t.Parallel()

	// Test simple global security
	parser := New(GenerateOpenAPI3Doc(true))
	err := parser.parseGeneralAPIInfoV3([]string{
		"@security ApiKeyAuth",
	})
	assert.NoError(t, err)
	assert.Len(t, parser.openAPI.Security, 1)
	assert.Contains(t, parser.openAPI.Security[0], "ApiKeyAuth")
	assert.Equal(t, []string{}, parser.openAPI.Security[0]["ApiKeyAuth"])

	// Test OAuth2 with scopes
	parser2 := New(GenerateOpenAPI3Doc(true))
	err2 := parser2.parseGeneralAPIInfoV3([]string{
		"@security OAuth2Implicit[read,write]",
	})
	assert.NoError(t, err2)
	assert.Len(t, parser2.openAPI.Security, 1)
	assert.Contains(t, parser2.openAPI.Security[0], "OAuth2Implicit")
	assert.Equal(t, []string{"read", "write"}, parser2.openAPI.Security[0]["OAuth2Implicit"])

	// Test OR logic
	parser3 := New(GenerateOpenAPI3Doc(true))
	err3 := parser3.parseGeneralAPIInfoV3([]string{
		"@security ApiKeyAuth || BasicAuth",
	})
	assert.NoError(t, err3)
	assert.Len(t, parser3.openAPI.Security, 1)
	assert.Contains(t, parser3.openAPI.Security[0], "ApiKeyAuth")
	assert.Contains(t, parser3.openAPI.Security[0], "BasicAuth")
	assert.Equal(t, []string{}, parser3.openAPI.Security[0]["ApiKeyAuth"])
	assert.Equal(t, []string{}, parser3.openAPI.Security[0]["BasicAuth"])

	// Test AND logic (multiple @security lines)
	parser4 := New(GenerateOpenAPI3Doc(true))
	err4 := parser4.parseGeneralAPIInfoV3([]string{
		"@security ApiKeyAuth",
		"@security BasicAuth",
	})
	assert.NoError(t, err4)
	assert.Len(t, parser4.openAPI.Security, 2)
	assert.Contains(t, parser4.openAPI.Security[0], "ApiKeyAuth")
	assert.Contains(t, parser4.openAPI.Security[1], "BasicAuth")
}

func TestParseTypeAlias(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/v3/type_alias_definition"

	p := New(GenerateOpenAPI3Doc(true))

	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join(searchDir, "expected.json"))
	require.NoError(t, err)

	result, err := json.Marshal(p.openAPI)
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(result))
}

func TestParseInterface(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/v3/interface"

	p := New(GenerateOpenAPI3Doc(true))

	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join(searchDir, "expected.json"))
	require.NoError(t, err)

	result, err := json.Marshal(p.openAPI)
	require.NoError(t, err)

	assert.JSONEq(t, string(expected), string(result))
}

func TestParseRecursionWithSchemaName(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/recursion_schema_name"
	p := New(GenerateOpenAPI3Doc(true))

	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	require.NoError(t, err)

	userSchema, exists := p.openAPI.Components.Spec.Schemas["User"]
	require.True(t, exists, "User schema should exist")
	require.NotNil(t, userSchema, "User schema should not be nil")
	require.NotNil(t, userSchema.Spec, "User schema spec should not be nil")

	assert.Equal(t, "object", (*userSchema.Spec.Type)[0])

	childrenProp, exists := userSchema.Spec.Properties["children"]
	require.True(t, exists, "children property should exist")
	require.NotNil(t, childrenProp.Spec, "children property spec should not be nil")

	assert.Equal(t, "array", (*childrenProp.Spec.Type)[0])

	require.NotNil(t, childrenProp.Spec.Items, "children items should not be nil")
	require.NotNil(t, childrenProp.Spec.Items.Schema, "children items schema should not be nil")

	expectedRef := "#/components/schemas/User"
	assert.Equal(t, expectedRef, childrenProp.Spec.Items.Schema.Ref.Ref)
}

func TestGetSchemaByRef(t *testing.T) {
	t.Parallel()

	p := New(GenerateOpenAPI3Doc(true))
	p.openAPI.Components.Spec.Schemas = make(map[string]*spec.RefOrSpec[spec.Schema])

	t.Run("Existing schema", func(t *testing.T) {
		testSchema := &spec.Schema{}
		testSchema.Type = &spec.SingleOrArray[string]{"string"}
		p.openAPI.Components.Spec.Schemas["TestSchema"] = spec.NewRefOrSpec(nil, testSchema)

		ref := &spec.Ref{Ref: "#/components/schemas/TestSchema"}
		result := p.getSchemaByRef(ref)

		require.NotNil(t, result)
		assert.Equal(t, testSchema, result)
	})

	t.Run("Non-existing schema returns empty schema", func(t *testing.T) {
		ref := &spec.Ref{Ref: "#/components/schemas/NonExistentSchema"}
		result := p.getSchemaByRef(ref)

		require.NotNil(t, result)
		assert.Equal(t, &spec.Schema{}, result)
	})
}
