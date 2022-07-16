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
        }
    },
    "definitions": {
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
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponse-web_Post"
                        }
                    },
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponse-web_GenericListResponse_web_Post"
                        }
                    },
                    "202": {
                        "description": "Accepted",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponseMulti-web_Post-web_GenericListResponse_web_Post"
                        }
                    },
                    "222": {
                        "description": "",
                        "schema": {
                            "$ref": "#/definitions/web.GenericListResponseMulti-web_GenericListResponse_web_Post-web_Post"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "web.GenericListResponse-web_GenericListResponse_web_Post": {
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
        "web.GenericListResponseMulti-web_GenericListResponse_web_Post-web_Post": {
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
                        },
                        "status": {
                            "description": "Status of some other stuff",
                            "type": "string"
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
        "web.GenericListResponseMulti-web_Post-web_GenericListResponse_web_Post": {
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
