package swag

import (
	"encoding/json"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	swagMode = test
	New()
}

func TestParser_ParseGeneralApiInfo(t *testing.T) {
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
        "version": "1.0"
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
	err := p.ParseGeneralAPIInfo("testdata/main.go")
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParser_ParseGeneralApiInfoTemplated(t *testing.T) {
	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
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
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
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
	// should be return an error because extension value is not a valid json
	func() {
		expected := "@x-google-endpoints need a valid json value"
		gopath := os.Getenv("GOPATH")
		assert.NotNil(t, gopath)
		p := New()
		err := p.ParseGeneralAPIInfo("testdata/extensionsFail1.go")
		if assert.Error(t, err) {
			assert.Equal(t, expected, err.Error())
		}
	}()

	// should be return an error because extension don't have a value
	func() {
		expected := "@x-google-endpoints need a value"
		gopath := os.Getenv("GOPATH")
		assert.NotNil(t, gopath)
		p := New()
		err := p.ParseGeneralAPIInfo("testdata/extensionsFail2.go")
		if assert.Error(t, err) {
			assert.Equal(t, expected, err.Error())
		}
	}()
}

func TestParser_ParseGeneralApiInfoWithOpsInSameFile(t *testing.T) {
	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.\nIt has a lot of beautiful features.",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {},
        "license": {},
        "version": "1.0"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
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

func TestParser_ParseGeneralApiInfoFailed(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)
	p := New()
	assert.Error(t, p.ParseGeneralAPIInfo("testdata/noexist.go"))
}

func TestGetAllGoFileInfo(t *testing.T) {
	searchDir := "testdata/pet"

	p := New()
	err := p.getAllGoFileInfo(searchDir)

	assert.NoError(t, err)
	assert.NotEmpty(t, p.files["testdata/pet/main.go"])
	assert.NotEmpty(t, p.files["testdata/pet/web/handler.go"])
	assert.Equal(t, 2, len(p.files))
}

func TestParser_ParseType(t *testing.T) {
	searchDir := "testdata/simple/"

	p := New()
	err := p.getAllGoFileInfo(searchDir)
	assert.NoError(t, err)

	for _, file := range p.files {
		p.ParseType(file)
	}

	assert.NotNil(t, p.TypeDefinitions["api"]["Pet3"])
	assert.NotNil(t, p.TypeDefinitions["web"]["Pet"])
	assert.NotNil(t, p.TypeDefinitions["web"]["Pet2"])
}

func TestGetSchemes(t *testing.T) {
	schemes := getSchemes("@schemes http https")
	expectedSchemes := []string{"http", "https"}
	assert.Equal(t, expectedSchemes, schemes)
}

func TestParseSimpleApi1(t *testing.T) {
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
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
                            "type": "object",
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.SwagReturn": {
            "type": "array",
            "items": {
                "type": "object",
                "additionalProperties": {}
            }
        },
        "cross.Cross": {
            "type": "object",
            "properties": {
                "Array": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "String": {
                    "type": "string"
                }
            }
        },
        "web.APIError": {
            "type": "object",
            "properties": {
                "CreatedAt": {
                    "type": "string"
                },
                "ErrorCode": {
                    "type": "integer"
                },
                "ErrorMessage": {
                    "type": "string"
                }
            }
        },
        "web.AnonymousStructArray": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "foo": {
                        "type": "string"
                    }
                }
            }
        },
        "web.CrossAlias": {
            "type": "object",
            "properties": {
                "Array": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "String": {
                    "type": "string"
                }
            }
        },
        "web.IndirectRecursiveTest": {
            "type": "object",
            "properties": {
                "Tags": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/web.Tag"
                    }
                }
            }
        },
        "web.Pet": {
            "type": "object",
            "required": [
                "name",
                "photo_urls"
            ],
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
                        "photo_urls": {
                            "type": "array",
                            "format": "url",
                            "items": {
                                "type": "string"
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
                                    "maxLength": 16,
                                    "minLength": 4,
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
                "data": {
                    "type": "object"
                },
                "decimal": {
                    "type": "number"
                },
                "enum_array": {
                    "type": "array",
                    "items": {
                        "type": "integer",
                        "enum": [
                            1,
                            2,
                            3,
                            5,
                            7
                        ]
                    }
                },
                "id": {
                    "type": "integer",
                    "format": "int64",
                    "example": 1
                },
                "int_array": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    },
                    "example": [
                        1,
                        2
                    ]
                },
                "is_alive": {
                    "type": "boolean",
                    "default": true,
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
                    "maximum": 1000,
                    "minimum": 1,
                    "example": 3.25
                },
                "status": {
                    "type": "string",
                    "enum": [
                        "healthy",
                        "ill"
                    ]
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
                "middlename": {
                    "type": "string",
                    "x-abc": "def",
                    "x-nullable": true
                }
            }
        },
        "web.Pet5a": {
            "type": "object",
            "required": [
                "name",
                "odd"
            ],
            "properties": {
                "name": {
                    "type": "string"
                },
                "odd": {
                    "type": "boolean"
                }
            }
        },
        "web.Pet5b": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "name": {
                    "type": "string"
                }
            }
        },
        "web.Pet5c": {
            "type": "object",
            "required": [
                "name",
                "odd"
            ],
            "properties": {
                "name": {
                    "type": "string"
                },
                "odd": {
                    "type": "boolean"
                }
            }
        },
        "web.RevValue": {
            "type": "object",
            "properties": {
                "Data": {
                    "type": "integer"
                },
                "Err": {
                    "type": "integer"
                },
                "Status": {
                    "type": "boolean"
                },
                "cross": {
                    "type": "object",
                    "$ref": "#/definitions/cross.Cross"
                },
                "crosses": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/cross.Cross"
                    }
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
        },
        "web.Tags": {
            "type": "array",
            "items": {
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
	searchDir := "testdata/simple"
	mainAPIFile := "main.go"
	p := New()
	p.PropNamingStrategy = PascalCase
	err := p.ParseAPI(searchDir, mainAPIFile)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseSimpleApi_ForSnakecase(t *testing.T) {
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
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
                            "type": "object",
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
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
                            "format": "url",
                            "items": {
                                "type": "string"
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
                "data": {
                    "type": "object"
                },
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
	mainAPIFile := "main.go"
	p := New()
	p.PropNamingStrategy = SnakeCase
	err := p.ParseAPI(searchDir, mainAPIFile)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseSimpleApi_ForLowerCamelcase(t *testing.T) {
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
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
                            "type": "object",
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
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
                            "format": "url",
                            "items": {
                                "type": "string"
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
                "data": {
                    "type": "object"
                },
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
	mainAPIFile := "main.go"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseStructComment(t *testing.T) {
	expected := `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.",
        "title": "Swagger Example API",
        "contact": {},
        "license": {},
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
                            "type": "object",
                            "$ref": "#/definitions/web.APIError"
                        }
                    },
                    "404": {
                        "description": "Can not find ID",
                        "schema": {
                            "type": "object",
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
                    "description": "Error context tick comment",
                    "type": "string"
                },
                "errorNo": {
                    "description": "Error number tick comment",
                    "type": "integer"
                }
            }
        },
        "web.Post": {
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
	searchDir := "testdata/struct_comment"
	mainAPIFile := "main.go"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParsePetApi(t *testing.T) {
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
	mainAPIFile := "main.go"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseModelNotUnderRoot(t *testing.T) {
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
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Upload file",
                "operationId": "file.upload",
                "parameters": [
                    {
                        "description": "Foo to create",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "$ref": "#/definitions/data.Foo"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
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
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "object",
                            "$ref": "#/definitions/data.Foo"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "data.Foo": {
            "type": "object",
            "properties": {
                "field1": {
                    "type": "string"
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
	searchDir := "testdata/model_not_under_root/cmd"
	mainAPIFile := "main.go"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseModelAsTypeAlias(t *testing.T) {
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
                            "type": "object",
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
	mainAPIFile := "main.go"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseComposition(t *testing.T) {
	searchDir := "testdata/composition"
	mainAPIFile := "main.go"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile)
	assert.NoError(t, err)

	expected, err := ioutil.ReadFile(path.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, string(expected), string(b))
}

func TestParser_ParseRouterApiInfoErr(t *testing.T) {
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

func TestParser_ParseRouterApiMultiple(t *testing.T) {
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

func TestSkip(t *testing.T) {
	folder1 := "/tmp/vendor"
	err := os.Mkdir(folder1, os.ModePerm)
	assert.NoError(t, err)
	f1, _ := os.Stat(folder1)

	parser := New()

	assert.True(t, parser.Skip(folder1, f1) == filepath.SkipDir)
	assert.NoError(t, os.Remove(folder1))

	folder2 := "/tmp/.git"
	err = os.Mkdir(folder2, os.ModePerm)
	assert.NoError(t, err)
	f2, _ := os.Stat(folder2)

	assert.True(t, parser.Skip(folder2, f2) == filepath.SkipDir)
	assert.NoError(t, os.Remove(folder2))

	currentPath := "./"
	currentPathInfo, _ := os.Stat(currentPath)
	assert.True(t, parser.Skip(currentPath, currentPathInfo) == nil)
}

func TestSkipMustParseVendor(t *testing.T) {
	folder1 := "/tmp/vendor"
	err := os.Mkdir(folder1, os.ModePerm)
	assert.NoError(t, err)

	f1, _ := os.Stat(folder1)

	parser := New()
	parser.ParseVendor = true

	assert.True(t, parser.Skip(folder1, f1) == nil)
	assert.NoError(t, os.Remove(folder1))

	folder2 := "/tmp/.git"
	err = os.Mkdir(folder2, os.ModePerm)
	assert.NoError(t, err)

	f2, _ := os.Stat(folder2)

	assert.True(t, parser.Skip(folder2, f2) == filepath.SkipDir)
	assert.NoError(t, os.Remove(folder2))

	currentPath := "./"
	currentPathInfo, _ := os.Stat(currentPath)
	assert.True(t, parser.Skip(currentPath, currentPathInfo) == nil)

	folder3 := "/tmp/test/vendor/github.com/swaggo/swag"
	assert.NoError(t, os.MkdirAll(folder3, os.ModePerm))
	f3, _ := os.Stat(folder3)

	assert.Nil(t, parser.Skip(folder3, f3))
	assert.NoError(t, os.RemoveAll("/tmp/test"))
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
// 				err := p.ParseAPI(searchDir, mainAPIFile)
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

func TestApiParseTag(t *testing.T) {
	searchDir := "testdata/tags"
	mainAPIFile := "main.go"
	p := New(SetMarkdownFileDirectory(searchDir))
	p.PropNamingStrategy = PascalCase
	err := p.ParseAPI(searchDir, mainAPIFile)
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

func TestParseTagMarkdownDescription(t *testing.T) {
	searchDir := "testdata/tags"
	mainAPIFile := "main.go"
	p := New(SetMarkdownFileDirectory(searchDir))
	p.PropNamingStrategy = PascalCase
	err := p.ParseAPI(searchDir, mainAPIFile)
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
	searchDir := "testdata/tags"
	mainAPIFile := "main.go"
	p := New(SetMarkdownFileDirectory(searchDir))
	p.PropNamingStrategy = PascalCase
	err := p.ParseAPI(searchDir, mainAPIFile)
	if err != nil {
		t.Error("Failed to parse api description: " + err.Error())
	}

	if p.swagger.Info.Description == "" {
		t.Error("Failed to parse api description: " + err.Error())
	}
}

func TestIgnoreInvalidPkg(t *testing.T) {
	searchDir := "testdata/deps_having_invalid_pkg"
	mainAPIFile := "main.go"
	p := New()
	if err := p.ParseAPI(searchDir, mainAPIFile); err != nil {
		t.Error("Failed to ignore valid pkg: " + err.Error())
	}
}

func TestFixes432(t *testing.T) {
	searchDir := "testdata/fixes-432"
	mainAPIFile := "cmd/main.go"

	p := New()
	if err := p.ParseAPI(searchDir, mainAPIFile); err != nil {
		t.Error("Failed to ignore valid pkg: " + err.Error())
	}
}

func TestParseOutsideDependencies(t *testing.T) {
	searchDir := "testdata/pare_outside_dependencies"
	mainAPIFile := "cmd/main.go"

	p := New()
	p.ParseDependency = true
	if err := p.ParseAPI(searchDir, mainAPIFile); err != nil {
		t.Error("Failed to parse api: " + err.Error())
	}
}
