package parse

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOperation_ParseAcceptComment(t *testing.T) {
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

	b, _ := json.MarshalIndent(operation, "", "    ")
	fmt.Printf("%+v", string(b))
}

func TestOperation_ParseResponseCommentWithObjectType(t *testing.T) {
	//@Success 200 {object} model.OrderRow "Error message, if code != 200
	operation := NewOperation()
	err := operation.ParseResponseComment(`200 {object} model.OrderRow "Error message, if code != 200"`)
	assert.NoError(t, err)
	response := operation.Responses.StatusCodeResponses[200]
	fmt.Printf("%+v\n", operation)
	assert.Equal(t, `Error message, if code != 200`, response.Description)
	assert.Equal(t, spec.StringOrArray{"object"}, response.Schema.Type)

	b, err := json.MarshalIndent(operation, "", "    ")

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
	operation := NewOperation()
	operation.ParseResponseComment(`200 {string} string "it's ok'"`)
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

func TestOperation_ParseComment(t *testing.T) {
	//operation := NewOperation()
	//
	//err := operation.ParseComment()
	//assert.NoError(t, err)
}

//func TestOperation_ParseParamComment(t *testing.T) {
//	operation := NewOperation()
//
//	path:="/testapi/{user}/{pwd}"
//	operation.ParseRouterParams(path)
//}
