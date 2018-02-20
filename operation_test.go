package swag

import (
	"encoding/json"
	"go/ast"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

func TestParseEmptyComment(t *testing.T) {
	operation := NewOperation()
	err := operation.ParseComment("//")

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
	operation := NewOperation()
	err := operation.ParseComment(comment)
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
        "application/vnd.api+json"
    ]
}`
	comment := `/@Accept json,xml,plain,html,mpfd,json-api`
	operation := NewOperation()
	err := operation.ParseComment(comment)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(operation, "", "    ")
	assert.JSONEq(t, expected, string(b))

}

func TestParseAcceptCommentErr(t *testing.T) {
	comment := `/@Accept unknown`
	operation := NewOperation()
	err := operation.ParseComment(comment)
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
        "application/vnd.api+json"
    ]
}`
	comment := `/@Produce json,xml,plain,html,mpfd,json-api`
	operation := new(Operation)
	operation.ParseComment(comment)
	b, _ := json.MarshalIndent(operation, "", "    ")
	assert.JSONEq(t, expected, string(b))
}

func TestParseProduceCommentErr(t *testing.T) {
	comment := `/@Produce foo`
	operation := new(Operation)
	err := operation.ParseComment(comment)
	assert.Error(t, err)
}

func TestParseRouterComment(t *testing.T) {
	comment := `/@Router /customer/get-wishlist/{wishlist_id} [get]`
	operation := NewOperation()
	err := operation.ParseComment(comment)
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}", operation.Path)
	assert.Equal(t, "GET", operation.HttpMethod)
}

func TestParseRouterCommentOccursErr(t *testing.T) {
	comment := `/@Router /customer/get-wishlist/{wishlist_id}`
	operation := NewOperation()
	err := operation.ParseComment(comment)
	assert.Error(t, err)
}

func TestParseResponseCommentWithObjectType(t *testing.T) {
	comment := `@Success 200 {object} model.OrderRow "Error message, if code != 200`
	operation := NewOperation()
	operation.parser = New()

	operation.parser.TypeDefinitions["model"] = make(map[string]*ast.TypeSpec)
	operation.parser.TypeDefinitions["model"]["OrderRow"] = &ast.TypeSpec{}

	err := operation.ParseComment(comment)
	assert.NoError(t, err)

	response := operation.Responses.StatusCodeResponses[200]
	assert.Equal(t, `Error message, if code != 200`, response.Description)
	assert.Equal(t, spec.StringOrArray{"object"}, response.Schema.Type)

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "Error message, if code != 200",
            "schema": {
                "type": "object",
                "$ref": "#/definitions/model.OrderRow"
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
	operation := NewOperation()
	operation.parser = New()

	operation.parser.TypeDefinitions["model"] = make(map[string]*ast.TypeSpec)
	operation.parser.TypeDefinitions["model"]["notexist"] = &ast.TypeSpec{}

	err := operation.ParseComment(comment)
	assert.Error(t, err)
}

func TestParseResponseCommentWithArrayType(t *testing.T) {
	comment := `@Success 200 {array} model.OrderRow "Error message, if code != 200`
	operation := NewOperation()
	err := operation.ParseComment(comment)
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
	operation := NewOperation()
	operation.ParseComment(comment)
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

func TestParseEmptyResponseComment(t *testing.T) {
	comment := `@Success 200 "it's ok"`
	operation := NewOperation()
	operation.ParseComment(comment)
	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {
            "description": "it's ok"
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentParamMissing(t *testing.T) {
	operation := NewOperation()

	paramLenErrComment := `@Success notIntCode {string}`
	paramLenErr := operation.ParseComment(paramLenErrComment)
	assert.EqualError(t, paramLenErr, `can not parse response comment "notIntCode {string}"
can not parse empty response comment "notIntCode {string}"`)
}

// Test ParseParamComment
func TestParseParamCommentByPathType(t *testing.T) {
	comment := `@Param some_id path int true "Some ID"`
	operation := NewOperation()
	err := operation.ParseComment(comment)

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

func TestParseParamCommentByQueryType(t *testing.T) {
	comment := `@Param some_id query int true "Some ID"`
	operation := NewOperation()
	err := operation.ParseComment(comment)

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
	operation := NewOperation()
	operation.parser = New()

	operation.parser.TypeDefinitions["model"] = make(map[string]*ast.TypeSpec)
	operation.parser.TypeDefinitions["model"]["OrderRow"] = &ast.TypeSpec{}
	err := operation.ParseComment(comment)

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
                "type": "object",
                "$ref": "#/definitions/model.OrderRow"
            }
        }
    ]
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByBodyTypeErr(t *testing.T) {
	comment := `@Param some_id body model.OrderRow true "Some ID"`
	operation := NewOperation()
	operation.parser = New()

	operation.parser.TypeDefinitions["model"] = make(map[string]*ast.TypeSpec)
	operation.parser.TypeDefinitions["model"]["notexist"] = &ast.TypeSpec{}
	err := operation.ParseComment(comment)

	assert.Error(t, err)
}

func TestParseParamCommentByFormDataType(t *testing.T) {
	comment := `@Param   file formData file true  "this is a test file"`
	operation := NewOperation()
	operation.parser = New()

	err := operation.ParseComment(comment)
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

func TestParseParamCommentNotMatch(t *testing.T) {
	comment := `@Param some_id body mock true`
	operation := NewOperation()
	err := operation.ParseComment(comment)

	assert.Error(t, err)
}

func TestParseIdComment(t *testing.T) {
	comment := `@Id myOperationId`
	operation := NewOperation()
	err := operation.ParseComment(comment)

	assert.NoError(t, err)
	assert.Equal(t, "myOperationId", operation.ID)
}
