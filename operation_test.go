package swag

import (
	"encoding/json"
	goparser "go/parser"
	"go/token"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

func TestParseEmptyComment(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment("//", nil)

	assert.NoError(t, err)
}

func TestParseTagsComment(t *testing.T) {
	t.Parallel()

	expected := `{
    "tags": [
        "pet",
        "store",
        "user"
    ]
}`
	operation := NewOperation(nil)
	err := operation.ParseComment(`/@Tags pet, store,user`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseAcceptComment(t *testing.T) {
	t.Parallel()

	expected := `{
    "consumes": [
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
		"application/health+json"
    ]
}`
	comment := `/@Accept json,xml,plain,html,mpfd,x-www-form-urlencoded,json-api,json-stream,octet-stream,png,jpeg,gif,application/xhtml+xml,application/health+json`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")
	assert.JSONEq(t, expected, string(b))
}

func TestParseAcceptCommentErr(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`/@Accept unknown`, nil)
	assert.Error(t, err)
}

func TestParseProduceComment(t *testing.T) {
	t.Parallel()

	expected := `{
    "produces": [
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
		"application/health+json"
    ]
}`
	comment := `/@Produce json,xml,plain,html,mpfd,x-www-form-urlencoded,json-api,json-stream,octet-stream,png,jpeg,gif,application/health+json`
	operation := new(Operation)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")
	assert.JSONEq(t, expected, string(b))
}

func TestParseProduceCommentErr(t *testing.T) {
	t.Parallel()

	operation := new(Operation)
	err := operation.ParseComment(`/@Produce foo`, nil)
	assert.Error(t, err)
}

func TestParseRouterComment(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`/@Router /customer/get-wishlist/{wishlist_id} [get]`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}", operation.Path)
	assert.Equal(t, "GET", operation.HTTPMethod)
}

func TestParseRouterOnlySlash(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`// @Router / [get]`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/", operation.Path)
	assert.Equal(t, "GET", operation.HTTPMethod)
}

func TestParseRouterCommentWithPlusSign(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`/@Router /customer/get-wishlist/{proxy+} [post]`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{proxy+}", operation.Path)
	assert.Equal(t, "POST", operation.HTTPMethod)
}

func TestParseRouterCommentWithColonSign(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`/@Router /customer/get-wishlist/{wishlist_id}:move [post]`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}:move", operation.Path)
	assert.Equal(t, "POST", operation.HTTPMethod)
}

func TestParseRouterCommentNoColonSignAtPathStartErr(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`/@Router :customer/get-wishlist/{wishlist_id}:move [post]`, nil)
	assert.Error(t, err)
}

func TestParseRouterCommentMethodSeparationErr(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`/@Router /api/{id}|,*[get`, nil)
	assert.Error(t, err)
}

func TestParseRouterCommentMethodMissingErr(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`/@Router /customer/get-wishlist/{wishlist_id}`, nil)
	assert.Error(t, err)
}

func TestParseResponseCommentWithObjectType(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.OrderRow "Error message, if code != 200`
	operation := NewOperation(nil)
	operation.parser.addTestType("model.OrderRow")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "$ref": "#/definitions/model.OrderRow"
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithNestedPrimitiveType(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data=string,data2=int} "Error message, if code != 200`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "allOf": [
                    {
                        "$ref": "#/definitions/model.CommonHeader"
                    },
                    {
                        "type": "object",
                        "properties": {
                            "data": {
                                "type": "string"
                            },
                            "data2": {
                                "type": "integer"
                            }
                        }
                    }
                ]
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithNestedPrimitiveArrayType(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data=[]string,data2=[]int} "Error message, if code != 200`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "allOf": [
                    {
                        "$ref": "#/definitions/model.CommonHeader"
                    },
                    {
                        "type": "object",
                        "properties": {
                            "data": {
                                "type": "array",
                                "items": {
                                    "type": "string"
                                }
                            },
                            "data2": {
                                "type": "array",
                                "items": {
                                    "type": "integer"
                                }
                            }
                        }
                    }
                ]
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithNestedObjectType(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data=model.Payload,data2=model.Payload2} "Error message, if code != 200`
	operation := NewOperation(nil)
	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")
	operation.parser.addTestType("model.Payload2")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "allOf": [
                    {
                        "$ref": "#/definitions/model.CommonHeader"
                    },
                    {
                        "type": "object",
                        "properties": {
                            "data": {
                                "$ref": "#/definitions/model.Payload"
                            },
                            "data2": {
                                "$ref": "#/definitions/model.Payload2"
                            }
                        }
                    }
                ]
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithNestedArrayObjectType(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data=[]model.Payload,data2=[]model.Payload2} "Error message, if code != 200`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")
	operation.parser.addTestType("model.Payload2")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "allOf": [
                    {
                        "$ref": "#/definitions/model.CommonHeader"
                    },
                    {
                        "type": "object",
                        "properties": {
                            "data": {
                                "type": "array",
                                "items": {
                                    "$ref": "#/definitions/model.Payload"
                                }
                            },
                            "data2": {
                                "type": "array",
                                "items": {
                                    "$ref": "#/definitions/model.Payload2"
                                }
                            }
                        }
                    }
                ]
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithNestedFields(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data1=int,data2=[]int,data3=model.Payload,data4=[]model.Payload} "Error message, if code != 200`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "allOf": [
                    {
                        "$ref": "#/definitions/model.CommonHeader"
                    },
                    {
                        "type": "object",
                        "properties": {
                            "data1": {
                                "type": "integer"
                            },
                            "data2": {
                                "type": "array",
                                "items": {
                                    "type": "integer"
                                }
                            },
                            "data3": {
                                "$ref": "#/definitions/model.Payload"
                            },
                            "data4": {
                                "type": "array",
                                "items": {
                                    "$ref": "#/definitions/model.Payload"
                                }
                            }
                        }
                    }
                ]
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithDeepNestedFields(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data1=int,data2=[]int,data3=model.Payload{data1=int,data2=model.DeepPayload},data4=[]model.Payload{data1=[]int,data2=[]model.DeepPayload}} "Error message, if code != 200`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")
	operation.parser.addTestType("model.DeepPayload")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "allOf": [
                    {
                        "$ref": "#/definitions/model.CommonHeader"
                    },
                    {
                        "type": "object",
                        "properties": {
                            "data1": {
                                "type": "integer"
                            },
                            "data2": {
                                "type": "array",
                                "items": {
                                    "type": "integer"
                                }
                            },
                            "data3": {
                                "allOf": [
                                    {
                                        "$ref": "#/definitions/model.Payload"
                                    },
                                    {
                                        "type": "object",
                                        "properties": {
                                            "data1": {
                                                "type": "integer"
                                            },
                                            "data2": {
                                                "$ref": "#/definitions/model.DeepPayload"
                                            }
                                        }
                                    }
                                ]
                            },
                            "data4": {
                                "type": "array",
                                "items": {
                                    "allOf": [
                                        {
                                            "$ref": "#/definitions/model.Payload"
                                        },
                                        {
                                            "type": "object",
                                            "properties": {
                                                "data1": {
                                                    "type": "array",
                                                    "items": {
                                                        "type": "integer"
                                                    }
                                                },
                                                "data2": {
                                                    "type": "array",
                                                    "items": {
                                                        "$ref": "#/definitions/model.DeepPayload"
                                                    }
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    }
                ]
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithNestedArrayMapFields(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} []map[string]model.CommonHeader{data1=[]map[string]model.Payload,data2=map[string][]int} "Error message, if code != 200`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "type": "array",
                "items": {
                    "type": "object",
                    "additionalProperties": {
                        "allOf": [
                            {
                                "$ref": "#/definitions/model.CommonHeader"
                            },
                            {
                                "type": "object",
                                "properties": {
                                    "data1": {
                                        "type": "array",
                                        "items": {
                                            "type": "object",
                                            "additionalProperties": {
                                                "$ref": "#/definitions/model.Payload"
                                            }
                                        }
                                    },
                                    "data2": {
                                        "type": "object",
                                        "additionalProperties": {
                                            "type": "array",
                                            "items": {
                                                "type": "integer"
                                            }
                                        }
                                    }
                                }
                            }
                        ]
                    }
                }
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithObjectTypeInSameFile(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	operation.parser.addTestType("swag.testOwner")

	fset := token.NewFileSet()
	astFile, err := goparser.ParseFile(fset, "operation_test.go", `package swag
	type testOwner struct {

	}
	`, goparser.ParseComments)
	assert.NoError(t, err)

	err = operation.ParseComment(`@Success 200 {object} testOwner "Error message, if code != 200"`, astFile)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "$ref": "#/definitions/swag.testOwner"
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithObjectTypeAnonymousField(t *testing.T) {
	//TODO: test Anonymous
}

func TestParseResponseCommentWithObjectTypeErr(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	operation.parser.addTestType("model.notexist")

	err := operation.ParseComment(`@Success 200 {object} model.OrderRow "Error message, if code != 200"`, nil)
	assert.Error(t, err)
}

func TestParseResponseCommentWithArrayType(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	operation.parser.addTestType("model.OrderRow")
	err := operation.ParseComment(`@Success 200 {array} model.OrderRow "Error message, if code != 200`, nil)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)
	assert.Equal(t, spec.StringOrArray{"array"}, response.Schema.Type)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "type": "array",
                "items": {
                    "$ref": "#/definitions/model.OrderRow"
                }
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithBasicType(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Success 200 {string} string "it's ok'"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "it's ok'",
            "schema": {
                "type": "string"
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithBasicTypeAndCodes(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Success 200,201,default {string} string "it's ok"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "it's ok",
            "schema": {
                "type": "string"
            }
        },
        "201": {
            "description": "it's ok",
            "schema": {
                "type": "string"
            }
        },
        "default": {
            "description": "it's ok",
            "schema": {
                "type": "string"
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseEmptyResponseComment(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Success 200 "it is ok"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "it is ok"
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseEmptyResponseCommentWithCodes(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Success 200,201,default "it is ok"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "it is ok"
        },
        "201": {
            "description": "it is ok"
        },
        "default": {
            "description": "it is ok"
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentWithHeader(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Success 200 "it's ok"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	err = operation.ParseComment(`@Header 200 {string} Token "qwerty"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "it's ok",
            "headers": {
                "Token": {
                    "type": "string",
                    "description": "qwerty"
                }
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))

	err = operation.ParseComment(`@Header 200 "Mallformed"`, nil)
	assert.Error(t, err, "ParseComment should not fail")
}

func TestParseResponseCommentWithHeaderForCodes(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	err := operation.ParseComment(`@Success 200,201,default "it's ok"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	err = operation.ParseComment(`@Header 200,201,default {string} Token "qwerty"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	err = operation.ParseComment(`@Header all {string} Token2 "qwerty"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "it's ok",
            "headers": {
                "Token": {
                    "type": "string",
                    "description": "qwerty"
                },
                "Token2": {
                    "type": "string",
                    "description": "qwerty"
                }
            }
        },
        "201": {
            "description": "it's ok",
            "headers": {
                "Token": {
                    "type": "string",
                    "description": "qwerty"
                },
                "Token2": {
                    "type": "string",
                    "description": "qwerty"
                }
            }
        },
        "default": {
            "description": "it's ok",
            "headers": {
                "Token": {
                    "type": "string",
                    "description": "qwerty"
                },
                "Token2": {
                    "type": "string",
                    "description": "qwerty"
                }
            }
        }
    }
}`
	assert.Equal(t, expected, string(b))

	err = operation.ParseComment(`@Header 200 "Mallformed"`, nil)
	assert.Error(t, err, "ParseComment should not fail")
}

func TestParseEmptyResponseOnlyCode(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Success 200`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": ""
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseEmptyResponseOnlyCodes(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Success 200,201,default`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": ""
        },
        "201": {
            "description": ""
        },
        "default": {
            "description": ""
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentParamMissing(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	err := operation.ParseComment(`@Success notIntCode`, nil)
	assert.EqualError(t, err, `can not parse response comment "notIntCode"`)

	err = operation.ParseComment(`@Success notIntCode {string} string "it ok"`, nil)
	assert.EqualError(t, err, `can not parse response comment "notIntCode {string} string "it ok""`)

	err = operation.ParseComment(`@Success notIntCode "it ok"`, nil)
	assert.EqualError(t, err, `can not parse response comment "notIntCode "it ok""`)
}

// Test ParseParamComment.
func TestParseParamCommentByPathType(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id path int true "Some ID"`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "type": "integer",
            "description": "Some ID",
            "name": "some_id",
            "in": "path",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

// Test ParseParamComment Query Params.
func TestParseParamCommentBodyArray(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param names body []string true "Users List"`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "description": "Users List",
            "name": "names",
            "in": "body",
            "required": true,
            "schema": {
                "type": "array",
                "items": {
                    "type": "string"
                }
            }
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

// Test ParseParamComment Query Params.
func TestParseParamCommentQueryArray(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param names query []string true "Users List"`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "type": "array",
            "items": {
                "type": "string"
            },
            "description": "Users List",
            "name": "names",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

// Test ParseParamComment Query Params.
func TestParseParamCommentQueryArrayFormat(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param names query []string true "Users List" collectionFormat(multi)`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "type": "array",
            "items": {
                "type": "string"
            },
            "collectionFormat": "multi",
            "description": "Users List",
            "name": "names",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByID(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param unsafe_id[lte] query int true "Unsafe query param"`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "type": "integer",
            "description": "Unsafe query param",
            "name": "unsafe_id[lte]",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByQueryType(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id query int true "Some ID"`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "type": "integer",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByBodyType(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	operation.parser.addTestType("model.OrderRow")

	err := operation.ParseComment(`@Param some_id body model.OrderRow true "Some ID"`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "description": "Some ID",
            "name": "some_id",
            "in": "body",
            "required": true,
            "schema": {
                "$ref": "#/definitions/model.OrderRow"
            }
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByBodyTypeWithDeepNestedFields(t *testing.T) {
	t.Parallel()

	comment := `@Param body body model.CommonHeader{data=string,data2=int} true "test deep"`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Len(t, operation.Parameters, 1)
	assert.Equal(t, "test deep", operation.Parameters[0].Description)
	assert.True(t, operation.Parameters[0].Required)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "description": "test deep",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
                "allOf": [
                    {
                        "$ref": "#/definitions/model.CommonHeader"
                    },
                    {
                        "type": "object",
                        "properties": {
                            "data": {
                                "type": "string"
                            },
                            "data2": {
                                "type": "integer"
                            }
                        }
                    }
                ]
            }
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByBodyTypeArrayOfPrimitiveGo(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id body []int true "Some ID"`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "description": "Some ID",
            "name": "some_id",
            "in": "body",
            "required": true,
            "schema": {
                "type": "array",
                "items": {
                    "type": "integer"
                }
            }
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByBodyTypeArrayOfPrimitiveGoWithDeepNestedFields(t *testing.T) {
	t.Parallel()

	comment := `@Param body body []model.CommonHeader{data=string,data2=int} true "test deep"`
	operation := NewOperation(nil)
	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Len(t, operation.Parameters, 1)
	assert.Equal(t, "test deep", operation.Parameters[0].Description)
	assert.True(t, operation.Parameters[0].Required)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "description": "test deep",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
                "type": "array",
                "items": {
                    "allOf": [
                        {
                            "$ref": "#/definitions/model.CommonHeader"
                        },
                        {
                            "type": "object",
                            "properties": {
                                "data": {
                                    "type": "string"
                                },
                                "data2": {
                                    "type": "integer"
                                }
                            }
                        }
                    ]
                }
            }
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByBodyTypeErr(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	operation.parser.addTestType("model.notexist")
	err := operation.ParseComment(`@Param some_id body model.OrderRow true "Some ID"`, nil)
	assert.Error(t, err)
}

func TestParseParamCommentByFormDataType(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	err := operation.ParseComment(`@Param file formData file true "this is a test file"`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "type": "file",
            "description": "this is a test file",
            "name": "file",
            "in": "formData",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByFormDataTypeUint64(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	err := operation.ParseComment(`@Param file formData uint64 true "this is a test file"`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "type": "integer",
            "description": "this is a test file",
            "name": "file",
            "in": "formData",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByNotSupportedType(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id not_supported int true "Some ID"`, nil)
	assert.Error(t, err)
}

func TestParseParamCommentNotMatch(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id body mock true`, nil)
	assert.Error(t, err)
}

func TestParseParamCommentByEnums(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id query string true "Some ID" Enums(A, B, C)`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "enum": [
                "A",
                "B",
                "C"
            ],
            "type": "string",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))

	operation = NewOperation(nil)
	err = operation.ParseComment(`@Param some_id query int true "Some ID" Enums(1, 2, 3)`, nil)
	assert.NoError(t, err)

	b, _ = json.MarshalIndent(operation, "", "    ")

	expected = `{
    "parameters": [
        {
            "enum": [
                1,
                2,
                3
            ],
            "type": "integer",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))

	operation = NewOperation(nil)
	err = operation.ParseComment(`@Param some_id query number true "Some ID" Enums(1.1, 2.2, 3.3)`, nil)
	assert.NoError(t, err)

	b, _ = json.MarshalIndent(operation, "", "    ")

	expected = `{
    "parameters": [
        {
            "enum": [
                1.1,
                2.2,
                3.3
            ],
            "type": "number",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))

	operation = NewOperation(nil)
	err = operation.ParseComment(`@Param some_id query bool true "Some ID" Enums(true, false)`, nil)
	assert.NoError(t, err)

	b, _ = json.MarshalIndent(operation, "", "    ")

	expected = `{
    "parameters": [
        {
            "enum": [
                true,
                false
            ],
            "type": "boolean",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))

	operation = NewOperation(nil)

	assert.Error(t, operation.ParseComment(`@Param some_id query int true "Some ID" Enums(A, B, C)`, nil))
	assert.Error(t, operation.ParseComment(`@Param some_id query number true "Some ID" Enums(A, B, C)`, nil))
	assert.Error(t, operation.ParseComment(`@Param some_id query boolean true "Some ID" Enums(A, B, C)`, nil))
	assert.Error(t, operation.ParseComment(`@Param some_id query Document true "Some ID" Enums(A, B, C)`, nil))
}

func TestParseParamCommentByMaxLength(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id query string true "Some ID" MaxLength(10)`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "maxLength": 10,
            "type": "string",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))

	err = operation.ParseComment(`@Param some_id query int true "Some ID" MaxLength(10)`, nil)
	assert.Error(t, err)

	err = operation.ParseComment(`@Param some_id query string true "Some ID" MaxLength(Goopher)`, nil)
	assert.Error(t, err)
}

func TestParseParamCommentByMinLength(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id query string true "Some ID" MinLength(10)`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "minLength": 10,
            "type": "string",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))

	assert.Error(t, operation.ParseComment(`@Param some_id query int true "Some ID" MinLength(10)`, nil))
	assert.Error(t, operation.ParseComment(`@Param some_id query string true "Some ID" MinLength(Goopher)`, nil))
}

func TestParseParamCommentByMinimum(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id query int true "Some ID" Minimum(10)`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "minimum": 10,
            "type": "integer",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))

	assert.NoError(t, operation.ParseComment(`@Param some_id query int true "Some ID" Mininum(10)`, nil))
	assert.Error(t, operation.ParseComment(`@Param some_id query string true "Some ID" Minimum(10)`, nil))
	assert.Error(t, operation.ParseComment(`@Param some_id query integer true "Some ID" Minimum(Goopher)`, nil))
}

func TestParseParamCommentByMaximum(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id query int true "Some ID" Maximum(10)`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "maximum": 10,
            "type": "integer",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))

	assert.NoError(t, operation.ParseComment(`@Param some_id query int true "Some ID" Maxinum(10)`, nil))
	assert.Error(t, operation.ParseComment(`@Param some_id query string true "Some ID" Maximum(10)`, nil))
	assert.Error(t, operation.ParseComment(`@Param some_id query integer true "Some ID" Maximum(Goopher)`, nil))
}

func TestParseParamCommentByDefault(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Param some_id query int true "Some ID" Default(10)`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "parameters": [
        {
            "type": "integer",
            "default": 10,
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseIdComment(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)
	err := operation.ParseComment(`@Id myOperationId`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "myOperationId", operation.ID)
}

func TestFindTypeDefCoreLib(t *testing.T) {
	t.Parallel()

	def, err := findTypeDef("net/http", "Request")
	assert.NoError(t, err)
	assert.NotNil(t, def)
}

func TestFindTypeDefExternalPkg(t *testing.T) {
	t.Parallel()

	def, err := findTypeDef("github.com/KyleBanks/depth", "Tree")
	assert.NoError(t, err)
	assert.NotNil(t, def)
}

func TestFindTypeDefInvalidPkg(t *testing.T) {
	t.Parallel()

	def, err := findTypeDef("does-not-exist", "foo")
	assert.Error(t, err)
	assert.Nil(t, def)
}

func TestParseSecurityComment(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	err := operation.ParseComment(`@Security OAuth2Implicit[read, write]`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "security": [
        {
            "OAuth2Implicit": [
                "read",
                "write"
            ]
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseMultiDescription(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	err := operation.ParseComment(`@Description line one`, nil)
	assert.NoError(t, err)

	err = operation.ParseComment(`@Tags multi`, nil)
	assert.NoError(t, err)

	err = operation.ParseComment(`@Description line two x`, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `"description": "line one\nline two x"`
	assert.Contains(t, string(b), expected)
}

func TestParseSummary(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	err := operation.ParseComment(`@summary line one`, nil)
	assert.NoError(t, err)

	err = operation.ParseComment(`@Summary line one`, nil)
	assert.NoError(t, err)
}

func TestParseDeprecationDescription(t *testing.T) {
	t.Parallel()

	operation := NewOperation(nil)

	err := operation.ParseComment(`@Deprecated`, nil)
	assert.NoError(t, err)

	if !operation.Deprecated {
		t.Error("Failed to parse @deprecated comment")
	}
}

func TestParseExtentions(t *testing.T) {
	t.Parallel()

	// Fail if there are no args for attributes.
	{
		comment := `@x-amazon-apigateway-integration`
		operation := NewOperation(nil)

		err := operation.ParseComment(comment, nil)
		assert.EqualError(t, err, "annotation @x-amazon-apigateway-integration need a value")
	}

	// Fail if args of attributes are broken.
	{
		comment := `@x-amazon-apigateway-integration ["broken"}]`
		operation := NewOperation(nil)

		err := operation.ParseComment(comment, nil)
		assert.EqualError(t, err, "annotation @x-amazon-apigateway-integration need a valid json value")
	}

	// OK
	{
		comment := `@x-amazon-apigateway-integration {"uri": "${some_arn}", "passthroughBehavior": "when_no_match", "httpMethod": "POST", "type": "aws_proxy"}`
		operation := NewOperation(nil)

		err := operation.ParseComment(comment, nil)
		assert.NoError(t, err)

		expected := `{
    "x-amazon-apigateway-integration": {
        "httpMethod": "POST",
        "passthroughBehavior": "when_no_match",
        "type": "aws_proxy",
        "uri": "${some_arn}"
    }
}`
		b, _ := json.MarshalIndent(operation, "", "    ")
		assert.Equal(t, expected, string(b))
	}

	// Test x-tagGroups
	{
		comment := `@x-tagGroups [{"name":"Natural Persons","tags":["Person","PersonRisk","PersonDocuments"]}]`
		operation := NewOperation(nil)

		err := operation.ParseComment(comment, nil)
		assert.NoError(t, err)

		expected := `{
    "x-tagGroups": [
        {
            "name": "Natural Persons",
            "tags": [
                "Person",
                "PersonRisk",
                "PersonDocuments"
            ]
        }
    ]
}`

		b, _ := json.MarshalIndent(operation, "", "    ")
		assert.Equal(t, expected, string(b))
	}
}

func TestParseCodeSamples(t *testing.T) {
	t.Parallel()

	t.Run("Find sample by file", func(t *testing.T) {
		operation := NewOperation(nil, SetCodeExampleFilesDirectory("testdata/code_examples"))
		operation.Summary = "example"

		err := operation.ParseComment(`@x-codeSamples file`, nil)
		assert.NoError(t, err, "no error should be thrown")

		b, _ := json.MarshalIndent(operation, "", "    ")

		expected := `{
    "summary": "example",
    "x-codeSamples": {
        "lang": "JavaScript",
        "source": "console.log('Hello World');"
    }
}`
		assert.Equal(t, expected, string(b))
	})

	t.Run("Example file not found", func(t *testing.T) {
		operation := NewOperation(nil, SetCodeExampleFilesDirectory("testdata/code_examples"))
		operation.Summary = "exampel"

		err := operation.ParseComment(`@x-codeSamples file`, nil)
		assert.Error(t, err, "error was expected, as file does not exist")
	})
}
