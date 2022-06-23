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
