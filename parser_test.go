package swag

import (
	"bytes"
	"encoding/json"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

const defaultParseDepth = 100

const mainAPIFile = "main.go"

func TestNew(t *testing.T) {
	New()
}

func TestSetMarkdownFileDirectory(t *testing.T) {
	t.Parallel()

	expected := "docs/markdown"
	p := New(SetMarkdownFileDirectory(expected))
	assert.Equal(t, expected, p.markdownFileDir)
}

func TestSetCodeExamplesDirectory(t *testing.T) {
	t.Parallel()

	expected := "docs/examples"
	p := New(SetCodeExamplesDirectory(expected))
	assert.Equal(t, expected, p.codeExampleFilesDir)
}

func TestSetStrict(t *testing.T) {
	t.Parallel()

	p := New()
	assert.Equal(t, false, p.Strict)

	p = New(SetStrict(true))
	assert.Equal(t, true, p.Strict)
}

func TestSetDebugger(t *testing.T) {
	t.Parallel()

	logger := log.New(&bytes.Buffer{}, "", log.LstdFlags)

	p := New(SetDebugger(logger))
	assert.Equal(t, p.debug, logger)
}

func TestParser_ParseDefinition(t *testing.T) {
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

	expected := &Schema{}
	p.parsedSchemas[definition] = expected

	schema, err := p.ParseDefinition(definition)
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
	_, err = p.ParseDefinition(definition)
	assert.Error(t, err)
}

func TestParser_ParseGeneralApiInfo(t *testing.T) {
	t.Parallel()

	expected := `{
    "schemes": [
        "http",
        "https"
    ],
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.\nIt has a lot of beautiful features.",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0",
        "x-logo": {
            "altText": "Petstore logo",
            "backgroundColor": "#FFFFFF",
            "url": "https://redocly.github.io/redoc/petstore-logo.png"
        }
    },
    "host": "petstore.swagger.io",
    "basePath": "/v2",
    "paths": {},
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        },
        "BasicAuth": {
            "type": "basic"
        },
        "OAuth2AccessCode": {
            "type": "oauth2",
            "flow": "accessCode",
            "authorizationUrl": "https://example.com/oauth/authorize",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information"
            },
            "x-tokenname": "id_token"
        },
        "OAuth2Application": {
            "type": "oauth2",
            "flow": "application",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "write": " Grants write access"
            }
        },
        "OAuth2Implicit": {
            "type": "oauth2",
            "flow": "implicit",
            "authorizationUrl": "https://example.com/oauth/authorize",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "write": " Grants write access"
            },
            "x-google-audiences": "some_audience.google.com"
        },
        "OAuth2Password": {
            "type": "oauth2",
            "flow": "password",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "read": " Grants read access",
                "write": " Grants write access"
            }
        }
    },
    "x-google-endpoints": [
        {
            "allowCors": true,
            "name": "name.endpoints.environment.cloud.goog"
        }
    ],
    "x-google-marks": "marks values"
}`
	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)
	p := New()
	err := p.ParseGeneralAPIInfo("testdata/main.go")
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParser_ParseGeneralApiInfoTemplated(t *testing.T) {
	t.Parallel()

	expected := `{
    "swagger": "2.0",
    "info": {
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        }
    },
    "paths": {},
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        },
        "BasicAuth": {
            "type": "basic"
        },
        "OAuth2AccessCode": {
            "type": "oauth2",
            "flow": "accessCode",
            "authorizationUrl": "https://example.com/oauth/authorize",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information"
            }
        },
        "OAuth2Application": {
            "type": "oauth2",
            "flow": "application",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "write": " Grants write access"
            }
        },
        "OAuth2Implicit": {
            "type": "oauth2",
            "flow": "implicit",
            "authorizationUrl": "https://example.com/oauth/authorize",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "write": " Grants write access"
            }
        },
        "OAuth2Password": {
            "type": "oauth2",
            "flow": "password",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "read": " Grants read access",
                "write": " Grants write access"
            }
        }
    },
    "x-google-endpoints": [
        {
            "allowCors": true,
            "name": "name.endpoints.environment.cloud.goog"
        }
    ],
    "x-google-marks": "marks values"
}`
	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)
	p := New()
	err := p.ParseGeneralAPIInfo("testdata/templated.go")
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParser_ParseGeneralApiInfoExtensions(t *testing.T) {
	// should return an error because extension value is not a valid json
	t.Run("Test invalid extension value", func(t *testing.T) {
		t.Parallel()

		expected := "annotation @x-google-endpoints need a valid json value"
		gopath := os.Getenv("GOPATH")
		assert.NotNil(t, gopath)
		p := New()
		err := p.ParseGeneralAPIInfo("testdata/extensionsFail1.go")
		if assert.Error(t, err) {
			assert.Equal(t, expected, err.Error())
		}
	})

	// should return an error because extension don't have a value
	t.Run("Test missing extension value", func(t *testing.T) {
		t.Parallel()

		expected := "annotation @x-google-endpoints need a value"
		gopath := os.Getenv("GOPATH")
		assert.NotNil(t, gopath)
		p := New()
		err := p.ParseGeneralAPIInfo("testdata/extensionsFail2.go")
		if assert.Error(t, err) {
			assert.Equal(t, expected, err.Error())
		}
	})
}

func TestParser_ParseGeneralApiInfoWithOpsInSameFile(t *testing.T) {
	t.Parallel()

	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.\nIt has a lot of beautiful features.",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {},
        "version": "1.0"
    },
    "paths": {}
}`

	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)
	p := New()
	err := p.ParseGeneralAPIInfo("testdata/single_file_api/main.go")
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParser_ParseGeneralAPIInfoMarkdown(t *testing.T) {
	t.Parallel()

	p := New(SetMarkdownFileDirectory("testdata"))
	mainAPIFile := "testdata/markdown.go"
	err := p.ParseGeneralAPIInfo(mainAPIFile)
	assert.NoError(t, err)

	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "Swagger Example API Markdown Description",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {},
        "version": "1.0"
    },
    "paths": {},
    "tags": [
        {
            "description": "Users Tag Markdown Description",
            "name": "users"
        }
    ]
}`
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))

	p = New()
	err = p.ParseGeneralAPIInfo(mainAPIFile)
	assert.Error(t, err)
}

func TestParser_ParseGeneralApiInfoFailed(t *testing.T) {
	t.Parallel()

	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)
	p := New()
	assert.Error(t, p.ParseGeneralAPIInfo("testdata/noexist.go"))
}

func TestParser_ParseAcceptComment(t *testing.T) {
	t.Parallel()

	expected := []string{
		"application/json",
		"text/xml",
		"text/plain",
		"text/html",
		"multipart/form-data",
		"application/x-www-form-urlencoded",
		"application/vnd.api+json",
		"application/x-json-stream",
		"application/octet-stream",
		"image/png",
		"image/jpeg",
		"image/gif",
		"application/xhtml+xml",
		"application/health+json",
	}

	comment := `@Accept json,xml,plain,html,mpfd,x-www-form-urlencoded,json-api,json-stream,octet-stream,png,jpeg,gif,application/xhtml+xml,application/health+json`

	parser := New()
	assert.NoError(t, parseGeneralAPIInfo(parser, []string{comment}))
	assert.Equal(t, parser.swagger.Consumes, expected)

	assert.Error(t, parseGeneralAPIInfo(parser, []string{`@Accept cookies,candies`}))

	parser = New()
	assert.NoError(t, parser.ParseAcceptComment(comment[len(acceptAttr)+1:]))
	assert.Equal(t, parser.swagger.Consumes, expected)
}

func TestParser_ParseProduceComment(t *testing.T) {
	t.Parallel()

	expected := []string{
		"application/json",
		"text/xml",
		"text/plain",
		"text/html",
		"multipart/form-data",
		"application/x-www-form-urlencoded",
		"application/vnd.api+json",
		"application/x-json-stream",
		"application/octet-stream",
		"image/png",
		"image/jpeg",
		"image/gif",
		"application/xhtml+xml",
		"application/health+json",
	}

	comment := `@Produce json,xml,plain,html,mpfd,x-www-form-urlencoded,json-api,json-stream,octet-stream,png,jpeg,gif,application/xhtml+xml,application/health+json`

	parser := New()
	assert.NoError(t, parseGeneralAPIInfo(parser, []string{comment}))
	assert.Equal(t, parser.swagger.Produces, expected)

	assert.Error(t, parseGeneralAPIInfo(parser, []string{`@Produce cookies,candies`}))

	parser = New()
	assert.NoError(t, parser.ParseProduceComment(comment[len(produceAttr)+1:]))
	assert.Equal(t, parser.swagger.Produces, expected)
}

func TestParser_ParseGeneralAPIInfoCollectionFromat(t *testing.T) {
	t.Parallel()

	parser := New()
	assert.NoError(t, parseGeneralAPIInfo(parser, []string{
		"@query.collection.format csv",
	}))
	assert.Equal(t, parser.collectionFormatInQuery, "csv")

	assert.NoError(t, parseGeneralAPIInfo(parser, []string{
		"@query.collection.format tsv",
	}))
	assert.Equal(t, parser.collectionFormatInQuery, "tsv")
}

func TestParser_ParseGeneralAPITagGroups(t *testing.T) {
	t.Parallel()

	parser := New()
	assert.NoError(t, parseGeneralAPIInfo(parser, []string{
		"@x-tagGroups [{\"name\":\"General\",\"tags\":[\"lanes\",\"video-recommendations\"]}]",
	}))

	expected := []interface{}{map[string]interface{}{"name": "General", "tags": []interface{}{"lanes", "video-recommendations"}}}
	assert.Equal(t, parser.swagger.Extensions["x-tagGroups"], expected)

}

func TestParser_ParseGeneralAPITagDocs(t *testing.T) {
	t.Parallel()

	parser := New()
	assert.Error(t, parseGeneralAPIInfo(parser, []string{
		"@tag.name Test",
		"@tag.docs.description Best example documentation"}))

	parser = New()
	err := parseGeneralAPIInfo(parser, []string{
		"@tag.name test",
		"@tag.description A test Tag",
		"@tag.docs.url https://example.com",
		"@tag.docs.description Best example documentation"})
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(parser.GetSwagger().Tags, "", "    ")
	expected := `[
    {
        "description": "A test Tag",
        "name": "test",
        "externalDocs": {
            "description": "Best example documentation",
            "url": "https://example.com"
        }
    }
]`
	assert.Equal(t, expected, string(b))

}

func TestParser_ParseGeneralAPISecurity(t *testing.T) {
	t.Run("ApiKey", func(t *testing.T) {
		t.Parallel()

		parser := New()
		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.apikey ApiKey"}))

		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.apikey ApiKey",
			"@in header"}))
		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.apikey ApiKey",
			"@name X-API-KEY"}))

		err := parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.apikey ApiKey",
			"@in header",
			"@name X-API-KEY"})
		assert.NoError(t, err)

		b, _ := json.MarshalIndent(parser.GetSwagger().SecurityDefinitions, "", "    ")
		expected := `{
    "ApiKey": {
        "type": "apiKey",
        "name": "X-API-KEY",
        "in": "header"
    }
}`
		assert.Equal(t, expected, string(b))
	})

	t.Run("OAuth2Application", func(t *testing.T) {
		t.Parallel()

		parser := New()
		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.application OAuth2Application"}))

		err := parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.application OAuth2Application",
			"@tokenUrl https://example.com/oauth/token"})
		assert.NoError(t, err)
		b, _ := json.MarshalIndent(parser.GetSwagger().SecurityDefinitions, "", "    ")
		expected := `{
    "OAuth2Application": {
        "type": "oauth2",
        "flow": "application",
        "tokenUrl": "https://example.com/oauth/token"
    }
}`
		assert.Equal(t, expected, string(b))
	})

	t.Run("OAuth2Implicit", func(t *testing.T) {
		t.Parallel()

		parser := New()
		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.implicit OAuth2Implicit"}))

		err := parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.implicit OAuth2Implicit",
			"@authorizationurl https://example.com/oauth/authorize"})
		assert.NoError(t, err)
		b, _ := json.MarshalIndent(parser.GetSwagger().SecurityDefinitions, "", "    ")
		expected := `{
    "OAuth2Implicit": {
        "type": "oauth2",
        "flow": "implicit",
        "authorizationUrl": "https://example.com/oauth/authorize"
    }
}`
		assert.Equal(t, expected, string(b))
	})

	t.Run("OAuth2Password", func(t *testing.T) {
		t.Parallel()

		parser := New()
		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.password OAuth2Password"}))

		err := parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.password OAuth2Password",
			"@tokenUrl https://example.com/oauth/token"})
		assert.NoError(t, err)
		b, _ := json.MarshalIndent(parser.GetSwagger().SecurityDefinitions, "", "    ")
		expected := `{
    "OAuth2Password": {
        "type": "oauth2",
        "flow": "password",
        "tokenUrl": "https://example.com/oauth/token"
    }
}`
		assert.Equal(t, expected, string(b))
	})

	t.Run("OAuth2AccessCode", func(t *testing.T) {
		t.Parallel()

		parser := New()
		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.accessCode OAuth2AccessCode"}))

		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.accessCode OAuth2AccessCode",
			"@tokenUrl https://example.com/oauth/token"}))

		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.accessCode OAuth2AccessCode",
			"@authorizationurl https://example.com/oauth/authorize"}))

		err := parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.accessCode OAuth2AccessCode",
			"@tokenUrl https://example.com/oauth/token",
			"@authorizationurl https://example.com/oauth/authorize"})
		assert.NoError(t, err)
		b, _ := json.MarshalIndent(parser.GetSwagger().SecurityDefinitions, "", "    ")
		expected := `{
    "OAuth2AccessCode": {
        "type": "oauth2",
        "flow": "accessCode",
        "authorizationUrl": "https://example.com/oauth/authorize",
        "tokenUrl": "https://example.com/oauth/token"
    }
}`
		assert.Equal(t, expected, string(b))

		assert.Error(t, parseGeneralAPIInfo(parser, []string{
			"@securitydefinitions.oauth2.accessCode OAuth2AccessCode",
			"@tokenUrl https://example.com/oauth/token",
			"@authorizationurl https://example.com/oauth/authorize",
			"@scope.read,write Multiple scope"}))
	})

}

func TestGetAllGoFileInfo(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/pet"

	p := New()
	err := p.getAllGoFileInfo("testdata", searchDir)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(p.packages.files))
}

func TestParser_ParseType(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/simple/"

	p := New()
	err := p.getAllGoFileInfo("testdata", searchDir)
	assert.NoError(t, err)

	_, err = p.packages.ParseTypes()

	assert.NoError(t, err)
	assert.NotNil(t, p.packages.uniqueDefinitions["api.Pet3"])
	assert.NotNil(t, p.packages.uniqueDefinitions["web.Pet"])
	assert.NotNil(t, p.packages.uniqueDefinitions["web.Pet2"])
}

func TestGetSchemes(t *testing.T) {
	t.Parallel()

	schemes := getSchemes("@schemes http https")
	expectedSchemes := []string{"http", "https"}
	assert.Equal(t, expectedSchemes, schemes)
}

func TestParseSimpleApi1(t *testing.T) {
	t.Parallel()

	expected, err := ioutil.ReadFile("testdata/simple/expected.json")
	assert.NoError(t, err)
	searchDir := "testdata/simple"
	p := New()
	p.PropNamingStrategy = PascalCase
	err = p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "  ")
	assert.Equal(t, string(expected), string(b))
}

func TestParseSimpleApi_ForSnakecase(t *testing.T) {
	t.Parallel()

	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0"
    },
    "host": "petstore.swagger.io",
    "basePath": "/v2",
    "paths": {
        "/file/upload": {
            "post": {
                "description": "Upload file",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Upload file",
                "operationId": "file.upload",
                "parameters": [
                    {
                        "type": "file",
                        "description": "this is a test file",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "We need ID!!",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    }
                }
            }
        },
        "/testapi/get-string-by-int/{some_id}": {
            "get": {
                "description": "get string by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Add a new pet to the store",
                "operationId": "get-string-by-int",
                "parameters": [
                    {
                        "type": "integer",
                        "format": "int64",
                        "description": "Some ID",
                        "name": "some_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Some ID",
                        "name": "some_id",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.Pet"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "We need ID!!",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    }
                }
            }
        },
        "/testapi/get-struct-array-by-string/{some_id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    },
                    {
                        "BasicAuth": []
                    },
                    {
                        "OAuth2Application": [
                            "write"
                        ]
                    },
                    {
                        "OAuth2Implicit": [
                            "read",
                            "admin"
                        ]
                    },
                    {
                        "OAuth2AccessCode": [
                            "read"
                        ]
                    },
                    {
                        "OAuth2Password": [
                            "admin"
                        ]
                    }
                ],
                "description": "get struct array by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "operationId": "get-struct-array-by-string",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Some ID",
                        "name": "some_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "enum": [
                            1,
                            2,
                            3
                        ],
                        "type": "integer",
                        "description": "Category",
                        "name": "category",
                        "in": "query",
                        "required": true
                    },
                    {
                        "minimum": 0,
                        "type": "integer",
                        "default": 0,
                        "description": "Offset",
                        "name": "offset",
                        "in": "query",
                        "required": true
                    },
                    {
                        "maximum": 50,
                        "type": "integer",
                        "default": 10,
                        "description": "Limit",
                        "name": "limit",
                        "in": "query",
                        "required": true
                    },
                    {
                        "maxLength": 50,
                        "minLength": 1,
                        "type": "string",
                        "default": "\"\"",
                        "description": "q",
                        "name": "q",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "We need ID!!",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "web.APIError": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "error_code": {
                    "type": "integer"
                },
                "error_message": {
                    "type": "string"
                }
            }
        },
        "web.Pet": {
            "type": "object",
            "required": [
                "price"
            ],
            "properties": {
                "birthday": {
                    "type": "integer"
                },
                "category": {
                    "type": "object",
                    "properties": {
                        "id": {
                            "type": "integer",
                            "example": 1
                        },
                        "name": {
                            "type": "string",
                            "example": "category_name"
                        },
                        "photo_urls": {
                            "type": "array",
                            "items": {
                                "type": "string",
                                "format": "url"
                            },
                            "example": [
                                "http://test/image/1.jpg",
                                "http://test/image/2.jpg"
                            ]
                        },
                        "small_category": {
                            "type": "object",
                            "required": [
                                "name"
                            ],
                            "properties": {
                                "id": {
                                    "type": "integer",
                                    "example": 1
                                },
                                "name": {
                                    "type": "string",
                                    "example": "detail_category_name"
                                },
                                "photo_urls": {
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    },
                                    "example": [
                                        "http://test/image/1.jpg",
                                        "http://test/image/2.jpg"
                                    ]
                                }
                            }
                        }
                    }
                },
                "coeffs": {
                    "type": "array",
                    "items": {
                        "type": "number"
                    }
                },
                "custom_string": {
                    "type": "string"
                },
                "custom_string_arr": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "data": {},
                "decimal": {
                    "type": "number"
                },
                "id": {
                    "type": "integer",
                    "format": "int64",
                    "example": 1
                },
                "is_alive": {
                    "type": "boolean",
                    "example": true
                },
                "name": {
                    "type": "string",
                    "example": "poti"
                },
                "null_int": {
                    "type": "integer"
                },
                "pets": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.Pet2"
                    }
                },
                "pets2": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.Pet2"
                    }
                },
                "photo_urls": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "http://test/image/1.jpg",
                        "http://test/image/2.jpg"
                    ]
                },
                "price": {
                    "type": "number",
                    "multipleOf": 0.01,
                    "example": 3.25
                },
                "status": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.Tag"
                    }
                },
                "uuid": {
                    "type": "string"
                }
            }
        },
        "web.Pet2": {
            "type": "object",
            "properties": {
                "deleted_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "middle_name": {
                    "type": "string"
                }
            }
        },
        "web.RevValue": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "integer"
                },
                "err": {
                    "type": "integer"
                },
                "status": {
                    "type": "boolean"
                }
            }
        },
        "web.Tag": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "format": "int64"
                },
                "name": {
                    "type": "string"
                },
                "pets": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.Pet"
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        },
        "BasicAuth": {
            "type": "basic"
        },
        "OAuth2AccessCode": {
            "type": "oauth2",
            "flow": "accessCode",
            "authorizationUrl": "https://example.com/oauth/authorize",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information"
            }
        },
        "OAuth2Application": {
            "type": "oauth2",
            "flow": "application",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "write": " Grants write access"
            }
        },
        "OAuth2Implicit": {
            "type": "oauth2",
            "flow": "implicit",
            "authorizationUrl": "https://example.com/oauth/authorize",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "write": " Grants write access"
            }
        },
        "OAuth2Password": {
            "type": "oauth2",
            "flow": "password",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "read": " Grants read access",
                "write": " Grants write access"
            }
        }
    }
}`
	searchDir := "testdata/simple2"
	p := New()
	p.PropNamingStrategy = SnakeCase
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseSimpleApi_ForLowerCamelcase(t *testing.T) {
	t.Parallel()

	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0"
    },
    "host": "petstore.swagger.io",
    "basePath": "/v2",
    "paths": {
        "/file/upload": {
            "post": {
                "description": "Upload file",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Upload file",
                "operationId": "file.upload",
                "parameters": [
                    {
                        "type": "file",
                        "description": "this is a test file",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "We need ID!!",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    }
                }
            }
        },
        "/testapi/get-string-by-int/{some_id}": {
            "get": {
                "description": "get string by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Add a new pet to the store",
                "operationId": "get-string-by-int",
                "parameters": [
                    {
                        "type": "integer",
                        "format": "int64",
                        "description": "Some ID",
                        "name": "some_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Some ID",
                        "name": "some_id",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.Pet"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "We need ID!!",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    }
                }
            }
        },
        "/testapi/get-struct-array-by-string/{some_id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    },
                    {
                        "BasicAuth": []
                    },
                    {
                        "OAuth2Application": [
                            "write"
                        ]
                    },
                    {
                        "OAuth2Implicit": [
                            "read",
                            "admin"
                        ]
                    },
                    {
                        "OAuth2AccessCode": [
                            "read"
                        ]
                    },
                    {
                        "OAuth2Password": [
                            "admin"
                        ]
                    }
                ],
                "description": "get struct array by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "operationId": "get-struct-array-by-string",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Some ID",
                        "name": "some_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "enum": [
                            1,
                            2,
                            3
                        ],
                        "type": "integer",
                        "description": "Category",
                        "name": "category",
                        "in": "query",
                        "required": true
                    },
                    {
                        "minimum": 0,
                        "type": "integer",
                        "default": 0,
                        "description": "Offset",
                        "name": "offset",
                        "in": "query",
                        "required": true
                    },
                    {
                        "maximum": 50,
                        "type": "integer",
                        "default": 10,
                        "description": "Limit",
                        "name": "limit",
                        "in": "query",
                        "required": true
                    },
                    {
                        "maxLength": 50,
                        "minLength": 1,
                        "type": "string",
                        "default": "\"\"",
                        "description": "q",
                        "name": "q",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "We need ID!!",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "web.APIError": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "errorCode": {
                    "type": "integer"
                },
                "errorMessage": {
                    "type": "string"
                }
            }
        },
        "web.Pet": {
            "type": "object",
            "properties": {
                "category": {
                    "type": "object",
                    "properties": {
                        "id": {
                            "type": "integer",
                            "example": 1
                        },
                        "name": {
                            "type": "string",
                            "example": "category_name"
                        },
                        "photoURLs": {
                            "type": "array",
                            "items": {
                                "type": "string",
                                "format": "url"
                            },
                            "example": [
                                "http://test/image/1.jpg",
                                "http://test/image/2.jpg"
                            ]
                        },
                        "smallCategory": {
                            "type": "object",
                            "properties": {
                                "id": {
                                    "type": "integer",
                                    "example": 1
                                },
                                "name": {
                                    "type": "string",
                                    "example": "detail_category_name"
                                },
                                "photoURLs": {
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    },
                                    "example": [
                                        "http://test/image/1.jpg",
                                        "http://test/image/2.jpg"
                                    ]
                                }
                            }
                        }
                    }
                },
                "data": {},
                "decimal": {
                    "type": "number"
                },
                "id": {
                    "type": "integer",
                    "format": "int64",
                    "example": 1
                },
                "isAlive": {
                    "type": "boolean",
                    "example": true
                },
                "name": {
                    "type": "string",
                    "example": "poti"
                },
                "pets": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.Pet2"
                    }
                },
                "pets2": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.Pet2"
                    }
                },
                "photoURLs": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "http://test/image/1.jpg",
                        "http://test/image/2.jpg"
                    ]
                },
                "price": {
                    "type": "number",
                    "multipleOf": 0.01,
                    "example": 3.25
                },
                "status": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.Tag"
                    }
                },
                "uuid": {
                    "type": "string"
                }
            }
        },
        "web.Pet2": {
            "type": "object",
            "properties": {
                "deletedAt": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "middleName": {
                    "type": "string"
                }
            }
        },
        "web.RevValue": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "integer"
                },
                "err": {
                    "type": "integer"
                },
                "status": {
                    "type": "boolean"
                }
            }
        },
        "web.Tag": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "format": "int64"
                },
                "name": {
                    "type": "string"
                },
                "pets": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.Pet"
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        },
        "BasicAuth": {
            "type": "basic"
        },
        "OAuth2AccessCode": {
            "type": "oauth2",
            "flow": "accessCode",
            "authorizationUrl": "https://example.com/oauth/authorize",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information"
            }
        },
        "OAuth2Application": {
            "type": "oauth2",
            "flow": "application",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "write": " Grants write access"
            }
        },
        "OAuth2Implicit": {
            "type": "oauth2",
            "flow": "implicit",
            "authorizationUrl": "https://example.com/oauth/authorize",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "write": " Grants write access"
            }
        },
        "OAuth2Password": {
            "type": "oauth2",
            "flow": "password",
            "tokenUrl": "https://example.com/oauth/token",
            "scopes": {
                "admin": " Grants read and write access to administrative information",
                "read": " Grants read access",
                "write": " Grants write access"
            }
        }
    }
}`
	searchDir := "testdata/simple3"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseStructComment(t *testing.T) {
	t.Parallel()

	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.",
        "title": "Swagger Example API",
        "contact": {},
        "version": "1.0"
    },
    "host": "localhost:4000",
    "basePath": "/api",
    "paths": {
        "/posts/{post_id}": {
            "get": {
                "description": "get string by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Add a new pet to the store",
                "parameters": [
                    {
                        "type": "integer",
                        "format": "int64",
                        "description": "Some ID",
                        "name": "post_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "We need ID!!",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "$ref": "#/definitions/web.APIError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "web.APIError": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "description": "Error time",
                    "type": "string"
                },
                "error": {
                    "description": "Error an Api error",
                    "type": "string"
                },
                "errorCtx": {
                    "description": "Error ` + "`" + `context` + "`" + ` tick comment",
                    "type": "string"
                },
                "errorNo": {
                    "description": "Error ` + "`" + `number` + "`" + ` tick comment",
                    "type": "integer"
                }
            }
        }
    }
}`
	searchDir := "testdata/struct_comment"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseNonExportedJSONFields(t *testing.T) {
	t.Parallel()

	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server.",
        "title": "Swagger Example API",
        "contact": {},
        "version": "1.0"
    },
    "host": "localhost:4000",
    "basePath": "/api",
    "paths": {
        "/so-something": {
            "get": {
                "description": "Does something, but internal (non-exported) fields inside a struct won't be marshaled into JSON",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Call DoSomething",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/main.MyStruct"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "main.MyStruct": {
            "type": "object",
            "properties": {
                "data": {
                    "description": "Post data",
                    "type": "object",
                    "properties": {
                        "name": {
                            "description": "Post tag",
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    }
                },
                "id": {
                    "type": "integer",
                    "format": "int64",
                    "example": 1
                },
                "name": {
                    "description": "Post name",
                    "type": "string",
                    "example": "poti"
                }
            }
        }
    }
}`

	searchDir := "testdata/non_exported_json_fields"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParsePetApi(t *testing.T) {
	t.Parallel()

	expected := `{
    "schemes": [
        "http",
        "https"
    ],
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.  You can find out more about     Swagger at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).      For this sample, you can use the api key 'special-key' to test the authorization     filters.",
        "title": "Swagger Petstore",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "email": "apiteam@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0"
    },
    "host": "petstore.swagger.io",
    "basePath": "/v2",
    "paths": {}
}`
	searchDir := "testdata/pet"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseModelAsTypeAlias(t *testing.T) {
	t.Parallel()

	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0"
    },
    "host": "petstore.swagger.io",
    "basePath": "/v2",
    "paths": {
        "/testapi/time-as-time-container": {
            "get": {
                "description": "test container with time and time alias",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get container with time and time alias",
                "operationId": "time-as-time-container",
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "$ref": "#/definitions/data.TimeContainer"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "data.TimeContainer": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        }
    }
}`
	searchDir := "testdata/alias_type"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseComposition(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/composition"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")

	//windows will fail: \r\n \n
	assert.Equal(t, string(expected), string(b))
}

func TestParseImportAliases(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/alias_import"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	//windows will fail: \r\n \n
	assert.Equal(t, string(expected), string(b))
}

func TestParseNested(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/nested"
	p := New()
	p.ParseDependency = true
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, string(expected), string(b))
}

func TestParseDuplicated(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/duplicated"
	p := New()
	p.ParseDependency = true
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.Errorf(t, err, "duplicated @id declarations successfully found")
}

func TestParseDuplicatedOtherMethods(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/duplicated2"
	p := New()
	p.ParseDependency = true
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.Errorf(t, err, "duplicated @id declarations successfully found")
}

func TestParseConflictSchemaName(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/conflict_name"
	p := New()
	p.ParseDependency = true
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(b))
}

func TestParseExternalModels(t *testing.T) {
	searchDir := "testdata/external_models/main"
	mainAPIFile := "main.go"
	p := New()
	p.ParseDependency = true
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	//ioutil.WriteFile("./testdata/external_models/main/expected.json",b,0777)
	expected, err := ioutil.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(b))
}

func TestParser_ParseStructArrayObject(t *testing.T) {
	t.Parallel()

	src := `
package api

type Response struct {
	Code int
	Table [][]string
	Data []struct{
		Field1 uint
		Field2 string
	}
}

// @Success 200 {object} Response
// @Router /api/{id} [get]
func Test(){
}
`
	expected := `{
   "api.Response": {
      "type": "object",
      "properties": {
         "code": {
            "type": "integer"
         },
         "data": {
            "type": "array",
            "items": {
               "type": "object",
               "properties": {
                  "field1": {
                     "type": "integer"
                  },
                  "field2": {
                     "type": "string"
                  }
               }
            }
         },
         "table": {
            "type": "array",
            "items": {
               "type": "array",
               "items": {
                  "type": "string"
               }
            }
         }
      }
   }
}`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	p.packages.CollectAstFile("api", "api/api.go", f)
	_, err = p.packages.ParseTypes()
	assert.NoError(t, err)

	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	out, err := json.MarshalIndent(p.swagger.Definitions, "", "   ")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(out))

}

func TestParser_ParseEmbededStruct(t *testing.T) {
	t.Parallel()

	src := `
package api

type Response struct {
	rest.ResponseWrapper
}

// @Success 200 {object} Response
// @Router /api/{id} [get]
func Test(){
}
`
	restsrc := `
package rest

type ResponseWrapper struct {
	Status   string
	Code     int
	Messages []string
	Result   interface{}
}
`
	expected := `{
   "api.Response": {
      "type": "object",
      "properties": {
         "code": {
            "type": "integer"
         },
         "messages": {
            "type": "array",
            "items": {
               "type": "string"
            }
         },
         "result": {},
         "status": {
            "type": "string"
         }
      }
   }
}`
	parser := New()
	parser.ParseDependency = true

	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)
	parser.packages.CollectAstFile("api", "api/api.go", f)

	f2, err := goparser.ParseFile(token.NewFileSet(), "", restsrc, goparser.ParseComments)
	assert.NoError(t, err)
	parser.packages.CollectAstFile("rest", "rest/rest.go", f2)

	_, err = parser.packages.ParseTypes()
	assert.NoError(t, err)

	err = parser.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	out, err := json.MarshalIndent(parser.swagger.Definitions, "", "   ")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(out))

}

func TestParser_ParseStructPointerMembers(t *testing.T) {
	t.Parallel()

	src := `
package api

type Child struct {
	Name string
}

type Parent struct {
	Test1 *string  //test1
	Test2 *Child   //test2
}

// @Success 200 {object} Parent
// @Router /api/{id} [get]
func Test(){
}
`

	expected := `{
   "api.Child": {
      "type": "object",
      "properties": {
         "name": {
            "type": "string"
         }
      }
   },
   "api.Parent": {
      "type": "object",
      "properties": {
         "test1": {
            "description": "test1",
            "type": "string"
         },
         "test2": {
            "description": "test2",
            "$ref": "#/definitions/api.Child"
         }
      }
   }
}`

	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	p.packages.CollectAstFile("api", "api/api.go", f)
	_, err = p.packages.ParseTypes()
	assert.NoError(t, err)

	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	out, err := json.MarshalIndent(p.swagger.Definitions, "", "   ")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(out))
}

func TestParser_ParseStructMapMember(t *testing.T) {
	t.Parallel()

	src := `
package api

type MyMapType map[string]string

type Child struct {
	Name string
}

type Parent struct {
	Test1 map[string]interface{}  //test1
	Test2 map[string]string		  //test2
	Test3 map[string]*string	  //test3
	Test4 map[string]Child		  //test4
	Test5 map[string]*Child		  //test5
	Test6 MyMapType				  //test6
	Test7 []Child				  //test7
	Test8 []*Child				  //test8
	Test9 []map[string]string	  //test9
}

// @Success 200 {object} Parent
// @Router /api/{id} [get]
func Test(){
}
`
	expected := `{
   "api.Child": {
      "type": "object",
      "properties": {
         "name": {
            "type": "string"
         }
      }
   },
   "api.MyMapType": {
      "type": "object",
      "additionalProperties": {
         "type": "string"
      }
   },
   "api.Parent": {
      "type": "object",
      "properties": {
         "test1": {
            "description": "test1",
            "type": "object",
            "additionalProperties": true
         },
         "test2": {
            "description": "test2",
            "type": "object",
            "additionalProperties": {
               "type": "string"
            }
         },
         "test3": {
            "description": "test3",
            "type": "object",
            "additionalProperties": {
               "type": "string"
            }
         },
         "test4": {
            "description": "test4",
            "type": "object",
            "additionalProperties": {
               "$ref": "#/definitions/api.Child"
            }
         },
         "test5": {
            "description": "test5",
            "type": "object",
            "additionalProperties": {
               "$ref": "#/definitions/api.Child"
            }
         },
         "test6": {
            "description": "test6",
            "$ref": "#/definitions/api.MyMapType"
         },
         "test7": {
            "description": "test7",
            "type": "array",
            "items": {
               "$ref": "#/definitions/api.Child"
            }
         },
         "test8": {
            "description": "test8",
            "type": "array",
            "items": {
               "$ref": "#/definitions/api.Child"
            }
         },
         "test9": {
            "description": "test9",
            "type": "array",
            "items": {
               "type": "object",
               "additionalProperties": {
                  "type": "string"
               }
            }
         }
      }
   }
}`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	p.packages.CollectAstFile("api", "api/api.go", f)

	_, err = p.packages.ParseTypes()
	assert.NoError(t, err)

	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	out, err := json.MarshalIndent(p.swagger.Definitions, "", "   ")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(out))
}

func TestParser_ParseRouterApiInfoErr(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Accept unknown
func Test(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	err = p.ParseRouterAPIInfo("", f)
	assert.EqualError(t, err, "ParseComment error in file  :unknown accept type can't be accepted")
}

func TestParser_ParseRouterApiGet(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/{id} [get]
func Test(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	ps := p.swagger.Paths.Paths

	val, ok := ps["/api/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Get)
}

func TestParser_ParseRouterApiPOST(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/{id} [post]
func Test(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	ps := p.swagger.Paths.Paths

	val, ok := ps["/api/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Post)
}

func TestParser_ParseRouterApiDELETE(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/{id} [delete]
func Test(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)
	p := New()

	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	ps := p.swagger.Paths.Paths

	val, ok := ps["/api/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Delete)
}

func TestParser_ParseRouterApiPUT(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/{id} [put]
func Test(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	ps := p.swagger.Paths.Paths

	val, ok := ps["/api/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Put)
}

func TestParser_ParseRouterApiPATCH(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/{id} [patch]
func Test(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	ps := p.swagger.Paths.Paths

	val, ok := ps["/api/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Patch)
}

func TestParser_ParseRouterApiHead(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/{id} [head]
func Test(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)
	p := New()

	err = p.ParseRouterAPIInfo("", f)

	assert.NoError(t, err)
	ps := p.swagger.Paths.Paths

	val, ok := ps["/api/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Head)
}

func TestParser_ParseRouterApiOptions(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/{id} [options]
func Test(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	ps := p.swagger.Paths.Paths

	val, ok := ps["/api/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Options)
}

func TestParser_ParseRouterApiMultipleRoutesForSameFunction(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/v1/{id} [get]
// @Router /api/v2/{id} [post]
func Test(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	ps := p.swagger.Paths.Paths

	val, ok := ps["/api/v1/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Get)

	val, ok = ps["/api/v2/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Post)
}

func TestParser_ParseRouterApiMultiple(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/{id} [get]
func Test1(){
}

// @Router /api/{id} [patch]
func Test2(){
}

// @Router /api/{id} [delete]
func Test3(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	ps := p.swagger.Paths.Paths

	val, ok := ps["/api/{id}"]

	assert.True(t, ok)
	assert.NotNil(t, val.Get)
	assert.NotNil(t, val.Patch)
	assert.NotNil(t, val.Delete)
}

// func TestParseDeterministic(t *testing.T) {
// 	mainAPIFile := "main.go"
// 	for _, searchDir := range []string{
// 		"testdata/simple",
// 		"testdata/model_not_under_root/cmd",
// 	} {
// 		t.Run(searchDir, func(t *testing.T) {
// 			var expected string

// 			// run the same code 100 times and check that the output is the same every time
// 			for i := 0; i < 100; i++ {
// 				p := New()
// 				p.PropNamingStrategy = PascalCase
// 				err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
// 				b, _ := json.MarshalIndent(p.swagger, "", "    ")
// 				assert.NotEqual(t, "", string(b))

// 				if expected == "" {
// 					expected = string(b)
// 				}

// 				assert.Equal(t, expected, string(b))
// 			}
// 		})
// 	}
// }

func TestParser_ParseRouterApiDuplicateRoute(t *testing.T) {
	t.Parallel()

	src := `
package test

// @Router /api/{id} [get]
func Test1(){
}
// @Router /api/{id} [get]
func Test2(){
}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New(SetStrict(true))
	err = p.ParseRouterAPIInfo("", f)
	assert.EqualError(t, err, "route GET /api/{id} is declared multiple times")

	p = New()
	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)
}

func TestApiParseTag(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/tags"
	p := New(SetMarkdownFileDirectory(searchDir))
	p.PropNamingStrategy = PascalCase
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)

	if len(p.swagger.Tags) != 3 {
		t.Error("Number of tags did not match")
	}

	dogs := p.swagger.Tags[0]
	if dogs.TagProps.Name != "dogs" || dogs.TagProps.Description != "Dogs are cool" {
		t.Error("Failed to parse dogs name or description")
	}

	cats := p.swagger.Tags[1]
	if cats.TagProps.Name != "cats" || cats.TagProps.Description != "Cats are the devil" {
		t.Error("Failed to parse cats name or description")
	}

	if cats.TagProps.ExternalDocs.URL != "https://google.de" || cats.TagProps.ExternalDocs.Description != "google is super useful to find out that cats are evil!" {
		t.Error("URL: ", cats.TagProps.ExternalDocs.URL)
		t.Error("Description: ", cats.TagProps.ExternalDocs.Description)
		t.Error("Failed to parse cats external documentation")
	}
}

func TestApiParseTag_NonExistendTag(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/tags_nonexistend_tag"
	p := New(SetMarkdownFileDirectory(searchDir))
	p.PropNamingStrategy = PascalCase
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.Error(t, err)
}

func TestParseTagMarkdownDescription(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/tags"
	p := New(SetMarkdownFileDirectory(searchDir))
	p.PropNamingStrategy = PascalCase
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	if err != nil {
		t.Error("Failed to parse api description: " + err.Error())
	}

	if len(p.swagger.Tags) != 3 {
		t.Error("Number of tags did not match")
	}

	apes := p.swagger.Tags[2]
	if apes.TagProps.Description == "" {
		t.Error("Failed to parse tag description markdown file")
	}
}

func TestParseApiMarkdownDescription(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/tags"
	p := New(SetMarkdownFileDirectory(searchDir))
	p.PropNamingStrategy = PascalCase
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	if err != nil {
		t.Error("Failed to parse api description: " + err.Error())
	}

	if p.swagger.Info.Description == "" {
		t.Error("Failed to parse api description: " + err.Error())
	}
}

func TestIgnoreInvalidPkg(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/deps_having_invalid_pkg"
	p := New()
	if err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth); err != nil {
		t.Error("Failed to ignore valid pkg: " + err.Error())
	}
}

func TestFixes432(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/fixes-432"
	mainAPIFile := "cmd/main.go"

	p := New()
	if err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth); err != nil {
		t.Error("Failed to ignore valid pkg: " + err.Error())
	}
}

func TestParseOutsideDependencies(t *testing.T) {
	t.Parallel()

	searchDir := "testdata/pare_outside_dependencies"
	mainAPIFile := "cmd/main.go"

	p := New()
	p.ParseDependency = true
	if err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth); err != nil {
		t.Error("Failed to parse api: " + err.Error())
	}
}

func TestParseStructParamCommentByQueryType(t *testing.T) {
	t.Parallel()

	src := `
package main

type Student struct {
	Name string
	Age int
	Teachers []string
	SkipField map[string]string
}

// @Param request query Student true "query params"
// @Success 200
// @Router /test [get]
func Fun()  {

}
`
	expected := `{
    "info": {
        "contact": {}
    },
    "paths": {
        "/test": {
            "get": {
                "parameters": [
                    {
                        "type": "integer",
                        "name": "age",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "name": "name",
                        "in": "query"
                    },
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "name": "teachers",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        }
    }
}`

	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	p.packages.CollectAstFile("api", "api/api.go", f)

	_, err = p.packages.ParseTypes()
	assert.NoError(t, err)

	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseRenamedStructDefinition(t *testing.T) {
	t.Parallel()

	src := `
package main

type Child struct {
	Name string
}//@name Student

type Parent struct {
	Name string
	Child Child
}//@name Teacher

// @Param request body Parent true "query params"
// @Success 200 {object} Parent
// @Router /test [get]
func Fun()  {

}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	p.packages.CollectAstFile("api", "api/api.go", f)
	_, err = p.packages.ParseTypes()
	assert.NoError(t, err)

	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	assert.NoError(t, err)
	teacher, ok := p.swagger.Definitions["Teacher"]
	assert.True(t, ok)
	ref := teacher.Properties["child"].SchemaProps.Ref
	assert.Equal(t, "#/definitions/Student", ref.String())
	_, ok = p.swagger.Definitions["Student"]
	assert.True(t, ok)
	path, ok := p.swagger.Paths.Paths["/test"]
	assert.True(t, ok)
	assert.Equal(t, "#/definitions/Teacher", path.Get.Parameters[0].Schema.Ref.String())
	ref = path.Get.Responses.ResponsesProps.StatusCodeResponses[200].ResponseProps.Schema.Ref
	assert.Equal(t, "#/definitions/Teacher", ref.String())
}

func TestPackagesDefinitions_CollectAstFileInit(t *testing.T) {
	t.Parallel()

	src := `
package main

// @Router /test [get]
func Fun()  {

}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	pkgs := NewPackagesDefinitions()

	// unset the .files and .packages and check that they're re-initialized by CollectAstFile
	pkgs.packages = nil
	pkgs.files = nil

	pkgs.CollectAstFile("api", "api/api.go", f)
	assert.NotNil(t, pkgs.packages)
	assert.NotNil(t, pkgs.files)
}

func TestCollectAstFileMultipleTimes(t *testing.T) {
	t.Parallel()

	src := `
package main

// @Router /test [get]
func Fun()  {

}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	p.packages.CollectAstFile("api", "api/api.go", f)
	assert.NotNil(t, p.packages.files[f])

	astFileInfo := p.packages.files[f]

	// if we collect the same again nothing should happen
	p.packages.CollectAstFile("api", "api/api.go", f)
	assert.Equal(t, astFileInfo, p.packages.files[f])
}

func TestParseJSONFieldString(t *testing.T) {
	t.Parallel()

	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server.",
        "title": "Swagger Example API",
        "contact": {},
        "version": "1.0"
    },
    "host": "localhost:4000",
    "basePath": "/",
    "paths": {
        "/do-something": {
            "post": {
                "description": "Does something",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Call DoSomething",
                "parameters": [
                    {
                        "description": "My Struct",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.MyStruct"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/main.MyStruct"
                        }
                    },
                    "500": {
                        "description": ""
                    }
                }
            }
        }
    },
    "definitions": {
        "main.MyStruct": {
            "type": "object",
            "properties": {
                "boolvar": {
                    "description": "boolean as a string",
                    "type": "string",
                    "example": "false"
                },
                "floatvar": {
                    "description": "float as a string",
                    "type": "string",
                    "example": "0"
                },
                "id": {
                    "type": "integer",
                    "format": "int64",
                    "example": 1
                },
                "myint": {
                    "description": "integer as string",
                    "type": "string",
                    "example": "0"
                },
                "name": {
                    "type": "string",
                    "example": "poti"
                },
                "truebool": {
                    "description": "boolean as a string",
                    "type": "string",
                    "example": "true"
                },
                "uuids": {
                    "description": "string array with format",
                    "type": "array",
                    "items": {
                        "type": "string",
                        "format": "uuid"
                    }
                }
            }
        }
    }
}`

	searchDir := "testdata/json_field_string"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseSwaggerignoreForEmbedded(t *testing.T) {
	t.Parallel()

	src := `
package main

type Child struct {
	ChildName string
}//@name Student

type Parent struct {
	Name string
	Child ` + "`swaggerignore:\"true\"`" + `
}//@name Teacher

// @Param request body Parent true "query params"
// @Success 200 {object} Parent
// @Router /test [get]
func Fun()  {

}
`
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.ParseComments)
	assert.NoError(t, err)

	p := New()
	p.packages.CollectAstFile("api", "api/api.go", f)
	p.packages.ParseTypes()
	err = p.ParseRouterAPIInfo("", f)
	assert.NoError(t, err)

	assert.NoError(t, err)
	teacher, ok := p.swagger.Definitions["Teacher"]
	assert.True(t, ok)

	name, ok := teacher.Properties["name"]
	assert.True(t, ok)
	assert.Len(t, name.Type, 1)
	assert.Equal(t, "string", name.Type[0])

	childName, ok := teacher.Properties["childName"]
	assert.False(t, ok)
	assert.Empty(t, childName)
}

func TestDefineTypeOfExample(t *testing.T) {

	t.Run("String type", func(t *testing.T) {
		t.Parallel()

		example, err := defineTypeOfExample("string", "", "example")
		assert.NoError(t, err)
		assert.Equal(t, example.(string), "example")
	})

	t.Run("Number type", func(t *testing.T) {
		t.Parallel()

		example, err := defineTypeOfExample("number", "", "12.34")
		assert.NoError(t, err)
		assert.Equal(t, example.(float64), 12.34)

		_, err = defineTypeOfExample("number", "", "two")
		assert.Error(t, err)
	})

	t.Run("Integer type", func(t *testing.T) {
		t.Parallel()

		example, err := defineTypeOfExample("integer", "", "12")
		assert.NoError(t, err)
		assert.Equal(t, example.(int), 12)

		_, err = defineTypeOfExample("integer", "", "two")
		assert.Error(t, err)
	})

	t.Run("Boolean type", func(t *testing.T) {
		t.Parallel()

		example, err := defineTypeOfExample("boolean", "", "true")
		assert.NoError(t, err)
		assert.Equal(t, example.(bool), true)

		_, err = defineTypeOfExample("boolean", "", "!true")
		assert.Error(t, err)
	})

	t.Run("Array type", func(t *testing.T) {
		t.Parallel()

		example, err := defineTypeOfExample("array", "", "one,two,three")
		assert.Error(t, err)
		assert.Nil(t, example)

		example, err = defineTypeOfExample("array", "string", "one,two,three")
		assert.NoError(t, err)
		arr := []string{}

		for _, v := range example.([]interface{}) {
			arr = append(arr, v.(string))
		}

		assert.Equal(t, arr, []string{"one", "two", "three"})
	})

	t.Run("Object type", func(t *testing.T) {
		t.Parallel()

		example, err := defineTypeOfExample("object", "", "key_one:one,key_two:two,key_three:three")
		assert.Error(t, err)
		assert.Nil(t, example)

		example, err = defineTypeOfExample("object", "string", "key_one,key_two,key_three")
		assert.Error(t, err)
		assert.Nil(t, example)

		example, err = defineTypeOfExample("object", "oops", "key_one:one,key_two:two,key_three:three")
		assert.Error(t, err)
		assert.Nil(t, example)

		example, err = defineTypeOfExample("object", "string", "key_one:one,key_two:two,key_three:three")
		assert.NoError(t, err)
		obj := map[string]string{}

		for k, v := range example.(map[string]interface{}) {
			obj[k] = v.(string)
		}

		assert.Equal(t, obj, map[string]string{"key_one": "one", "key_two": "two", "key_three": "three"})
	})

	t.Run("Invalid type", func(t *testing.T) {
		t.Parallel()

		example, err := defineTypeOfExample("oops", "", "")
		assert.Error(t, err)
		assert.Nil(t, example)
	})
}

func TestParseFieldTag(t *testing.T) {

	t.Run("Example tag", func(t *testing.T) {
		t.Parallel()

		parser := New()

		field, err := parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" example:"one"`,
				},
			},
			[]string{"string"})
		assert.NoError(t, err)
		assert.Equal(t, &structField{
			schemaType:   "string",
			exampleValue: "one",
		}, field)

		_, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" example:"one"`,
				},
			},
			[]string{"float"})
		assert.Error(t, err)
	})

	t.Run("Format tag", func(t *testing.T) {
		t.Parallel()
		parser := New()

		field, err := parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" format:"csv"`,
				},
			},
			[]string{"string"})
		assert.NoError(t, err)
		assert.Equal(t, &structField{
			schemaType: "string",
			formatType: "csv",
		}, field)
	})

	t.Run("Required tag", func(t *testing.T) {
		t.Parallel()
		parser := New()

		field, err := parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" binding:"required"`,
				},
			},
			[]string{"string"})
		assert.NoError(t, err)
		assert.Equal(t, &structField{
			schemaType: "string",
			isRequired: true,
		}, field)

		field, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" validate:"required"`,
				},
			},
			[]string{"string"})
		assert.NoError(t, err)
		assert.Equal(t, &structField{
			schemaType: "string",
			isRequired: true,
		}, field)

	})

	t.Run("Extensions tag", func(t *testing.T) {

	})

	t.Run("Enums tag", func(t *testing.T) {
		t.Parallel()
		parser := New()

		field, err := parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" enums:"a,b,c"`,
				},
			},
			[]string{"string"})
		assert.NoError(t, err)
		assert.Equal(t, &structField{
			schemaType: "string",
			enums:      []interface{}{"a", "b", "c"},
		}, field)

		_, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" enums:"a,b,c"`,
				},
			},
			[]string{"float"})
		assert.Error(t, err)
	})

	t.Run("Default tag", func(t *testing.T) {
		t.Parallel()
		parser := New()

		field, err := parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" default:"pass"`,
				},
			},
			[]string{"string"})
		assert.NoError(t, err)
		assert.Equal(t, &structField{
			schemaType:   "string",
			defaultValue: "pass",
		}, field)

		_, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" default:"pass"`,
				},
			},
			[]string{"float"})
		assert.Error(t, err)
	})

	t.Run("Numeric value", func(t *testing.T) {
		t.Parallel()
		parser := New()

		field, err := parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" maximum:"1"`,
				},
			},
			[]string{"integer"})
		assert.NoError(t, err)
		max := float64(1)
		assert.Equal(t, &structField{
			schemaType: "integer",
			maximum:    &max,
		}, field)

		_, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" maximum:"one"`,
				},
			},
			[]string{"integer"})
		assert.Error(t, err)

		field, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" maximum:"1"`,
				},
			},
			[]string{"number"})
		assert.NoError(t, err)
		max = float64(1)
		assert.Equal(t, &structField{
			schemaType: "number",
			maximum:    &max,
		}, field)

		_, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" maximum:"one"`,
				},
			},
			[]string{"number"})
		assert.Error(t, err)

		field, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" multipleOf:"1"`,
				},
			},
			[]string{"number"})
		assert.NoError(t, err)
		multipleOf := float64(1)
		assert.Equal(t, &structField{
			schemaType: "number",
			multipleOf: &multipleOf,
		}, field)

		_, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" multipleOf:"one"`,
				},
			},
			[]string{"number"})
		assert.Error(t, err)

		field, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" minimum:"1"`,
				},
			},
			[]string{"integer"})
		assert.NoError(t, err)
		min := float64(1)
		assert.Equal(t, &structField{
			schemaType: "integer",
			minimum:    &min,
		}, field)

		_, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" minimum:"one"`,
				},
			},
			[]string{"integer"})
		assert.Error(t, err)

	})

	t.Run("String value", func(t *testing.T) {
		t.Parallel()
		parser := New()

		field, err := parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" maxLength:"1"`,
				},
			},
			[]string{"string"})
		assert.NoError(t, err)
		max := int64(1)
		assert.Equal(t, &structField{
			schemaType: "string",
			maxLength:  &max,
		}, field)

		_, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" maxLength:"one"`,
				},
			},
			[]string{"string"})
		assert.Error(t, err)

		field, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" minLength:"1"`,
				},
			},
			[]string{"string"})
		assert.NoError(t, err)
		min := int64(1)
		assert.Equal(t, &structField{
			schemaType: "string",
			minLength:  &min,
		}, field)

		_, err = parser.parseFieldTag(
			&ast.Field{
				Tag: &ast.BasicLit{
					Value: `json:"test" minLength:"one"`,
				},
			},
			[]string{"string"})
		assert.Error(t, err)
	})
}

func TestSetRouteMethodOp(t *testing.T) {
	t.Parallel()

	op := spec.NewOperation("dummy")

	// choosing to test each method explicitly instead of table driven to avoid reliance on helpers

	pathItem := spec.PathItem{}
	setRouteMethodOp(&pathItem, http.MethodGet, op)
	assert.Equal(t, op, pathItem.Get)

	pathItem = spec.PathItem{}
	setRouteMethodOp(&pathItem, http.MethodPost, op)
	assert.Equal(t, op, pathItem.Post)

	pathItem = spec.PathItem{}
	setRouteMethodOp(&pathItem, http.MethodDelete, op)
	assert.Equal(t, op, pathItem.Delete)

	pathItem = spec.PathItem{}
	setRouteMethodOp(&pathItem, http.MethodPut, op)
	assert.Equal(t, op, pathItem.Put)

	pathItem = spec.PathItem{}
	setRouteMethodOp(&pathItem, http.MethodPatch, op)
	assert.Equal(t, op, pathItem.Patch)

	pathItem = spec.PathItem{}
	setRouteMethodOp(&pathItem, http.MethodHead, op)
	assert.Equal(t, op, pathItem.Head)

	pathItem = spec.PathItem{}
	setRouteMethodOp(&pathItem, http.MethodOptions, op)
	assert.Equal(t, op, pathItem.Options)
}

func TestHasRouteMethodOp(t *testing.T) {
	t.Parallel()

	pathItem := spec.PathItem{}

	// assert that an invalid http method produces false
	assert.False(t, hasRouteMethodOp(pathItem, "OOPSIE"))

	// test each (supported) http method
	httpMethods := []string{
		http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut,
		http.MethodPatch, http.MethodHead, http.MethodOptions,
	}
	for _, httpMethod := range httpMethods {
		pathItem = spec.PathItem{}

		// should be false before setting
		assert.False(t, hasRouteMethodOp(pathItem, httpMethod))

		// and true after we set it
		// we rely on setRouteMethodOp, which is tested more thoroughly above
		setRouteMethodOp(&pathItem, httpMethod, spec.NewOperation("dummy"))
		assert.True(t, hasRouteMethodOp(pathItem, httpMethod))
	}
}

type mockFS struct {
	os.FileInfo
	FileName    string
	IsDirectory bool
}

func (fs *mockFS) Name() string {
	return fs.FileName
}

func (fs *mockFS) IsDir() bool {
	return fs.IsDirectory
}

func TestParser_Skip(t *testing.T) {
	t.Parallel()

	parser := New()
	parser.ParseVendor = true

	assert.NoError(t, parser.Skip("", &mockFS{FileName: "vendor"}))
	assert.NoError(t, parser.Skip("", &mockFS{FileName: "vendor", IsDirectory: true}))

	parser.ParseVendor = false
	assert.NoError(t, parser.Skip("", &mockFS{FileName: "vendor"}))
	assert.Error(t, parser.Skip("", &mockFS{FileName: "vendor", IsDirectory: true}))

	assert.NoError(t, parser.Skip("", &mockFS{FileName: "models", IsDirectory: true}))
	assert.NoError(t, parser.Skip("", &mockFS{FileName: "admin", IsDirectory: true}))
	assert.NoError(t, parser.Skip("", &mockFS{FileName: "release", IsDirectory: true}))

	parser = New(SetExcludedDirsAndFiles("admin/release,admin/models"))
	assert.NoError(t, parser.Skip("admin", &mockFS{IsDirectory: true}))
	assert.NoError(t, parser.Skip("admin/service", &mockFS{IsDirectory: true}))
	assert.Error(t, parser.Skip("admin/models", &mockFS{IsDirectory: true}))
	assert.Error(t, parser.Skip("admin/release", &mockFS{IsDirectory: true}))
}

func TestGetFieldType(t *testing.T) {
	t.Parallel()

	field, err := getFieldType(&ast.Ident{Name: "User"})
	assert.NoError(t, err)
	assert.Equal(t, "User", field)

	_, err = getFieldType(&ast.FuncType{})
	assert.Error(t, err)

	field, err = getFieldType(&ast.SelectorExpr{X: &ast.Ident{Name: "models"}, Sel: &ast.Ident{Name: "User"}})
	assert.NoError(t, err)
	assert.Equal(t, "models.User", field)

	_, err = getFieldType(&ast.SelectorExpr{X: &ast.FuncType{}, Sel: &ast.Ident{Name: "User"}})
	assert.Error(t, err)

	field, err = getFieldType(&ast.StarExpr{X: &ast.Ident{Name: "User"}})
	assert.NoError(t, err)
	assert.Equal(t, "User", field)

	field, err = getFieldType(&ast.StarExpr{X: &ast.FuncType{}})
	assert.Error(t, err)

	field, err = getFieldType(&ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "models"}, Sel: &ast.Ident{Name: "User"}}})
	assert.NoError(t, err)
	assert.Equal(t, "models.User", field)

}
