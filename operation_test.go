package swag

import (
	"encoding/json"
	"go/token"

	"go/ast"
	goparser "go/parser"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

func TestParseEmptyComment(t *testing.T) {
	operation := NewOperation()
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
	operation := NewOperation()
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
	operation := NewOperation()
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(operation, "", "    ")
	assert.JSONEq(t, expected, string(b))

}

func TestParseAcceptCommentErr(t *testing.T) {
	comment := `/@Accept unknown`
	operation := NewOperation()
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
	operation := NewOperation()
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}", operation.Path)
	assert.Equal(t, "GET", operation.HTTPMethod)
}

func TestParseRouterCommentWithPlusSign(t *testing.T) {
	comment := `/@Router /customer/get-wishlist/{proxy+} [post]`
	operation := NewOperation()
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/customer/get-wishlist/{proxy+}", operation.Path)
	assert.Equal(t, "POST", operation.HTTPMethod)
}

func TestParseRouterCommentOccursErr(t *testing.T) {
	comment := `/@Router /customer/get-wishlist/{wishlist_id}`
	operation := NewOperation()
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseResponseCommentWithObjectType(t *testing.T) {
	comment := `@Success 200 {object} model.OrderRow "Error message, if code != 200`
	operation := NewOperation()
	operation.parser = New()

	operation.parser.TypeDefinitions["model"] = make(map[string]*ast.TypeSpec)
	operation.parser.TypeDefinitions["model"]["OrderRow"] = &ast.TypeSpec{}

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

func TestParseResponseCommentWithObjectTypeAnonymousField(t *testing.T) {
	//TODO: test Anonymous
}

func TestParseResponseCommentWithObjectTypeErr(t *testing.T) {
	comment := `@Success 200 {object} model.OrderRow "Error message, if code != 200"`
	operation := NewOperation()
	operation.parser = New()

	operation.parser.TypeDefinitions["model"] = make(map[string]*ast.TypeSpec)
	operation.parser.TypeDefinitions["model"]["notexist"] = &ast.TypeSpec{}

	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseResponseCommentWithArrayType(t *testing.T) {
	comment := `@Success 200 {array} model.OrderRow "Error message, if code != 200`
	operation := NewOperation()
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
	operation := NewOperation()
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

func TestParseEmptyResponseComment(t *testing.T) {
	comment := `@Success 200 "it's ok"`
	operation := NewOperation()
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

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

func TestParseResponseCommentWithHeader(t *testing.T) {
	comment := `@Success 200 "it's ok"`
	operation := NewOperation()
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

func TestParseEmptyResponseOnlyCode(t *testing.T) {
	comment := `@Success 200`
	operation := NewOperation()
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	b, _ := json.MarshalIndent(operation, "", "    ")

	expected := `{
    "responses": {
        "200": {}
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseCommentParamMissing(t *testing.T) {
	operation := NewOperation()

	paramLenErrComment := `@Success notIntCode {string}`
	paramLenErr := operation.ParseComment(paramLenErrComment, nil)
	assert.EqualError(t, paramLenErr, `can not parse response comment "notIntCode {string}"`)
}

// Test ParseParamComment
func TestParseParamCommentByPathType(t *testing.T) {
	comment := `@Param some_id path int true "Some ID"`
	operation := NewOperation()
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

func TestParseParamCommentByID(t *testing.T) {
	comment := `@Param unsafe_id[lte] query int true "Unsafe query param"`
	operation := NewOperation()
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
	operation := NewOperation()
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
	operation := NewOperation()
	operation.parser = New()

	operation.parser.TypeDefinitions["model"] = make(map[string]*ast.TypeSpec)
	operation.parser.TypeDefinitions["model"]["OrderRow"] = &ast.TypeSpec{}
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
	err := operation.ParseComment(comment, nil)

	assert.Error(t, err)
}

func TestParseParamCommentByFormDataType(t *testing.T) {
	comment := `@Param file formData file true "this is a test file"`
	operation := NewOperation()
	operation.parser = New()

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
	operation := NewOperation()
	operation.parser = New()

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
	operation := NewOperation()
	err := operation.ParseComment(comment, nil)

	assert.Error(t, err)
}

func TestParseParamCommentNotMatch(t *testing.T) {
	comment := `@Param some_id body mock true`
	operation := NewOperation()
	err := operation.ParseComment(comment, nil)

	assert.Error(t, err)
}

func TestParseParamCommentByEnums(t *testing.T) {
	comment := `@Param some_id query string true "Some ID" Enums(A, B, C)`
	operation := NewOperation()
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
	operation = NewOperation()
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
	operation = NewOperation()
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
	operation = NewOperation()
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

	operation = NewOperation()

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
	operation := NewOperation()
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
	operation := NewOperation()
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

func TestParseParamCommentByMininum(t *testing.T) {
	comment := `@Param some_id query int true "Some ID" Mininum(10)`
	operation := NewOperation()
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

	comment = `@Param some_id query string true "Some ID" Mininum(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query integer true "Some ID" Mininum(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByMaxinum(t *testing.T) {
	comment := `@Param some_id query int true "Some ID" Maxinum(10)`
	operation := NewOperation()
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

	comment = `@Param some_id query string true "Some ID" Maxinum(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query integer true "Some ID" Maxinum(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))

}

func TestParseParamCommentByDefault(t *testing.T) {
	comment := `@Param some_id query int true "Some ID" Default(10)`
	operation := NewOperation()
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

	comment = `@Param some_id query time.Duration true "Some ID" Default(10)`
	operation = NewOperation()
	assert.NoError(t, operation.ParseComment(comment, nil))
}

func TestParseIdComment(t *testing.T) {
	comment := `@Id myOperationId`
	operation := NewOperation()
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
	spec, err := findTypeDef("github.com/stretchr/testify/assert", "Assertions")
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
	operation := NewOperation()
	operation.parser = New()
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
	operation := NewOperation()
	operation.parser = New()

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
	operation := NewOperation()
	operation.parser = New()

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	comment = `@Summary line one`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err)
}

func TestParseDeprecationDescription(t *testing.T) {
	comment := `@Deprecated`
	operation := NewOperation()
	operation.parser = New()

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	if !operation.Deprecated {
		t.Error("Failed to parse @deprecated comment")
	}
}

func TestRegisterSchemaType(t *testing.T) {
	operation := NewOperation()
	assert.NoError(t, operation.registerSchemaType("string", nil))

	fset := token.NewFileSet()
	astFile, err := goparser.ParseFile(fset, "main.go", `package main
	import "timer"
`, goparser.ParseComments)

	assert.NoError(t, err)

	operation.parser = New()
	assert.Error(t, operation.registerSchemaType("timer.Location", astFile))
}

func TestParseExtentions(t *testing.T) {
	// Fail if there are no args for attributes.
	{
		comment := `@x-amazon-apigateway-integration`
		operation := NewOperation()
		operation.parser = New()

		err := operation.ParseComment(comment, nil)
		assert.EqualError(t, err, "@x-amazon-apigateway-integration need a value")
	}

	// Fail if args of attributes are broken.
	{
		comment := `@x-amazon-apigateway-integration ["broken"}]`
		operation := NewOperation()
		operation.parser = New()

		err := operation.ParseComment(comment, nil)
		assert.EqualError(t, err, "@x-amazon-apigateway-integration need a valid json value")
	}

	// OK
	{
		comment := `@x-amazon-apigateway-integration {"uri": "${some_arn}", "passthroughBehavior": "when_no_match", "httpMethod": "POST", "type": "aws_proxy"}`
		operation := NewOperation()
		operation.parser = New()

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
}
