package parser

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseAcceptComment(t *testing.T) {
	expected := `{
    "consumes": [
        "application/json",
        "text/xml",
        "text/plain",
        "text/html",
        "multipart/form-data"
    ]
}`
	operation := new(Operation)
	operation.ParseAcceptComment("json,xml,plain,html,mpfd")

	b, _ := json.MarshalIndent(operation, "", "    ")
	assert.Equal(t, expected, string(b))
}

func TestOperation_ParseProduceComment(t *testing.T) {
	expected := `{
    "produces": [
        "application/json",
        "text/xml",
        "text/plain",
        "text/html",
        "multipart/form-data"
    ]
}`

	operation := new(Operation)
	operation.ParseProduceComment("json,xml,plain,html,mpfd")
	b, _ := json.MarshalIndent(operation, "", "    ")
	fmt.Printf("%+v", string(b))
	assert.Equal(t, expected, string(b))
}

func TestOperation_ParseRouterComment(t *testing.T) {
	//@Router /customer/get-wishlist/{wishlist_id} [get]
	operation := NewOperation()
	err := operation.ParseRouterComment("/customer/get-wishlist/{wishlist_id} [get]")
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}", operation.Path)
	assert.Equal(t, "GET", operation.HttpMethod)
}

func TestParseResponseCommentWithObjectType(t *testing.T) {
	comment := `@Success 200 {object} model.OrderRow "Error message, if code != 200`
	operation := NewOperation()
	err := operation.ParseComment(comment)
	assert.NoError(t, err)
	response := operation.Responses.StatusCodeResponses[200]
	fmt.Printf("%+v\n", operation)
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

func TestParseResponseCommentWithBasicType(t *testing.T) {
	comment := `@Success 200 {string} string "it's ok'"`
	operation := NewOperation()
	operation.ParseComment(comment)
	b, _ := json.MarshalIndent(operation, "", "    ")
	fmt.Printf("%+v", string(b))

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
            "type": "int",
            "description": "Some ID",
            "name": "some_id",
            "in": "path",
            "required": true
        }
    ],
    "responses": {}
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
            "type": "int",
            "description": "Some ID",
            "name": "some_id",
            "in": "query",
            "required": true
        }
    ],
    "responses": {}
}`
	assert.Equal(t, expected, string(b))
}

func TestParseParamCommentByBodyType(t *testing.T) {
	//TODO: add tests coverage swagger definitions
	comment := `@Param some_id body mock true "Some ID"`
	operation := NewOperation()
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
                "type": "object"
            }
        }
    ],
    "responses": {}
}`
	assert.Equal(t, expected, string(b))
}