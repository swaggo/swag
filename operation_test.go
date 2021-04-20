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
	operation := NewOperation(nil)
	err := operation.ParseComment("//", nil)

	assert.NoError(t, err)
}

func TestParseTagsComment(t *testing.T) {
	expected := `{
    "tags": [
        "pet",
        "store",
        "user"
    ]
}`
	comment := `/@Tags pet, store,user`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(operation, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestParseAcceptComment(t *testing.T) {
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
	comment := `/@Accept unknown`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseProduceComment(t *testing.T) {
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
	comment := `/@Produce foo`
	operation := new(Operation)
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseRouterComment(t *testing.T) {
	comment := `/@Router /customer/get-wishlist/{wishlist_id} [get]`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}", operation.Path)
	assert.Equal(t, "GET", operation.HTTPMethod)
}

func TestParseRouterOnlySlash(t *testing.T) {
	comment := `// @Router / [get]`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/", operation.Path)
	assert.Equal(t, "GET", operation.HTTPMethod)
}

func TestParseRouterCommentWithPlusSign(t *testing.T) {
	comment := `/@Router /customer/get-wishlist/{proxy+} [post]`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{proxy+}", operation.Path)
	assert.Equal(t, "POST", operation.HTTPMethod)
}

func TestParseRouterCommentWithColonSign(t *testing.T) {
	comment := `/@Router /customer/get-wishlist/{wishlist_id}:move [post]`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}:move", operation.Path)
	assert.Equal(t, "POST", operation.HTTPMethod)
}

func TestParseRouterCommentNoColonSignAtPathStartErr(t *testing.T) {
	comment := `/@Router :customer/get-wishlist/{wishlist_id}:move [post]`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseRouterCommentMethodSeparationErr(t *testing.T) {
	comment := `/@Router /api/{id}|,*[get`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseRouterCommentMethodMissingErr(t *testing.T) {
	comment := `/@Router /customer/get-wishlist/{wishlist_id}`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseResponseCommentWithObjectType(t *testing.T) {
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
	comment := `@Success 200 {object} testOwner "Error message, if code != 200"`
	operation := NewOperation(nil)

	operation.parser.addTestType("swag.testOwner")

	fset := token.NewFileSet()
	astFile, err := goparser.ParseFile(fset, "operation_test.go", `package swag
	type testOwner struct {

	}
	`, goparser.ParseComments)
	assert.NoError(t, err)

	err = operation.ParseComment(comment, astFile)
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
	comment := `@Success 200 {object} model.OrderRow "Error message, if code != 200"`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.notexist")

	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseResponseCommentWithArrayType(t *testing.T) {
	comment := `@Success 200 {array} model.OrderRow "Error message, if code != 200`
	operation := NewOperation(nil)
	operation.parser.addTestType("model.OrderRow")
	err := operation.ParseComment(comment, nil)
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
	comment := `@Success 200 {string} string "it's ok'"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
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
	comment := `@Success 200,201,default {string} string "it's ok"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
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
	comment := `@Success 200 "it is ok"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
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
	comment := `@Success 200,201,default "it is ok"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
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
	comment := `@Success 200 "it's ok"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	comment = `@Header 200 {string} Token "qwerty"`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, err := json.MarshalIndent(operation, "", "    ")
	assert.NoError(t, err)

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

	comment = `@Header 200 "Mallformed"`
	err = operation.ParseComment(comment, nil)
	assert.Error(t, err, "ParseComment should not fail")
}

func TestParseResponseCommentWithHeaderForCodes(t *testing.T) {
	operation := NewOperation(nil)

	comment := `@Success 200,201,default "it's ok"`
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	comment = `@Header 200,201,default {string} Token "qwerty"`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	comment = `@Header all {string} Token2 "qwerty"`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, err := json.MarshalIndent(operation, "", "    ")
	assert.NoError(t, err)

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

	comment = `@Header 200 "Mallformed"`
	err = operation.ParseComment(comment, nil)
	assert.Error(t, err, "ParseComment should not fail")
}

func TestParseEmptyResponseOnlyCode(t *testing.T) {
	comment := `@Success 200`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
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
	comment := `@Success 200,201,default`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)
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
	operation := NewOperation(nil)

	paramLenErrComment := `@Success notIntCode`
	paramLenErr := operation.ParseComment(paramLenErrComment, nil)
	assert.EqualError(t, paramLenErr, `can not parse response comment "notIntCode"`)

	paramLenErrComment = `@Success notIntCode {string} string "it ok"`
	paramLenErr = operation.ParseComment(paramLenErrComment, nil)
	assert.EqualError(t, paramLenErr, `can not parse response comment "notIntCode {string} string "it ok""`)

	paramLenErrComment = `@Success notIntCode "it ok"`
	paramLenErr = operation.ParseComment(paramLenErrComment, nil)
	assert.EqualError(t, paramLenErr, `can not parse response comment "notIntCode "it ok""`)
}

// Test ParseParamComment
func TestParseParamCommentByPathType(t *testing.T) {
	comment := `@Param some_id path int true "Some ID"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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

// Test ParseParamComment Query Params
func TestParseParamCommentBodyArray(t *testing.T) {
	comment := `@Param names body []string true "Users List"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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

// Test ParseParamComment Query Params
func TestParseParamCommentQueryArray(t *testing.T) {
	comment := `@Param names query []string true "Users List"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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

// Test ParseParamComment Query Params
func TestParseParamCommentQueryArrayFormat(t *testing.T) {
	comment := `@Param names query []string true "Users List" collectionFormat(multi)`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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
	comment := `@Param unsafe_id[lte] query int true "Unsafe query param"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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
	comment := `@Param some_id query int true "Some ID"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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
	comment := `@Param some_id body model.OrderRow true "Some ID"`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.OrderRow")
	err := operation.ParseComment(comment, nil)

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
	comment := `@Param body body model.CommonHeader{data=string,data2=int} true "test deep"`
	operation := NewOperation(nil)

	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Len(t, operation.Parameters, 1)
	assert.Equal(t, "test deep", operation.Parameters[0].Description)
	assert.True(t, operation.Parameters[0].Required)

	b, err := json.MarshalIndent(operation, "", "    ")
	assert.NoError(t, err)
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
	comment := `@Param some_id body []int true "Some ID"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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
	comment := `@Param body body []model.CommonHeader{data=string,data2=int} true "test deep"`
	operation := NewOperation(nil)
	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Len(t, operation.Parameters, 1)
	assert.Equal(t, "test deep", operation.Parameters[0].Description)
	assert.True(t, operation.Parameters[0].Required)

	b, err := json.MarshalIndent(operation, "", "    ")
	assert.NoError(t, err)
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
	comment := `@Param some_id body model.OrderRow true "Some ID"`
	operation := NewOperation(nil)
	operation.parser.addTestType("model.notexist")
	err := operation.ParseComment(comment, nil)

	assert.Error(t, err)
}

func TestParseParamCommentByFormDataType(t *testing.T) {
	comment := `@Param file formData file true "this is a test file"`
	operation := NewOperation(nil)

	err := operation.ParseComment(comment, nil)
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
	comment := `@Param file formData uint64 true "this is a test file"`
	operation := NewOperation(nil)

	err := operation.ParseComment(comment, nil)
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
	comment := `@Param some_id not_supported int true "Some ID"`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

	assert.Error(t, err)
}

func TestParseParamCommentNotMatch(t *testing.T) {
	comment := `@Param some_id body mock true`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

	assert.Error(t, err)
}

func TestParseParamCommentByEnums(t *testing.T) {
	comment := `@Param some_id query string true "Some ID" Enums(A, B, C)`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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

	comment = `@Param some_id query int true "Some ID" Enums(1, 2, 3)`
	operation = NewOperation(nil)
	err = operation.ParseComment(comment, nil)

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

	comment = `@Param some_id query number true "Some ID" Enums(1.1, 2.2, 3.3)`
	operation = NewOperation(nil)
	err = operation.ParseComment(comment, nil)

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

	comment = `@Param some_id query bool true "Some ID" Enums(true, false)`
	operation = NewOperation(nil)
	err = operation.ParseComment(comment, nil)

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

	comment = `@Param some_id query int true "Some ID" Enums(A, B, C)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query number true "Some ID" Enums(A, B, C)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query boolean true "Some ID" Enums(A, B, C)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query Document true "Some ID" Enums(A, B, C)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByMaxLength(t *testing.T) {
	comment := `@Param some_id query string true "Some ID" MaxLength(10)`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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

	comment = `@Param some_id query int true "Some ID" MaxLength(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query string true "Some ID" MaxLength(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByMinLength(t *testing.T) {
	comment := `@Param some_id query string true "Some ID" MinLength(10)`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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

	comment = `@Param some_id query int true "Some ID" MinLength(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query string true "Some ID" MinLength(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByMinimum(t *testing.T) {
	comment := `@Param some_id query int true "Some ID" Minimum(10)`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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

	comment = `@Param some_id query int true "Some ID" Mininum(10)`
	assert.NoError(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query string true "Some ID" Minimum(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query integer true "Some ID" Minimum(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByMaximum(t *testing.T) {
	comment := `@Param some_id query int true "Some ID" Maximum(10)`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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

	comment = `@Param some_id query int true "Some ID" Maxinum(10)`
	assert.NoError(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query string true "Some ID" Maximum(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query integer true "Some ID" Maximum(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))

}

func TestParseParamCommentByDefault(t *testing.T) {
	comment := `@Param some_id query int true "Some ID" Default(10)`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

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

func TestParseParamArrayWithEnums(t *testing.T) {
	comment := `@Param field query []string true "An enum collection" collectionFormat(csv) enums(also,valid)`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

	assert.NoError(t, err)
	b, _ := json.MarshalIndent(operation, "", "    ")
	expected := `{
    "parameters": [
        {
            "type": "array",
            "items": {
                "enum": [
                    "also",
                    "valid"
                ],
                "type": "string"
            },
            "collectionFormat": "csv",
            "description": "An enum collection",
            "name": "field",
            "in": "query",
            "required": true
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseIdComment(t *testing.T) {
	comment := `@Id myOperationId`
	operation := NewOperation(nil)
	err := operation.ParseComment(comment, nil)

	assert.NoError(t, err)
	assert.Equal(t, "myOperationId", operation.ID)
}

func TestFindTypeDefCoreLib(t *testing.T) {
	spec, err := findTypeDef("net/http", "Request")
	assert.NoError(t, err)
	assert.NotNil(t, spec)
}

func TestFindTypeDefExternalPkg(t *testing.T) {
	spec, err := findTypeDef("github.com/KyleBanks/depth", "Tree")
	assert.NoError(t, err)
	assert.NotNil(t, spec)
}

func TestFindTypeDefInvalidPkg(t *testing.T) {
	spec, err := findTypeDef("does-not-exist", "foo")
	assert.Error(t, err)
	assert.Nil(t, spec)
}

func TestParseSecurityComment(t *testing.T) {
	comment := `@Security OAuth2Implicit[read, write]`
	operation := NewOperation(nil)

	err := operation.ParseComment(comment, nil)
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
	comment := `@Description line one`
	operation := NewOperation(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	comment = `@Tags multi`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	comment = `@Description line two x`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `"description": "line one\nline two x"`
	assert.Contains(t, string(b), expected)
}

func TestParseSummary(t *testing.T) {
	comment := `@summary line one`
	operation := NewOperation(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	comment = `@Summary line one`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err)
}

func TestParseDeprecationDescription(t *testing.T) {
	comment := `@Deprecated`
	operation := NewOperation(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	if !operation.Deprecated {
		t.Error("Failed to parse @deprecated comment")
	}
}

func TestParseExtentions(t *testing.T) {
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
	t.Run("Find sample by file", func(t *testing.T) {
		comment := `@x-codeSamples file`
		operation := NewOperation(nil, SetCodeExampleFilesDirectory("testdata/code_examples"))
		operation.Summary = "example"

		err := operation.ParseComment(comment, nil)
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
		comment := `@x-codeSamples file`
		operation := NewOperation(nil, SetCodeExampleFilesDirectory("testdata/code_examples"))
		operation.Summary = "exampel"

		err := operation.ParseComment(comment, nil)
		assert.Error(t, err, "error was expected, as file does not exist")
	})
}
