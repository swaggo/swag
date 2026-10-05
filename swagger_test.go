package swag

import (
	"fmt"
	"sync"
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
	Clear()
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

func TestUnregister(t *testing.T) {
	setup()
	instance := &s{}

	// Register then Read works
	Register(Name, instance)
	d, err := ReadDoc()
	assert.NoError(t, err)
	assert.Equal(t, doc, d)
	assert.Equal(t, instance, GetSwagger(Name))

	// Unregister removes the registered instance
	Unregister(Name)
	assert.Nil(t, GetSwagger(Name))

	// Read returns ErrDocIsNotRegistered
	d, err = ReadDoc()
	assert.Empty(t, d)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDocIsNotRegistered)

	// Register again works without panicking
	assert.NotPanics(t, func() {
		Register(Name, instance)
	})
	d, err = ReadDoc()
	assert.NoError(t, err)
	assert.Equal(t, doc, d)
}

func TestUnregister_Named(t *testing.T) {
	setup()
	inst1 := &s{}
	inst2 := &s{}

	Register("doc1", inst1)
	Register("doc2", inst2)

	assert.Equal(t, inst1, GetSwagger("doc1"))
	assert.Equal(t, inst2, GetSwagger("doc2"))

	Unregister("doc1")

	assert.Nil(t, GetSwagger("doc1"))
	assert.Equal(t, inst2, GetSwagger("doc2"))

	_, err := ReadDoc("doc1")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDocIsNotRegistered)

	d2, err := ReadDoc("doc2")
	assert.NoError(t, err)
	assert.Equal(t, doc, d2)
}

func TestUnregister_NilOrNonExistent(t *testing.T) {
	setup()

	// Calling Unregister on an empty registry does not panic
	assert.NotPanics(t, func() {
		Unregister("non_existent")
	})

	// Calling Unregister for an unregistered name does not affect existing entries
	Register(Name, &s{})
	assert.NotPanics(t, func() {
		Unregister("non_existent")
	})

	d, err := ReadDoc()
	assert.NoError(t, err)
	assert.Equal(t, doc, d)
}

func TestClear(t *testing.T) {
	setup()
	Register(Name, &s{})
	Register("another_name", &s{})
	Register("third_name", &s{})

	assert.NotNil(t, GetSwagger(Name))
	assert.NotNil(t, GetSwagger("another_name"))
	assert.NotNil(t, GetSwagger("third_name"))

	Clear()

	assert.Nil(t, GetSwagger(Name))
	assert.Nil(t, GetSwagger("another_name"))
	assert.Nil(t, GetSwagger("third_name"))

	_, err := ReadDoc()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDocIsNotRegistered)

	_, err = ReadDoc("another_name")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDocIsNotRegistered)

	// Can register again cleanly after Clear
	assert.NotPanics(t, func() {
		Register(Name, &s{})
	})
	d, err := ReadDoc()
	assert.NoError(t, err)
	assert.Equal(t, doc, d)
}

func TestClear_NilMap(t *testing.T) {
	setup()
	assert.NotPanics(t, func() {
		Clear()
	})
	_, err := ReadDoc()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDocIsNotRegistered)
}

func TestConcurrentAccess(t *testing.T) {
	setup()

	const workers = 20
	const iterations = 50
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			docName := fmt.Sprintf("doc_%d", workerID)
			instance := &s{}

			for j := 0; j < iterations; j++ {
				switch j % 5 {
				case 0:
					Unregister(docName)
					Register(docName, instance)
				case 1:
					_ = GetSwagger(docName)
				case 2:
					_, _ = ReadDoc(docName)
				case 3:
					Unregister(docName)
				case 4:
					if workerID == 0 && j%10 == 0 {
						Clear()
					}
				}
			}
		}(i)
	}

	wg.Wait()
}
