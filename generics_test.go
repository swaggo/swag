//go:build go1.18
// +build go1.18

package swag

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGenericsBasic(t *testing.T) {
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
        "/posts-multi/": {
            "post": {
                "description": "get string by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Add new pets to the store",
                "parameters": [
                    {
                        "description": "Some ID",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.GenericBodyMulti-web_Post-web_Post"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.GenericResponse-web_Post"
                        }
                    },
                    "222": {
                        "description": "",
                        "schema": {
                            "$ref": "#/definitions/web.GenericResponseMulti-web_Post-web_Post"
                        }
                    }
                }
            }
        },
        "/posts-multis/": {
            "post": {
                "description": "get string by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Add new pets to the store",
                "parameters": [
                    {
                        "description": "Some ID",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.GenericBodyMulti-array_web_Post-array2_web_Post"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.GenericResponse-array_web_Post"
                        }
                    },
                    "222": {
                        "description": "",
                        "schema": {
                            "$ref": "#/definitions/web.GenericResponseMulti-array_web_Post-array2_web_Post"
                        }
                    }
                }
            }
        },
        "/posts/": {
            "post": {
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
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.GenericBody-web_Post"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.GenericResponse-web_Post"
                        }
                    },
                    "222": {
                        "description": "",
                        "schema": {
                            "$ref": "#/definitions/web.GenericResponseMulti-web_Post-web_Post"
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
            "description": "API error with information about it",
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
                    "description": "Error ` + "`context`" + ` tick comment",
                    "type": "string"
                },
                "errorNo": {
                    "description": "Error ` + "`number`" + ` tick comment",
                    "type": "integer"
                }
            }
        },
        "web.GenericBody-web_Post": {
            "type": "object",
            "properties": {
                "data": {
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
        },
        "web.GenericBodyMulti-array_web_Post-array2_web_Post": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
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
                },
                "meta": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
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
                }
            }
        },
        "web.GenericBodyMulti-web_Post-web_Post": {
            "type": "object",
            "properties": {
                "data": {
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
                },
                "meta": {
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
        },
        "web.GenericResponse-array_web_Post": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
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
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "web.GenericResponse-web_Post": {
            "type": "object",
            "properties": {
                "data": {
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
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "web.GenericResponseMulti-array_web_Post-array2_web_Post": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
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
                },
                "meta": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
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
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "web.GenericResponseMulti-web_Post-web_Post": {
            "type": "object",
            "properties": {
                "data": {
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
                },
                "meta": {
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
                },
                "status": {
                    "type": "string"
                }
            }
        }
    }
}`

	searchDir := "testdata/generics_basic"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(b))
}

func TestParseGenericsArrays(t *testing.T) {
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
        "/posts": {
            "get": {
                "description": "Get All of the Posts",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "List Posts",
                "parameters": [
                    {
                        "description": "Some ID",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.GenericListBody-web_Post"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponse-web_Post"
                        }
                    },
                    "222": {
                        "description": "",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponseMulti-web_Post-web_Post"
                        }
                    }
                }
            }
        },
        "/posts-multi": {
            "get": {
                "description": "get string by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Add new pets to the store",
                "parameters": [
                    {
                        "description": "Some ID",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.GenericListBodyMulti-web_Post-web_Post"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponse-web_Post"
                        }
                    },
                    "222": {
                        "description": "",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponseMulti-web_Post-web_Post"
                        }
                    }
                }
            }
        },
        "/posts-multis": {
            "get": {
                "description": "get string by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Add new pets to the store",
                "parameters": [
                    {
                        "description": "Some ID",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.GenericListBodyMulti-web_Post-array_web_Post"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponse-array_web_Post"
                        }
                    },
                    "222": {
                        "description": "",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponseMulti-web_Post-array_web_Post"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "web.GenericListBody-web_Post": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
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
            }
        },
        "web.GenericListBodyMulti-web_Post-array_web_Post": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
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
                },
                "meta": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
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
                }
            }
        },
        "web.GenericListBodyMulti-web_Post-web_Post": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
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
                },
                "meta": {
                    "type": "array",
                    "items": {
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
            }
        },
        "web.GenericListResponse-array_web_Post": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "Items from the list response",
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
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
                },
                "status": {
                    "description": "Status of some other stuff",
                    "type": "string"
                }
            }
        },
        "web.GenericListResponse-web_Post": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "Items from the list response",
                    "type": "array",
                    "items": {
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
                },
                "status": {
                    "description": "Status of some other stuff",
                    "type": "string"
                }
            }
        },
        "web.GenericListResponseMulti-web_Post-array_web_Post": {
            "type": "object",
            "properties": {
                "itemsOne": {
                    "description": "ItemsOne is the first thing",
                    "type": "array",
                    "items": {
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
                },
                "itemsTwo": {
                    "description": "ItemsTwo is the second thing",
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
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
                },
                "status": {
                    "description": "Status of the things",
                    "type": "string"
                }
            }
        },
        "web.GenericListResponseMulti-web_Post-web_Post": {
            "type": "object",
            "properties": {
                "itemsOne": {
                    "description": "ItemsOne is the first thing",
                    "type": "array",
                    "items": {
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
                },
                "itemsTwo": {
                    "description": "ItemsTwo is the second thing",
                    "type": "array",
                    "items": {
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
                },
                "status": {
                    "description": "Status of the things",
                    "type": "string"
                }
            }
        }
    }
}`

	searchDir := "testdata/generics_arrays"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(b))
}

func TestParseGenericsNested(t *testing.T) {
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
        "/posts": {
            "get": {
                "description": "Get All of the Posts",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "List Posts",
                "parameters": [
                    {
                        "description": "Some ID",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedBody-web_GenericInnerType_web_Post"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponse-web_Post"
                        }
                    },
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponse-web_GenericInnerType_web_Post"
                        }
                    },
                    "202": {
                        "description": "Accepted",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponseMulti-web_Post-web_GenericInnerMultiType_web_Post_web_Post"
                        }
                    },
                    "203": {
                        "description": "Non-Authoritative Information",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponseMulti-web_Post-web_GenericInnerMultiType_web_Post_web_GenericInnerType_web_Post"
                        }
                    },
                    "222": {
                        "description": "",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponseMulti-web_GenericInnerType_web_Post-web_Post"
                        }
                    }
                }
            }
        },
        "/posts-multis/": {
            "get": {
                "description": "Get All of the Posts",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "List Posts",
                "parameters": [
                    {
                        "description": "Some ID",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedBody-web_GenericInnerType_array_web_Post"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponse-array_web_Post"
                        }
                    },
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponse-array_web_GenericInnerType_web_Post"
                        }
                    },
                    "202": {
                        "description": "Accepted",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponse-array_web_GenericInnerType_array_web_Post"
                        }
                    },
                    "203": {
                        "description": "Non-Authoritative Information",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponseMulti-array_web_Post-web_GenericInnerMultiType_array_web_Post_web_Post"
                        }
                    },
                    "204": {
                        "description": "No Content",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponseMulti-array_web_Post-array_web_GenericInnerMultiType_array_web_Post_web_Post"
                        }
                    },
                    "205": {
                        "description": "Reset Content",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponseMulti-web_Post-web_GenericInnerMultiType_web_Post_array_web_GenericInnerType_array2_web_Post"
                        }
                    },
                    "222": {
                        "description": "",
                        "schema": {
                            "$ref": "#/definitions/web.GenericNestedResponseMulti-web_GenericInnerType_array_web_Post-array_web_Post"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "web.GenericNestedBody-web_GenericInnerType_array_web_Post": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "Items from the list response",
                    "type": "object",
                    "properties": {
                        "items": {
                            "description": "Items from the list response",
                            "type": "array",
                            "items": {
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
                    }
                },
                "status": {
                    "description": "Status of some other stuff",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedBody-web_GenericInnerType_web_Post": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "Items from the list response",
                    "type": "object",
                    "properties": {
                        "items": {
                            "description": "Items from the list response",
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
                },
                "status": {
                    "description": "Status of some other stuff",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponse-array_web_GenericInnerType_array_web_Post": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "Items from the list response",
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "properties": {
                                "items": {
                                    "description": "Items from the list response",
                                    "type": "array",
                                    "items": {
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
                            }
                        }
                    }
                },
                "status": {
                    "description": "Status of some other stuff",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponse-array_web_GenericInnerType_web_Post": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "Items from the list response",
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "properties": {
                                "items": {
                                    "description": "Items from the list response",
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
                        }
                    }
                },
                "status": {
                    "description": "Status of some other stuff",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponse-array_web_Post": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "Items from the list response",
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
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
                },
                "status": {
                    "description": "Status of some other stuff",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponse-web_GenericInnerType_web_Post": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "Items from the list response",
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "items": {
                                "description": "Items from the list response",
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
                    }
                },
                "status": {
                    "description": "Status of some other stuff",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponse-web_Post": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "Items from the list response",
                    "type": "array",
                    "items": {
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
                },
                "status": {
                    "description": "Status of some other stuff",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponseMulti-array_web_Post-array_web_GenericInnerMultiType_array_web_Post_web_Post": {
            "type": "object",
            "properties": {
                "itemOne": {
                    "description": "ItemsOne is the first thing",
                    "type": "array",
                    "items": {
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
                },
                "itemsTwo": {
                    "description": "ItemsTwo is the second thing",
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "object",
                            "properties": {
                                "itemOne": {
                                    "description": "ItemsOne is the first thing",
                                    "type": "array",
                                    "items": {
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
                                },
                                "itemsTwo": {
                                    "description": "ItemsTwo is the second thing",
                                    "type": "array",
                                    "items": {
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
                            }
                        }
                    }
                },
                "status": {
                    "description": "Status of the things",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponseMulti-array_web_Post-web_GenericInnerMultiType_array_web_Post_web_Post": {
            "type": "object",
            "properties": {
                "itemOne": {
                    "description": "ItemsOne is the first thing",
                    "type": "array",
                    "items": {
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
                },
                "itemsTwo": {
                    "description": "ItemsTwo is the second thing",
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "itemOne": {
                                "description": "ItemsOne is the first thing",
                                "type": "array",
                                "items": {
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
                            },
                            "itemsTwo": {
                                "description": "ItemsTwo is the second thing",
                                "type": "array",
                                "items": {
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
                        }
                    }
                },
                "status": {
                    "description": "Status of the things",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponseMulti-web_GenericInnerType_array_web_Post-array_web_Post": {
            "type": "object",
            "properties": {
                "itemOne": {
                    "description": "ItemsOne is the first thing",
                    "type": "object",
                    "properties": {
                        "items": {
                            "description": "Items from the list response",
                            "type": "array",
                            "items": {
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
                    }
                },
                "itemsTwo": {
                    "description": "ItemsTwo is the second thing",
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
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
                },
                "status": {
                    "description": "Status of the things",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponseMulti-web_GenericInnerType_web_Post-web_Post": {
            "type": "object",
            "properties": {
                "itemOne": {
                    "description": "ItemsOne is the first thing",
                    "type": "object",
                    "properties": {
                        "items": {
                            "description": "Items from the list response",
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
                },
                "itemsTwo": {
                    "description": "ItemsTwo is the second thing",
                    "type": "array",
                    "items": {
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
                },
                "status": {
                    "description": "Status of the things",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponseMulti-web_Post-web_GenericInnerMultiType_web_Post_array_web_GenericInnerType_array2_web_Post": {
            "type": "object",
            "properties": {
                "itemOne": {
                    "description": "ItemsOne is the first thing",
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
                },
                "itemsTwo": {
                    "description": "ItemsTwo is the second thing",
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "itemOne": {
                                "description": "ItemsOne is the first thing",
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
                            },
                            "itemsTwo": {
                                "description": "ItemsTwo is the second thing",
                                "type": "array",
                                "items": {
                                    "type": "array",
                                    "items": {
                                        "type": "object",
                                        "properties": {
                                            "items": {
                                                "description": "Items from the list response",
                                                "type": "array",
                                                "items": {
                                                    "type": "array",
                                                    "items": {
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
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                "status": {
                    "description": "Status of the things",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponseMulti-web_Post-web_GenericInnerMultiType_web_Post_web_GenericInnerType_web_Post": {
            "type": "object",
            "properties": {
                "itemOne": {
                    "description": "ItemsOne is the first thing",
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
                },
                "itemsTwo": {
                    "description": "ItemsTwo is the second thing",
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "itemOne": {
                                "description": "ItemsOne is the first thing",
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
                            },
                            "itemsTwo": {
                                "description": "ItemsTwo is the second thing",
                                "type": "array",
                                "items": {
                                    "type": "object",
                                    "properties": {
                                        "items": {
                                            "description": "Items from the list response",
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
                                }
                            }
                        }
                    }
                },
                "status": {
                    "description": "Status of the things",
                    "type": "string"
                }
            }
        },
        "web.GenericNestedResponseMulti-web_Post-web_GenericInnerMultiType_web_Post_web_Post": {
            "type": "object",
            "properties": {
                "itemOne": {
                    "description": "ItemsOne is the first thing",
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
                },
                "itemsTwo": {
                    "description": "ItemsTwo is the second thing",
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "itemOne": {
                                "description": "ItemsOne is the first thing",
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
                            },
                            "itemsTwo": {
                                "description": "ItemsTwo is the second thing",
                                "type": "array",
                                "items": {
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
                        }
                    }
                },
                "status": {
                    "description": "Status of the things",
                    "type": "string"
                }
            }
        }
    }
}`

	searchDir := "testdata/generics_nested"
	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(b))
}
