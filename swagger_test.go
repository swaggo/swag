package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var doc = `{
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
                        "description": "Some ID",
                        "name": "some_id",
                        "in": "path",
                        "required": true,
                        "schema": {
                            "type": "int"
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
                        "description": "Some ID",
                        "name": "some_id",
                        "in": "path",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    },
                    {
                        "description": "Offset",
                        "name": "offset",
                        "in": "query",
                        "required": true,
                        "schema": {
                            "type": "int"
                        }
                    },
                    {
                        "description": "Offset",
                        "name": "limit",
                        "in": "query",
                        "required": true,
                        "schema": {
                            "type": "int"
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
        }
    },
    "securityDefinitions": {
        "ApiKey": {
            "description: "some",
            "type": "apiKey",
            "name": "X-API-KEY",
            "in": "header"
        }
    }
}`

type s struct{}

func (s *s) ReadDoc() string {
	return doc
}

func TestRegister(t *testing.T) {
	setup()
	Register(Name, &s{})
	d, _ := ReadDoc()
	assert.Equal(t, doc, d)
}

func TestRegisterByName(t *testing.T) {
	setup()
	Register("another_name", &s{})
	d, _ := ReadDoc("another_name")
	assert.Equal(t, doc, d)
}

func TestRegisterMultiple(t *testing.T) {
	setup()
	Register(Name, &s{})
	Register("another_name", &s{})
	d1, _ := ReadDoc(Name)
	d2, _ := ReadDoc("another_name")
	assert.Equal(t, doc, d1)
	assert.Equal(t, doc, d2)
}

func TestReadDocBeforeRegistered(t *testing.T) {
	setup()
	_, err := ReadDoc()
	assert.Error(t, err)
}

func TestReadDocWithInvalidName(t *testing.T) {
	setup()
	Register(Name, &s{})
	_, err := ReadDoc("invalid")
	assert.Error(t, err)
}

func TestNilRegister(t *testing.T) {
	setup()
	var swagger Swagger
	assert.Panics(t, func() {
		Register(Name, swagger)
	})
}

func TestCalledTwicelRegister(t *testing.T) {
	setup()
	assert.Panics(t, func() {
		Register(Name, &s{})
		Register(Name, &s{})
	})
}

func setup() {
	swags = nil
}

func TestGetSwagger(t *testing.T) {
	setup()
	instance := &s{}
	Register(Name, instance)
	swagger := GetSwagger(Name)
	assert.Equal(t, instance, swagger)

	swagger = GetSwagger("invalid")
	assert.Nil(t, swagger)
}
