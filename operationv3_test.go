package swag

import (
	"encoding/json"
	goparser "go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sv-tools/openapi/spec"
)

var typeObject = spec.SingleOrArray[string](spec.SingleOrArray[string]{OBJECT})
var typeArray = spec.SingleOrArray[string](spec.SingleOrArray[string]{ARRAY})
var typeInteger = spec.SingleOrArray[string](spec.SingleOrArray[string]{INTEGER})
var typeString = spec.SingleOrArray[string](spec.SingleOrArray[string]{STRING})

func TestParseEmptyCommentV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(nil)
	err := operation.ParseComment("//", nil)

	require.NoError(t, err)
}

func TestParseTagsCommentV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(nil)
	err := operation.ParseComment(`/@Tags pet, store,user`, nil)
	require.NoError(t, err)
	assert.Equal(t, operation.Tags, []string{"pet", "store", "user"})
}

func TestParseRouterCommentV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router /customer/get-wishlist/{wishlist_id} [get]`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	assert.Len(t, operation.RouterProperties, 1)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}", operation.RouterProperties[0].Path)
	assert.Equal(t, "GET", operation.RouterProperties[0].HTTPMethod)
}

func TestParseRouterMultipleCommentsV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router /customer/get-wishlist/{wishlist_id} [get]`
	anotherComment := `/@Router /customer/get-the-wishlist/{wishlist_id} [post]`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	err = operation.ParseComment(anotherComment, nil)
	require.NoError(t, err)

	assert.Len(t, operation.RouterProperties, 2)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}", operation.RouterProperties[0].Path)
	assert.Equal(t, "GET", operation.RouterProperties[0].HTTPMethod)
	assert.Equal(t, "/customer/get-the-wishlist/{wishlist_id}", operation.RouterProperties[1].Path)
	assert.Equal(t, "POST", operation.RouterProperties[1].HTTPMethod)
}

func TestParseRouterOnlySlashV3(t *testing.T) {
	t.Parallel()

	comment := `// @Router / [get]`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)
	assert.Len(t, operation.RouterProperties, 1)
	assert.Equal(t, "/", operation.RouterProperties[0].Path)
	assert.Equal(t, "GET", operation.RouterProperties[0].HTTPMethod)
}

func TestParseRouterCommentWithPlusSignV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router /customer/get-wishlist/{proxy+} [post]`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)
	assert.Len(t, operation.RouterProperties, 1)
	assert.Equal(t, "/customer/get-wishlist/{proxy+}", operation.RouterProperties[0].Path)
	assert.Equal(t, "POST", operation.RouterProperties[0].HTTPMethod)
}

func TestParseRouterCommentWithDollarSignV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router /customer/get-wishlist/{wishlist_id}$move [post]`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)
	assert.Len(t, operation.RouterProperties, 1)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}$move", operation.RouterProperties[0].Path)
	assert.Equal(t, "POST", operation.RouterProperties[0].HTTPMethod)
}

func TestParseRouterCommentNoDollarSignAtPathStartErrV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router $customer/get-wishlist/{wishlist_id}$move [post]`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseRouterCommentWithColonSignV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router /customer/get-wishlist/{wishlist_id}:move [post]`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)
	assert.Len(t, operation.RouterProperties, 1)
	assert.Equal(t, "/customer/get-wishlist/{wishlist_id}:move", operation.RouterProperties[0].Path)
	assert.Equal(t, "POST", operation.RouterProperties[0].HTTPMethod)
}

func TestParseRouterCommentNoColonSignAtPathStartErrV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router :customer/get-wishlist/{wishlist_id}:move [post]`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseRouterCommentMethodSeparationErrV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router /api/{id}|,*[get`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseRouterCommentMethodMissingErrV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router /customer/get-wishlist/{wishlist_id}`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestOperation_ParseResponseWithDefaultV3(t *testing.T) {
	t.Parallel()

	comment := `@Success default {object} nil "An empty response"`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	assert.Equal(t, "An empty response", operation.Responses.Spec.Default.Spec.Spec.Description)

	comment = `@Success 200,default {string} Response "A response"`
	operation = NewOperationV3(nil)

	err = operation.ParseComment(comment, nil)
	require.NoError(t, err)

	assert.Equal(t, "A response", operation.Responses.Spec.Default.Spec.Spec.Description)
	assert.Equal(t, "A response", operation.Responses.Spec.Response["200"].Spec.Spec.Description)
}

func TestParseResponseSuccessCommentWithEmptyResponseV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} nil "An empty response"`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `An empty response`, response.Spec.Spec.Description)
}

func TestParseResponseFailureCommentWithEmptyResponseV3(t *testing.T) {
	t.Parallel()

	comment := `@Failure 500 {object} nil`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	assert.Equal(t, "Internal Server Error", operation.Responses.Spec.Response["500"].Spec.Spec.Description)
}

func TestParseResponseCommentWithObjectTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.OrderRow "Error message, if code != 200`
	parser := New()
	operation := NewOperationV3(parser)
	operation.parser.addTestType("model.OrderRow")

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)

	assert.Equal(t, "#/components/model.OrderRow", response.Spec.Spec.Content["application/json"].Spec.Schema.Ref.Ref)
}

func TestParseResponseCommentWithNestedPrimitiveTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data=string,data2=int} "Error message, if code != 200`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)
	require.NotNil(t, response.Spec.Spec.Content["application/json"].Spec.Schema)

	allOf := operation.Responses.Spec.Default.Spec.Spec.Content["application/json"].Spec.Schema.Spec.AllOf
	require.NotNil(t, allOf)
	assert.Equal(t, 2, len(allOf))
	assert.Equal(t, "#/components/data", allOf[0].Ref.Ref)
	assert.Equal(t, "#/components/data2", allOf[1].Ref.Ref)
}

func TestParseResponseCommentWithNestedPrimitiveArrayTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data=[]string,data2=[]int} "Error message, if code != 200`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)
	assert.NotNil(t, operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"])
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"string"}), operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Spec.Items.Schema.Spec.Type)
}

func TestParseResponseCommentWithNestedObjectTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data=model.Payload,data2=model.Payload2} "Error message, if code != 200`
	operation := NewOperationV3(New())
	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")
	operation.parser.addTestType("model.Payload2")

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)
	assert.Equal(t, 2, len(response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.AllOf))
	assert.Equal(t, 5, len(operation.parser.openAPI.Components.Spec.Schemas))

	assert.Equal(t, "#/components/model.Payload", operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Ref.Ref)
	assert.Equal(t, "#/components/model.Payload2", operation.parser.openAPI.Components.Spec.Schemas["data2"].Spec.Properties["data2"].Ref.Ref)
}

func TestParseResponseCommentWithNestedArrayObjectTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data=[]model.Payload,data2=[]model.Payload2} "Error message, if code != 200`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")
	operation.parser.addTestType("model.Payload2")

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)

	allOf := response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.AllOf
	assert.Equal(t, 2, len(allOf))

	assert.Equal(t, "#/components/model.Payload", operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Spec.Items.Schema.Ref.Ref)
	assert.Equal(t, typeArray, operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Spec.Type)

	assert.Equal(t, "#/components/model.Payload2", operation.parser.openAPI.Components.Spec.Schemas["data2"].Spec.Properties["data2"].Spec.Items.Schema.Ref.Ref)
	assert.Equal(t, typeArray, operation.parser.openAPI.Components.Spec.Schemas["data2"].Spec.Properties["data2"].Spec.Type)
}

func TestParseResponseCommentWithNestedFieldsV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data1=int,data2=[]int,data3=model.Payload,data4=[]model.Payload} "Error message, if code != 200`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)

	allOf := response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.AllOf
	assert.Equal(t, 4, len(allOf))

	schemas := operation.parser.openAPI.Components.Spec.Schemas

	assert.Equal(t, typeInteger, schemas["data1"].Spec.Properties["data1"].Spec.Type)
	assert.Equal(t, typeObject, schemas["data1"].Spec.Type)

	assert.Equal(t, typeArray, schemas["data2"].Spec.Properties["data2"].Spec.Type)
	assert.Equal(t, typeInteger, schemas["data2"].Spec.Properties["data2"].Spec.Items.Schema.Spec.Type)
	assert.Equal(t, typeObject, schemas["data2"].Spec.Type)

	assert.Equal(t, "#/components/model.Payload", schemas["data3"].Spec.Properties["data3"].Ref.Ref)
	assert.Equal(t, typeObject, schemas["data3"].Spec.Type)

	assert.Equal(t, "#/components/model.Payload", schemas["data4"].Spec.Properties["data4"].Spec.Items.Schema.Ref.Ref)
	assert.Equal(t, typeArray, schemas["data4"].Spec.Properties["data4"].Spec.Type)
	assert.Equal(t, typeObject, schemas["data4"].Spec.Type)
}

func TestParseResponseCommentWithDeepNestedFieldsV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data1=int,data2=[]int,data3=model.Payload{data1=int,data2=model.DeepPayload},data4=[]model.Payload{data1=[]int,data2=[]model.DeepPayload}} "Error message, if code != 200`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")
	operation.parser.addTestType("model.DeepPayload")

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)

	allOf := response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.AllOf
	assert.Equal(t, 4, len(allOf))

	schemas := operation.parser.openAPI.Components.Spec.Schemas

	assert.Equal(t, typeInteger, schemas["data1"].Spec.Properties["data1"].Spec.Type)
	assert.Equal(t, typeObject, schemas["data1"].Spec.Type)

	assert.Equal(t, typeArray, schemas["data2"].Spec.Properties["data2"].Spec.Type)
	assert.Equal(t, typeInteger, schemas["data2"].Spec.Properties["data2"].Spec.Items.Schema.Spec.Type)
	assert.Equal(t, typeObject, schemas["data2"].Spec.Type)

	assert.Equal(t, typeObject, schemas["data3"].Spec.Type)
	assert.Equal(t, typeObject, schemas["data3"].Spec.Properties["data3"].Spec.Type)
	assert.Equal(t, 2, len(schemas["data3"].Spec.Properties["data3"].Spec.AllOf))

	assert.Equal(t, typeObject, schemas["data4"].Spec.Type)
	assert.Equal(t, typeArray, schemas["data4"].Spec.Properties["data4"].Spec.Type)
	assert.Equal(t, typeObject, schemas["data4"].Spec.Properties["data4"].Spec.Items.Schema.Spec.Type)
	assert.Equal(t, 2, len(schemas["data4"].Spec.Properties["data4"].Spec.Items.Schema.Spec.AllOf))
}

func TestParseResponseCommentWithNestedArrayMapFieldsV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} []map[string]model.CommonHeader{data1=[]map[string]model.Payload,data2=map[string][]int} "Error message, if code != 200`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")

	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)

	content := response.Spec.Spec.Content["application/json"]
	assert.NotNil(t, content)
	assert.NotNil(t, content.Spec)
	assert.NotNil(t, content.Spec.Schema.Spec.Items.Schema.Spec.AdditionalProperties.Schema)

	assert.Equal(t, 2, len(content.Spec.Schema.Spec.Items.Schema.Spec.AdditionalProperties.Schema.Spec.AllOf))
	assert.Equal(t, typeArray, content.Spec.Schema.Spec.Type)
	assert.Equal(t, typeObject, content.Spec.Schema.Spec.Items.Schema.Spec.Type)
	assert.Equal(t, typeObject, content.Spec.Schema.Spec.Items.Schema.Spec.AdditionalProperties.Schema.Spec.Type)

	schemas := operation.parser.openAPI.Components.Spec.Schemas

	data1 := schemas["data1"]
	assert.NotNil(t, data1)
	assert.NotNil(t, data1.Spec)
	assert.NotNil(t, data1.Spec.Properties)

	assert.Equal(t, typeObject, data1.Spec.Type)
	assert.Equal(t, typeArray, data1.Spec.Properties["data1"].Spec.Type)
	assert.Equal(t, typeObject, data1.Spec.Properties["data1"].Spec.Items.Schema.Spec.Type)
	assert.Equal(t, "#/components/model.Payload", data1.Spec.Properties["data1"].Spec.Items.Schema.Spec.AdditionalProperties.Schema.Ref.Ref)

	data2 := schemas["data2"]
	assert.NotNil(t, data2)
	assert.NotNil(t, data2.Spec)
	assert.NotNil(t, data2.Spec.Properties)

	assert.Equal(t, typeObject, data2.Spec.Type)
	assert.Equal(t, typeObject, data2.Spec.Properties["data2"].Spec.Type)
	assert.Equal(t, typeArray, data2.Spec.Properties["data2"].Spec.AdditionalProperties.Schema.Spec.Type)
	assert.Equal(t, typeInteger, data2.Spec.Properties["data2"].Spec.AdditionalProperties.Schema.Spec.Items.Schema.Spec.Type)

	commonHeader := schemas["model.CommonHeader"]
	assert.NotNil(t, commonHeader)
	assert.NotNil(t, commonHeader.Spec)
	assert.Equal(t, 2, len(commonHeader.Spec.AllOf))
	assert.Equal(t, typeObject, commonHeader.Spec.Type)

	payload := schemas["model.Payload"]
	assert.NotNil(t, payload)
	assert.NotNil(t, payload.Spec)
	assert.Equal(t, typeObject, payload.Spec.Type)
}

func TestParseResponseCommentWithObjectTypeInSameFileV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} testOwner "Error message, if code != 200"`
	operation := NewOperationV3(New())

	operation.parser.addTestType("swag.testOwner")

	fset := token.NewFileSet()
	astFile, err := goparser.ParseFile(fset, "operation_test.go", `package swag
	type testOwner struct {

	}
	`, goparser.ParseComments)
	assert.NoError(t, err)

	err = operation.ParseComment(comment, astFile)
	assert.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)
	assert.Equal(t, "#/components/swag.testOwner", response.Spec.Spec.Content["application/json"].Spec.Schema.Ref.Ref)
}

func TestParseResponseCommentWithObjectTypeErrV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.OrderRow "Error message, if code != 200"`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.notexist")

	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseResponseCommentWithArrayTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {array} model.OrderRow "Error message, if code != 200`
	operation := NewOperationV3(New())
	operation.parser.addTestType("model.OrderRow")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)
	assert.Equal(t, typeArray, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)
	assert.Equal(t, "#/components/model.OrderRow", response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Items.Schema.Ref.Ref)

}

func TestParseResponseCommentWithBasicTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {string} string "it's ok'"`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	require.NoError(t, err, "ParseComment should not fail")

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok'", response.Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)
}

func TestParseResponseCommentWithBasicTypeAndCodesV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200,201,default {string} string "it's ok"`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Response["201"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Default
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)
}

func TestParseEmptyResponseCommentV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 "it is ok"`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it is ok", response.Spec.Spec.Description)
}

func TestParseEmptyResponseCommentWithCodesV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200,201,default "it is ok"`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it is ok", response.Spec.Spec.Description)

	response = operation.Responses.Spec.Response["201"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it is ok", response.Spec.Spec.Description)

	response = operation.Responses.Spec.Default
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it is ok", response.Spec.Spec.Description)
}

func TestParseResponseCommentWithHeaderV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(nil)
	err := operation.ParseComment(`@Success 200 "it's ok"`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	err = operation.ParseComment(`@Header 200 {string} Token "qwerty"`, nil)
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

	err = operation.ParseComment(`@Header 200 "Mallformed"`, nil)
	assert.Error(t, err, "ParseComment should not fail")

	err = operation.ParseComment(`@Header 200,asdsd {string} Token "qwerty"`, nil)
	assert.Error(t, err, "ParseComment should not fail")
}
