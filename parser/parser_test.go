package parser

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestNew(t *testing.T) {
	New()
}

func TestParser_ParseGeneralApiInfo(t *testing.T) {
	var expected = `{
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
    "paths": {}
}`
	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)
	p := New()
	p.ParseGeneralApiInfo(path.Join(gopath, "src", "github.com/swaggo/swag/example/simple/main.go"))

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestGetAllGoFileInfo(t *testing.T) {
	searchDir := "../example/simple"

	p := New()
	p.getAllGoFileInfo(searchDir)

	assert.NotEmpty(t, p.files["../example/simple/main.go"])
	assert.NotEmpty(t, p.files["../example/simple/web/handler.go"])
	assert.Equal(t, 4, len(p.files))
}

func TestParser_ParseType(t *testing.T) {
	searchDir := "../example/simple/"

	p := New()
	p.getAllGoFileInfo(searchDir)

	for _, file := range p.files {
		p.ParseType(file)
	}

	assert.NotNil(t, p.TypeDefinitions["api"]["Pet3"])
	assert.NotNil(t, p.TypeDefinitions["web"]["Pet"])
	assert.NotNil(t, p.TypeDefinitions["web"]["Pet2"])
}

func TestGetSchemes(t *testing.T) {
	fmt.Println(GetSchemes("@schemes http https"))

}
func TestParseSimpleApi(t *testing.T) {
	var expected = `{
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
                "parameters": [
                    {
                        "type": "int",
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
                "description": "get struct array by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Some ID",
                        "name": "some_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "int",
                        "description": "Offset",
                        "name": "offset",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "int",
                        "description": "Offset",
                        "name": "limit",
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
                "ErrorCode": {
                    "type": "int"
                },
                "ErrorMessage": {
                    "type": "string"
                }
            }
        },
        "web.Pet": {
            "type": "object",
            "properties": {
                "Category": {
                    "type": "object"
                },
                "ID": {
                    "type": "int"
                },
                "Name": {
                    "type": "string"
                },
                "PhotoUrls": {
                    "type": "array"
                },
                "Status": {
                    "type": "string"
                },
                "Tags": {
                    "type": "array"
                }
            }
        }
    }
}`
	searchDir := "../example/simple"
	mainApiFile := "main.go"
	p := New()
	p.ParseApi(searchDir, mainApiFile)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParsePetApi(t *testing.T) {
	//var expected = ``
	searchDir := "../example/pet"
	mainApiFile := "main.go"
	p := New()
	p.ParseApi(searchDir, mainApiFile)

	b, _ := json.MarshalIndent(p.swagger, "", "    ")
	//assert.Equal(t, expected, string(b))
	fmt.Println(string(b))
}