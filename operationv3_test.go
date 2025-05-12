package swag

import (
	"go/ast"
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
var typeFile = spec.SingleOrArray[string](spec.SingleOrArray[string]{"file"})
var typeNumber = spec.SingleOrArray[string](spec.SingleOrArray[string]{NUMBER})
var typeBool = spec.SingleOrArray[string](spec.SingleOrArray[string]{BOOLEAN})

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

	assert.Equal(t, "#/components/schemas/model.OrderRow", response.Spec.Spec.Content["application/json"].Spec.Schema.Ref.Ref)
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

	allOf := operation.Responses.Spec.Response["200"].Spec.Spec.Content["application/json"].Spec.Schema.Spec.AllOf
	require.NotNil(t, allOf)
	assert.Equal(t, 2, len(allOf))
	found := map[string]struct{}{}
	for _, schema := range allOf {
		assert.NotNil(t, schema.Ref.Ref)
		found[schema.Ref.Ref] = struct{}{}
	}
	assert.NotNil(t, found["#/components/schemas/data"])
	assert.NotNil(t, found["#/components/schemas/data2"])
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
	assert.Equal(t, &typeString, operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Spec.Items.Schema.Spec.Type)
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

	assert.Equal(t, "#/components/schemas/model.Payload", operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Ref.Ref)
	assert.Equal(t, "#/components/schemas/model.Payload2", operation.parser.openAPI.Components.Spec.Schemas["data2"].Spec.Properties["data2"].Ref.Ref)
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

	assert.Equal(t, "#/components/schemas/model.Payload", operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Spec.Items.Schema.Ref.Ref)
	assert.Equal(t, &typeArray, operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Spec.Type)

	assert.Equal(t, "#/components/schemas/model.Payload2", operation.parser.openAPI.Components.Spec.Schemas["data2"].Spec.Properties["data2"].Spec.Items.Schema.Ref.Ref)
	assert.Equal(t, &typeArray, operation.parser.openAPI.Components.Spec.Schemas["data2"].Spec.Properties["data2"].Spec.Type)
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

	assert.Equal(t, &typeInteger, schemas["data1"].Spec.Properties["data1"].Spec.Type)
	assert.Equal(t, &typeObject, schemas["data1"].Spec.Type)

	assert.Equal(t, &typeArray, schemas["data2"].Spec.Properties["data2"].Spec.Type)
	assert.Equal(t, &typeInteger, schemas["data2"].Spec.Properties["data2"].Spec.Items.Schema.Spec.Type)
	assert.Equal(t, &typeObject, schemas["data2"].Spec.Type)

	assert.Equal(t, "#/components/schemas/model.Payload", schemas["data3"].Spec.Properties["data3"].Ref.Ref)
	assert.Equal(t, &typeObject, schemas["data3"].Spec.Type)

	assert.Equal(t, "#/components/schemas/model.Payload", schemas["data4"].Spec.Properties["data4"].Spec.Items.Schema.Ref.Ref)
	assert.Equal(t, &typeArray, schemas["data4"].Spec.Properties["data4"].Spec.Type)
	assert.Equal(t, &typeObject, schemas["data4"].Spec.Type)
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

	assert.Equal(t, &typeInteger, schemas["data1"].Spec.Properties["data1"].Spec.Type)
	assert.Equal(t, &typeObject, schemas["data1"].Spec.Type)

	assert.Equal(t, &typeArray, schemas["data2"].Spec.Properties["data2"].Spec.Type)
	assert.Equal(t, &typeInteger, schemas["data2"].Spec.Properties["data2"].Spec.Items.Schema.Spec.Type)
	assert.Equal(t, &typeObject, schemas["data2"].Spec.Type)

	assert.Equal(t, &typeObject, schemas["data3"].Spec.Type)
	assert.Equal(t, &typeObject, schemas["data3"].Spec.Properties["data3"].Spec.Type)
	assert.Equal(t, 2, len(schemas["data3"].Spec.Properties["data3"].Spec.AllOf))

	assert.Equal(t, &typeObject, schemas["data4"].Spec.Type)
	assert.Equal(t, &typeArray, schemas["data4"].Spec.Properties["data4"].Spec.Type)
	assert.Equal(t, &typeObject, schemas["data4"].Spec.Properties["data4"].Spec.Items.Schema.Spec.Type)
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
	assert.Equal(t, &typeArray, content.Spec.Schema.Spec.Type)
	assert.Equal(t, &typeObject, content.Spec.Schema.Spec.Items.Schema.Spec.Type)
	assert.Equal(t, &typeObject, content.Spec.Schema.Spec.Items.Schema.Spec.AdditionalProperties.Schema.Spec.Type)

	schemas := operation.parser.openAPI.Components.Spec.Schemas

	data1 := schemas["data1"]
	assert.NotNil(t, data1)
	assert.NotNil(t, data1.Spec)
	assert.NotNil(t, data1.Spec.Properties)

	assert.Equal(t, &typeObject, data1.Spec.Type)
	assert.Equal(t, &typeArray, data1.Spec.Properties["data1"].Spec.Type)
	assert.Equal(t, &typeObject, data1.Spec.Properties["data1"].Spec.Items.Schema.Spec.Type)
	assert.Equal(t, "#/components/schemas/model.Payload", data1.Spec.Properties["data1"].Spec.Items.Schema.Spec.AdditionalProperties.Schema.Ref.Ref)

	data2 := schemas["data2"]
	assert.NotNil(t, data2)
	assert.NotNil(t, data2.Spec)
	assert.NotNil(t, data2.Spec.Properties)

	assert.Equal(t, &typeObject, data2.Spec.Type)
	assert.Equal(t, &typeObject, data2.Spec.Properties["data2"].Spec.Type)
	assert.Equal(t, &typeArray, data2.Spec.Properties["data2"].Spec.AdditionalProperties.Schema.Spec.Type)
	assert.Equal(t, &typeInteger, data2.Spec.Properties["data2"].Spec.AdditionalProperties.Schema.Spec.Items.Schema.Spec.Type)

	commonHeader := schemas["model.CommonHeader"]
	assert.NotNil(t, commonHeader)
	assert.NotNil(t, commonHeader.Spec)
	assert.Equal(t, 2, len(commonHeader.Spec.AllOf))
	assert.Equal(t, &typeObject, commonHeader.Spec.Type)

	payload := schemas["model.Payload"]
	assert.NotNil(t, payload)
	assert.NotNil(t, payload.Spec)
	assert.Equal(t, &typeObject, payload.Spec.Type)
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
	assert.Equal(t, "#/components/schemas/swag.testOwner", response.Spec.Spec.Content["application/json"].Spec.Schema.Ref.Ref)
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
	assert.Equal(t, &typeArray, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)
	assert.Equal(t, "#/components/schemas/model.OrderRow", response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Items.Schema.Ref.Ref)

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
	assert.Equal(t, &typeString, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)
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
	assert.Equal(t, &typeString, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Response["201"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Default
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)
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

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	err = operation.ParseComment(`@Header 200 "Mallformed"`, nil)
	assert.Error(t, err, "ParseComment should fail")

	err = operation.ParseComment(`@Header 200,asdsd {string} Token "qwerty"`, nil)
	assert.Error(t, err, "ParseComment should fail")
}

func TestParseResponseCommentWithHeaderForCodesV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(nil)

	comment := `@Success 200,201,default "it's ok"`
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	comment = `@Header 200,201,default {string} Token "qwerty"`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	comment = `@Header all {string} Token2 "qwerty"`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token2"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token2"].Spec.Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Response["201"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token2"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token2"].Spec.Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Default
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token2"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token2"].Spec.Spec.Schema.Spec.Type)

	comment = `@Header 200 "Mallformed"`
	err = operation.ParseComment(comment, nil)
	assert.Error(t, err, "ParseComment should not fail")
}

func TestParseResponseCommentWithHeaderOnlyAllV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(nil)

	comment := `@Success 200,201,default "it's ok"`
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	comment = `@Header all {string} Token "qwerty"`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Response["201"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Default
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, &typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	comment = `@Header 200 "Mallformed"`
	err = operation.ParseComment(comment, nil)
	assert.Error(t, err, "ParseComment should not fail")
}

func TestParseEmptyResponseOnlyCodeV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(nil)
	err := operation.ParseComment(`@Success 200`, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "OK", response.Spec.Spec.Description)
}

func TestParseEmptyResponseOnlyCodesV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200,201,default`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err, "ParseComment should not fail")

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "OK", response.Spec.Spec.Description)

	response = operation.Responses.Spec.Response["201"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "Created", response.Spec.Spec.Description)

	response = operation.Responses.Spec.Default
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "", response.Spec.Spec.Description)
}

func TestParseResponseCommentParamMissingV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(nil)

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

func TestOperation_ParseParamCommentV3(t *testing.T) {
	t.Parallel()

	t.Run("integer", func(t *testing.T) {
		t.Parallel()
		for _, paramType := range []string{"header", "path", "query"} {
			t.Run(paramType, func(t *testing.T) {
				o := NewOperationV3(New())
				err := o.ParseComment(`@Param some_id `+paramType+` int true "Some ID"`, nil)

				assert.NoError(t, err)

				expected := &spec.RefOrSpec[spec.Extendable[spec.Parameter]]{
					Spec: &spec.Extendable[spec.Parameter]{
						Spec: &spec.Parameter{
							Name:        "some_id",
							Description: "Some ID",
							In:          paramType,
							Required:    true,
							Schema: &spec.RefOrSpec[spec.Schema]{
								Spec: &spec.Schema{
									JsonSchema: spec.JsonSchema{
										JsonSchemaCore: spec.JsonSchemaCore{
											Type: &typeInteger,
										},
									},
								},
							},
						},
					},
				}

				expectedArray := []*spec.RefOrSpec[spec.Extendable[spec.Parameter]]{expected}
				assert.Equal(t, o.Parameters, expectedArray)
			})
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		for _, paramType := range []string{"header", "path", "query"} {
			t.Run(paramType, func(t *testing.T) {
				o := NewOperationV3(New())
				err := o.ParseComment(`@Param some_string `+paramType+` string true "Some String"`, nil)

				assert.NoError(t, err)
				expected := &spec.RefOrSpec[spec.Extendable[spec.Parameter]]{
					Spec: &spec.Extendable[spec.Parameter]{
						Spec: &spec.Parameter{
							Description: "Some String",
							Name:        "some_string",
							In:          paramType,
							Required:    true,
							Schema: &spec.RefOrSpec[spec.Schema]{
								Spec: &spec.Schema{
									JsonSchema: spec.JsonSchema{
										JsonSchemaCore: spec.JsonSchemaCore{
											Type: &typeString,
										},
									},
								},
							},
						},
					},
				}

				expectedArray := []*spec.RefOrSpec[spec.Extendable[spec.Parameter]]{expected}
				assert.Equal(t, o.Parameters, expectedArray)
			})
		}
	})

	t.Run("object", func(t *testing.T) {
		t.Parallel()
		for _, paramType := range []string{"header", "path", "query"} {
			t.Run(paramType, func(t *testing.T) {
				assert.Error(t,
					NewOperationV3(New()).
						ParseComment(`@Param some_object `+paramType+` main.Object true "Some Object"`,
							nil))
			})
		}
	})

	t.Run("struct queries", func(t *testing.T) {
		t.Parallel()
		parser := New()
		parser.packages.uniqueDefinitions["main.Object"] = &TypeSpecDef{
			File: &ast.File{Name: &ast.Ident{Name: "test"}},
			TypeSpec: &ast.TypeSpec{
				Name:       &ast.Ident{Name: "Field"},
				TypeParams: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "T"}}}}},
				Type: &ast.StructType{
					Struct: 100,
					Fields: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{Name: "T"},
								},
								Type: ast.NewIdent("string"),
							},
							{
								Names: []*ast.Ident{
									{Name: "T2"},
								},
								Type: ast.NewIdent("string"),
							},
						},
					},
				},
			},
		}
		o := NewOperationV3(parser)
		err := o.ParseComment(`@Param some_object query main.Object true "Some Object"`,
			nil)

		assert.NoError(t, err)

		expectedT := &spec.RefOrSpec[spec.Extendable[spec.Parameter]]{
			Spec: &spec.Extendable[spec.Parameter]{
				Spec: &spec.Parameter{
					Name: "t",
					In:   "query",
					Schema: &spec.RefOrSpec[spec.Schema]{
						Spec: &spec.Schema{
							JsonSchema: spec.JsonSchema{
								JsonSchemaCore: spec.JsonSchemaCore{
									Type: &typeString,
								},
							},
						},
					},
				},
			},
		}
		expectedT2 := &spec.RefOrSpec[spec.Extendable[spec.Parameter]]{
			Spec: &spec.Extendable[spec.Parameter]{
				Spec: &spec.Parameter{
					Name: "t2",
					In:   "query",
					Schema: &spec.RefOrSpec[spec.Schema]{
						Spec: &spec.Schema{
							JsonSchema: spec.JsonSchema{
								JsonSchemaCore: spec.JsonSchemaCore{
									Type: &typeString,
								},
							},
						},
					},
				},
			},
		}

		assert.Len(t, o.Parameters, 2)
		tFound := false
		t2Found := false
		for _, param := range o.Parameters {
			switch param.Spec.Spec.Name {
			case "t":
				assert.EqualValues(t, expectedT, param)
				tFound = true
			case "t2":
				assert.EqualValues(t, expectedT2, param)
				t2Found = true
			default:
				assert.Fail(t, "unexpected result")
			}
		}

		assert.True(t, tFound, "results should contain t")
		assert.True(t, t2Found, "results should contain t2")
	})
}

// Test ParseParamComment Query Params
func TestParseParamCommentBodyArrayV3(t *testing.T) {
	t.Parallel()

	comment := `@Param names body []string true "Users List"`
	o := NewOperationV3(New())
	err := o.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.NotNil(t, o.RequestBody)
	assert.Equal(t, "Users List", o.RequestBody.Spec.Spec.Description)
	assert.True(t, o.RequestBody.Spec.Spec.Required)
	assert.Equal(t, &typeArray, o.RequestBody.Spec.Spec.Content["application/json"].Spec.Schema.Spec.Type)
}

func TestParseParamCommentArrayV3(t *testing.T) {
	paramTypes := []string{"header", "path", "query"}

	for _, paramType := range paramTypes {
		t.Run(paramType, func(t *testing.T) {
			operation := NewOperationV3(New())
			err := operation.ParseComment(`@Param names `+paramType+` []string true "Users List"`, nil)
			assert.NoError(t, err)

			parameters := operation.Operation.Parameters
			assert.NotNil(t, parameters)

			parameterSpec := parameters[0].Spec.Spec
			assert.NotNil(t, parameterSpec)
			assert.Equal(t, "Users List", parameterSpec.Description)
			assert.Equal(t, "names", parameterSpec.Name)
			assert.Equal(t, &typeArray, parameterSpec.Schema.Spec.Type)
			assert.Equal(t, true, parameterSpec.Required)
			assert.Equal(t, paramType, parameterSpec.In)
			assert.Equal(t, &typeString, parameterSpec.Schema.Spec.Items.Schema.Spec.Type)

			err = operation.ParseComment(`@Param names `+paramType+` []model.User true "Users List"`, nil)
			assert.Error(t, err)
		})
	}
}

func TestParseParamCommentDefaultValueV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(New())
	err := operation.ParseComment(`@Param names query string true "Users List" default(test)`, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Users List", parameterSpec.Description)
	assert.Equal(t, "names", parameterSpec.Name)
	assert.Equal(t, &typeString, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, true, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, "test", parameterSpec.Schema.Spec.Default)
}

func TestParseParamCommentQueryArrayFormatV3(t *testing.T) {
	t.Parallel()

	comment := `@Param names query []string true "Users List" collectionFormat(multi)`
	operation := NewOperationV3(New())
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Users List", parameterSpec.Description)
	assert.Equal(t, "names", parameterSpec.Name)
	assert.Equal(t, &typeArray, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, true, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, &typeString, parameterSpec.Schema.Spec.Items.Schema.Spec.Type)
	assert.Equal(t, "form", parameterSpec.Style)

}

func TestParseParamCommentByIDV3(t *testing.T) {
	t.Parallel()

	comment := `@Param unsafe_id[lte] query int true "Unsafe query param"`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Unsafe query param", parameterSpec.Description)
	assert.Equal(t, "unsafe_id[lte]", parameterSpec.Name)
	assert.Equal(t, &typeInteger, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, true, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
}

func TestParseParamCommentByQueryTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id query int true "Some ID"`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.Equal(t, &typeInteger, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, true, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
}

func TestParseParamCommentByBodyTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id body model.OrderRow true "Some ID"`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.OrderRow")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	requestBody := operation.RequestBody
	assert.NotNil(t, requestBody)

	requestBodySpec := requestBody.Spec.Spec
	assert.NotNil(t, requestBodySpec)
	assert.Equal(t, "Some ID", requestBodySpec.Description)
	assert.Equal(t, true, requestBodySpec.Required)
	assert.Equal(t, "#/components/schemas/model.OrderRow", requestBodySpec.Content["application/json"].Spec.Schema.Ref.Ref)
}

func TestParseParamCommentByBodyTextPlainV3(t *testing.T) {
	t.Parallel()

	comment := `@Param text body string true "Text to process"`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	requestBody := operation.RequestBody
	assert.NotNil(t, requestBody)

	requestBodySpec := requestBody.Spec.Spec
	assert.NotNil(t, requestBodySpec)
	assert.Equal(t, "Text to process", requestBodySpec.Description)
	assert.Equal(t, true, requestBodySpec.Required)
	assert.Equal(t, &typeString, requestBodySpec.Content["text/plain"].Spec.Schema.Spec.Type)
}

func TestParseParamCommentByBodyTypeWithDeepNestedFieldsV3(t *testing.T) {
	t.Parallel()

	comment := `@Param body body model.CommonHeader{data=string,data2=int} true "test deep"`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Len(t, operation.Parameters, 0)

	requestBody := operation.RequestBody
	assert.NotNil(t, requestBody)

	requestBodySpec := requestBody.Spec.Spec
	assert.NotNil(t, requestBodySpec)
	assert.Equal(t, "test deep", requestBodySpec.Description)
	assert.True(t, requestBodySpec.Required)

	assert.Equal(t, 2, len(requestBodySpec.Content["application/json"].Spec.Schema.Spec.AllOf))
	assert.Equal(t, 3, len(operation.parser.openAPI.Components.Spec.Schemas))
}

func TestParseParamCommentByBodyTypeArrayOfPrimitiveGoV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id body []int true "Some ID"`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	requestBody := operation.RequestBody
	assert.NotNil(t, requestBody)

	requestBodySpec := requestBody.Spec.Spec
	assert.NotNil(t, requestBodySpec)
	assert.Equal(t, "Some ID", requestBodySpec.Description)
	assert.True(t, requestBodySpec.Required)
	assert.Equal(t, &typeArray, requestBodySpec.Content["application/json"].Spec.Schema.Spec.Type)
	assert.Equal(t, &typeInteger, requestBodySpec.Content["application/json"].Spec.Schema.Spec.Items.Schema.Spec.Type)
}

func TestParseParamCommentByBodyTypeArrayOfPrimitiveGoWithDeepNestedFieldsV3(t *testing.T) {
	t.Parallel()

	comment := `@Param body body []model.CommonHeader{data=string,data2=int} true "test deep"`
	operation := NewOperationV3(New())
	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Len(t, operation.Parameters, 0)

	assert.NotNil(t, operation.RequestBody)

	parameterSpec := operation.RequestBody.Spec.Spec.Content["application/json"].Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "test deep", operation.RequestBody.Spec.Spec.Description)
	assert.Equal(t, &typeArray, parameterSpec.Schema.Spec.Type)
	assert.True(t, operation.RequestBody.Spec.Spec.Required)
	assert.Equal(t, 2, len(parameterSpec.Schema.Spec.Items.Schema.Spec.AllOf))
}

func TestParseParamCommentByBodyTypeErrV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id body model.OrderRow true "Some ID"`
	operation := NewOperationV3(New())
	operation.parser.addTestType("model.notexist")

	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseParamCommentByFormDataTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Param file formData file true "this is a test file"`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Len(t, operation.Parameters, 0)
	assert.NotNil(t, operation.RequestBody)

	requestBody := operation.RequestBody
	assert.True(t, requestBody.Spec.Spec.Required)
	assert.Equal(t, "this is a test file", requestBody.Spec.Spec.Description)
	assert.NotNil(t, requestBody)

	requestBodySpec := requestBody.Spec.Spec
	assert.NotNil(t, requestBodySpec)
	assert.Equal(t, &typeFile, requestBodySpec.Content["application/x-www-form-urlencoded"].Spec.Schema.Spec.Type)
}

func TestParseParamCommentByFormDataTypeUint64V3(t *testing.T) {
	t.Parallel()

	comment := `@Param file formData uint64 true "this is a test file"`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Len(t, operation.Parameters, 0)

	requestBody := operation.RequestBody
	assert.NotNil(t, requestBody)
	assert.Equal(t, "this is a test file", requestBody.Spec.Spec.Description)

	requestBodySpec := requestBody.Spec.Spec.Content["application/x-www-form-urlencoded"].Spec
	assert.NotNil(t, requestBodySpec)
	assert.Equal(t, &typeInteger, requestBodySpec.Schema.Spec.Type)
}

func TestParseParamCommentByNotSupportedTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id not_supported int true "Some ID"`
	operation := NewOperationV3(New())
	err := operation.ParseComment(comment, nil)

	assert.Error(t, err)
}

func TestParseParamCommentNotMatchV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id body mock true`
	operation := NewOperationV3(New())
	err := operation.ParseComment(comment, nil)

	assert.Error(t, err)
}

func TestParseParamCommentByEnumsV3(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		comment := `@Param some_id query string true "Some ID" Enums(A, B, C)`
		operation := NewOperationV3(New())

		err := operation.ParseComment(comment, nil)
		assert.NoError(t, err)

		assert.Len(t, operation.Parameters, 1)

		parameters := operation.Operation.Parameters
		assert.NotNil(t, parameters)

		parameterSpec := parameters[0].Spec.Spec
		assert.NotNil(t, parameterSpec)
		assert.Equal(t, "Some ID", parameterSpec.Description)
		assert.Equal(t, "some_id", parameterSpec.Name)
		assert.True(t, parameterSpec.Required)
		assert.Equal(t, "query", parameterSpec.In)
		assert.Equal(t, &typeString, parameterSpec.Schema.Spec.Type)
		assert.Equal(t, 3, len(parameterSpec.Schema.Spec.Enum))

		enums := []interface{}{"A", "B", "C"}
		assert.EqualValues(t, enums, parameterSpec.Schema.Spec.Enum)
	})

	t.Run("int", func(t *testing.T) {
		comment := `@Param some_id query int true "Some ID" Enums(1, 2, 3)`
		operation := NewOperationV3(New())

		err := operation.ParseComment(comment, nil)
		assert.NoError(t, err)

		parameters := operation.Operation.Parameters
		assert.NotNil(t, parameters)

		parameterSpec := parameters[0].Spec.Spec
		assert.NotNil(t, parameterSpec)
		assert.Equal(t, "Some ID", parameterSpec.Description)
		assert.Equal(t, "some_id", parameterSpec.Name)
		assert.True(t, parameterSpec.Required)
		assert.Equal(t, "query", parameterSpec.In)
		assert.Equal(t, &typeInteger, parameterSpec.Schema.Spec.Type)
		assert.Equal(t, 3, len(parameterSpec.Schema.Spec.Enum))

		enums := []interface{}{1, 2, 3}
		assert.EqualValues(t, enums, parameterSpec.Schema.Spec.Enum)
	})

	t.Run("number", func(t *testing.T) {
		comment := `@Param some_id query number true "Some ID" Enums(1.1, 2.2, 3.3)`
		operation := NewOperationV3(New())

		err := operation.ParseComment(comment, nil)
		assert.NoError(t, err)

		parameters := operation.Operation.Parameters
		assert.NotNil(t, parameters)

		parameterSpec := parameters[0].Spec.Spec
		assert.NotNil(t, parameterSpec)
		assert.Equal(t, "Some ID", parameterSpec.Description)
		assert.Equal(t, "some_id", parameterSpec.Name)
		assert.True(t, parameterSpec.Required)
		assert.Equal(t, "query", parameterSpec.In)
		assert.Equal(t, &typeNumber, parameterSpec.Schema.Spec.Type)
		assert.Equal(t, 3, len(parameterSpec.Schema.Spec.Enum))

		enums := []interface{}{1.1, 2.2, 3.3}
		assert.EqualValues(t, enums, parameterSpec.Schema.Spec.Enum)
	})

	t.Run("bool", func(t *testing.T) {
		comment := `@Param some_id query bool true "Some ID" Enums(true, false)`
		operation := NewOperationV3(New())

		err := operation.ParseComment(comment, nil)
		assert.NoError(t, err)

		parameters := operation.Operation.Parameters
		assert.NotNil(t, parameters)

		parameterSpec := parameters[0].Spec.Spec
		assert.NotNil(t, parameterSpec)
		assert.Equal(t, "Some ID", parameterSpec.Description)
		assert.Equal(t, "some_id", parameterSpec.Name)
		assert.True(t, parameterSpec.Required)
		assert.Equal(t, "query", parameterSpec.In)
		assert.Equal(t, &typeBool, parameterSpec.Schema.Spec.Type)
		assert.Equal(t, 2, len(parameterSpec.Schema.Spec.Enum))

		enums := []interface{}{true, false}
		assert.EqualValues(t, enums, parameterSpec.Schema.Spec.Enum)
	})

	operation := NewOperationV3(New())

	comment := `@Param some_id query int true "Some ID" Enums(A, B, C)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query number true "Some ID" Enums(A, B, C)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query boolean true "Some ID" Enums(A, B, C)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query Document true "Some ID" Enums(A, B, C)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByMaxLengthV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id query string true "Some ID" MaxLength(10)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, &typeString, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, 10, *parameterSpec.Schema.Spec.MaxLength)

	comment = `@Param some_id query int true "Some ID" MaxLength(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query string true "Some ID" MaxLength(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByMinLengthV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id query string true "Some ID" MinLength(10)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, &typeString, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, 10, *parameterSpec.Schema.Spec.MinLength)

	comment = `@Param some_id query int true "Some ID" MinLength(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query string true "Some ID" MinLength(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByMinimumV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id query int true "Some ID" Minimum(10)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, &typeInteger, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, 10, *parameterSpec.Schema.Spec.Minimum)

	comment = `@Param some_id query int true "Some ID" Mininum(10)`
	assert.NoError(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query string true "Some ID" Minimum(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query integer true "Some ID" Minimum(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByMaximumV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id query int true "Some ID" Maximum(10)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, &typeInteger, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, 10, *parameterSpec.Schema.Spec.Maximum)

	comment = `@Param some_id query int true "Some ID" Maxinum(10)`
	assert.NoError(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query string true "Some ID" Maximum(10)`
	assert.Error(t, operation.ParseComment(comment, nil))

	comment = `@Param some_id query integer true "Some ID" Maximum(Goopher)`
	assert.Error(t, operation.ParseComment(comment, nil))
}

func TestParseParamCommentByDefaultV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id query int true "Some ID" Default(10)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, &typeInteger, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, 10, parameterSpec.Schema.Spec.Default)
}

func TestParseParamCommentByExampleIntV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id query int true "Some ID" Example(10)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, &typeInteger, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, 10, parameterSpec.Example)
}

func TestParseParamCommentByExampleStringV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id query string true "Some ID" Example(True feelings)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, &typeString, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, "True feelings", parameterSpec.Example)
}

func TestParseParamCommentBySchemaExampleStringV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id body string true "Some ID" SchemaExample(True feelings)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	requestBody := operation.RequestBody
	assert.NotNil(t, requestBody)

	requestBodySpec := requestBody.Spec.Spec
	assert.NotNil(t, requestBodySpec)
	assert.Equal(t, "Some ID", requestBodySpec.Description)
	assert.True(t, requestBodySpec.Required)
	assert.Equal(t, "True feelings", requestBodySpec.Content["text/plain"].Spec.Schema.Spec.Example)
	assert.Equal(t, &typeString, requestBodySpec.Content["text/plain"].Spec.Schema.Spec.Type)
}

func TestParseParamCommentBySchemaExampleUnsupportedTypeV3(t *testing.T) {
	t.Parallel()
	var param spec.Parameter

	setSchemaExampleV3(nil, "something", "random value")
	assert.Nil(t, param.Schema)

	setSchemaExampleV3(nil, STRING, "string value")
	assert.Nil(t, param.Schema)

	param.Schema = spec.NewSchemaSpec()
	setSchemaExampleV3(param.Schema.Spec, STRING, "string value")
	assert.Equal(t, "string value", param.Schema.Spec.Example)

	setSchemaExampleV3(param.Schema.Spec, INTEGER, "10")
	assert.Equal(t, 10, param.Schema.Spec.Example)

	setSchemaExampleV3(param.Schema.Spec, NUMBER, "10")
	assert.Equal(t, float64(10), param.Schema.Spec.Example)

	setSchemaExampleV3(param.Schema.Spec, STRING, "string \\r\\nvalue")
	assert.Equal(t, "string \r\nvalue", param.Schema.Spec.Example)
}

func TestParseParamArrayWithEnumsV3(t *testing.T) {
	t.Parallel()

	comment := `@Param field query []string true "An enum collection" collectionFormat(csv) enums(also,valid)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "An enum collection", parameterSpec.Description)
	assert.Equal(t, "field", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, &typeArray, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, "form", parameterSpec.Style)

	enums := []interface{}{"also", "valid"}
	assert.EqualValues(t, enums, parameterSpec.Schema.Spec.Items.Schema.Spec.Enum)
	assert.Equal(t, &typeString, parameterSpec.Schema.Spec.Items.Schema.Spec.Type)
}

func TestParseAndExtractionParamAttributeV3(t *testing.T) {
	t.Parallel()

	op := NewOperationV3(New())

	t.Run("number", func(t *testing.T) {
		numberParam := spec.Parameter{
			Schema: spec.NewSchemaSpec(),
		}
		err := op.parseParamAttribute(
			" default(1) maximum(100) minimum(0) format(csv)",
			"",
			NUMBER,
			&numberParam,
		)
		assert.NoError(t, err)
		assert.Equal(t, int(0), *numberParam.Schema.Spec.Minimum)
		assert.Equal(t, int(100), *numberParam.Schema.Spec.Maximum)
		assert.Equal(t, "csv", numberParam.Schema.Spec.Format)
		assert.Equal(t, float64(1), numberParam.Schema.Spec.Default)

		err = op.parseParamAttribute(" minlength(1)", "", NUMBER, nil)
		assert.Error(t, err)

		err = op.parseParamAttribute(" maxlength(1)", "", NUMBER, nil)
		assert.Error(t, err)
	})

	t.Run("string", func(t *testing.T) {
		stringParam := spec.Parameter{
			Schema: spec.NewSchemaSpec(),
		}
		err := op.parseParamAttribute(
			" default(test) maxlength(100) minlength(0) format(csv)",
			"",
			STRING,
			&stringParam,
		)
		assert.NoError(t, err)
		assert.Equal(t, int(0), *stringParam.Schema.Spec.MinLength)
		assert.Equal(t, int(100), *stringParam.Schema.Spec.MaxLength)
		assert.Equal(t, "csv", stringParam.Schema.Spec.Format)
		err = op.parseParamAttribute(" minimum(0)", "", STRING, nil)
		assert.Error(t, err)

		err = op.parseParamAttribute(" maximum(0)", "", STRING, nil)
		assert.Error(t, err)
	})

	t.Run("array", func(t *testing.T) {
		arrayParam := spec.Parameter{
			Schema: spec.NewSchemaSpec(),
		}

		arrayParam.In = "path"
		err := op.parseParamAttribute(" collectionFormat(simple)", ARRAY, STRING, &arrayParam)
		assert.Equal(t, "simple", arrayParam.Style)
		assert.NoError(t, err)

		err = op.parseParamAttribute(" collectionFormat(simple)", STRING, STRING, nil)
		assert.Error(t, err)

		err = op.parseParamAttribute(" default(0)", "", ARRAY, nil)
		assert.Error(t, err)
	})
}

func TestParseParamCommentByExtensionsV3(t *testing.T) {
	comment := `@Param some_id path int true "Some ID" extensions(x-example=test,x-custom=Gopher,x-custom2)`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.Equal(t, "path", parameterSpec.In)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, &typeInteger, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, "Gopher", parameterSpec.Schema.Spec.Extensions["x-custom"])
	assert.Equal(t, true, parameterSpec.Schema.Spec.Extensions["x-custom2"])
	assert.Equal(t, "test", parameterSpec.Schema.Spec.Extensions["x-example"])
}

func TestParseIdCommentV3(t *testing.T) {
	t.Parallel()

	comment := `@Id myOperationId`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)

	assert.NoError(t, err)
	assert.Equal(t, "myOperationId", operation.Operation.OperationID)
}

func TestParseSecurityCommentV3(t *testing.T) {
	t.Parallel()

	comment := `@Security OAuth2Implicit[read, write]`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	expected := []spec.SecurityRequirement{{
		"OAuth2Implicit": {"read", "write"},
	}}

	assert.Equal(t, expected, operation.Security)
}

func TestParseSecurityCommentSimpleV3(t *testing.T) {
	t.Parallel()

	comment := `@Security ApiKeyAuth`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	expected := []spec.SecurityRequirement{{
		"ApiKeyAuth": {},
	}}

	assert.Equal(t, expected, operation.Security)
}

func TestParseSecurityCommentOrV3(t *testing.T) {
	t.Parallel()

	comment := `@Security OAuth2Implicit[read, write] || Firebase[]`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	expected := []spec.SecurityRequirement{{
		"OAuth2Implicit": {"read", "write"},
		"Firebase":       {""},
	}}

	assert.Equal(t, expected, operation.Security)
}

func TestParseMultiDescriptionV3(t *testing.T) {
	t.Parallel()

	comment := `@Description line one`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	comment = `@Tags multi`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	comment = `@Description line two x`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Equal(t, "line one\nline two x", operation.Description)
}

func TestParseDescriptionMarkdownV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(New())
	operation.parser.markdownFileDir = "example/markdown"

	comment := `@description.markdown admin.md`

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	comment = `@description.markdown missing.md`

	err = operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseSummaryV3(t *testing.T) {
	t.Parallel()

	comment := `@summary line one`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Equal(t, "line one", operation.Summary)

	comment = `@Summary line one`
	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err)
}

func TestParseDeprecationDescriptionV3(t *testing.T) {
	t.Parallel()

	comment := `@Deprecated`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.True(t, operation.Deprecated)
}

func TestParseExtensionsV3(t *testing.T) {
	t.Parallel()
	// Fail if there are no args for attributes.
	{
		comment := `@x-amazon-apigateway-integration`
		operation := NewOperationV3(New())

		err := operation.ParseComment(comment, nil)
		assert.EqualError(t, err, "annotation @x-amazon-apigateway-integration need a value")
	}

	// Fail if args of attributes are broken.
	{
		comment := `@x-amazon-apigateway-integration ["broken"}]`
		operation := NewOperationV3(New())

		err := operation.ParseComment(comment, nil)
		assert.EqualError(t, err, "annotation @x-amazon-apigateway-integration need a valid json value. error: invalid character '}' after array element")
	}

	// OK
	{
		comment := `@x-amazon-apigateway-integration {"uri": "${some_arn}", "passthroughBehavior": "when_no_match", "httpMethod": "POST", "type": "aws_proxy"}`
		operation := NewOperationV3(New())

		err := operation.ParseComment(comment, nil)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"httpMethod":          "POST",
			"passthroughBehavior": "when_no_match",
			"type":                "aws_proxy",
			"uri":                 "${some_arn}",
		}, operation.Responses.Extensions["x-amazon-apigateway-integration"])
	}

	// Test x-tagGroups
	{
		comment := `@x-tagGroups [{"name":"Natural Persons","tags":["Person","PersonRisk","PersonDocuments"]}]`
		operation := NewOperationV3(New())

		err := operation.ParseComment(comment, nil)
		assert.NoError(t, err)
		assert.Equal(t,
			[]interface{}{map[string]interface{}{
				"name": "Natural Persons",
				"tags": []interface{}{
					"Person",
					"PersonRisk",
					"PersonDocuments",
				},
			}}, operation.Responses.Extensions["x-tagGroups"])
	}
}

func TestParseResponseHeaderCommentV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(New())

	err := operation.ParseResponseComment(`default {string} string "other error"`, nil)
	assert.NoError(t, err)
	err = operation.ParseResponseHeaderComment(`all {string} Token "qwerty"`, nil)
	assert.NoError(t, err)
}

func TestParseCodeSamplesV3(t *testing.T) {
	t.Parallel()
	const comment = `@x-codeSamples file`
	t.Run("Find sample by file", func(t *testing.T) {

		operation := NewOperationV3(New(), SetCodeExampleFilesDirectoryV3("testdata/code_examples"))
		operation.Summary = "example"

		err := operation.ParseComment(comment, nil)
		require.NoError(t, err, "no error should be thrown")

		assert.Equal(t, "example", operation.Summary)
		assert.Equal(t, CodeSamples(CodeSamples{map[string]string{"lang": "JavaScript", "source": "console.log('Hello World');"}}),
			operation.Responses.Extensions["x-codeSamples"],
		)
	})

	t.Run("With broken file sample", func(t *testing.T) {
		operation := NewOperationV3(New(), SetCodeExampleFilesDirectoryV3("testdata/code_examples"))
		operation.Summary = "broken"

		err := operation.ParseComment(comment, nil)
		assert.Error(t, err, "no error should be thrown")
	})

	t.Run("Example file not found", func(t *testing.T) {
		operation := NewOperationV3(New(), SetCodeExampleFilesDirectoryV3("testdata/code_examples"))
		operation.Summary = "badExample"

		err := operation.ParseComment(comment, nil)
		assert.Error(t, err, "error was expected, as file does not exist")
	})

	t.Run("Without line reminder", func(t *testing.T) {
		comment := `@x-codeSamples`
		operation := NewOperationV3(New(), SetCodeExampleFilesDirectoryV3("testdata/code_examples"))
		operation.Summary = "example"

		err := operation.ParseComment(comment, nil)
		assert.Error(t, err, "no error should be thrown")
	})

	t.Run(" broken dir", func(t *testing.T) {
		operation := NewOperationV3(New(), SetCodeExampleFilesDirectoryV3("testdata/fake_examples"))
		operation.Summary = "code"

		err := operation.ParseComment(comment, nil)
		assert.Error(t, err, "no error should be thrown")
	})
}

func TestParseAcceptCommentV3(t *testing.T) {
	t.Parallel()

	comment := `//@Accept json,xml,plain,html,mpfd,x-www-form-urlencoded,json-api,json-stream,octet-stream,png,jpeg,gif,application/xhtml+xml,application/health+json`
	operation := NewOperationV3(New())
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	resultMapKeys := []string{
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
		"application/health+json"}

	content := operation.RequestBody.Spec.Spec.Content
	for _, key := range resultMapKeys {
		assert.NotNil(t, content[key])
	}

	assert.Equal(t, &typeObject, content["application/json"].Spec.Schema.Spec.Type)
	assert.Equal(t, &typeObject, content["text/xml"].Spec.Schema.Spec.Type)
	assert.Equal(t, &typeString, content["image/png"].Spec.Schema.Spec.Type)
	assert.Equal(t, "binary", content["image/png"].Spec.Schema.Spec.Format)
}

func TestParseAcceptCommentErrV3(t *testing.T) {
	t.Parallel()

	comment := `//@Accept unknown`
	operation := NewOperationV3(New())
	err := operation.ParseComment(comment, nil)
	assert.Error(t, err)
}

func TestParseProduceCommandV3(t *testing.T) {
	t.Parallel()

	t.Run("Produce success", func(t *testing.T) {
		t.Parallel()

		const comment = "//@Produce application/json,text/csv,application/zip"

		operation := NewOperationV3(New())
		err := operation.ParseComment(comment, nil)
		assert.NoError(t, err)

		assert.Equal(t, 3, len(operation.responseMimeTypes))
	})

	t.Run("Produce Invalid Mime Type", func(t *testing.T) {
		t.Parallel()

		const comment = "//@Produce text,stuff,gophers"

		operation := NewOperationV3(New())
		err := operation.ParseComment(comment, nil)
		assert.Error(t, err)
	})
}

func TestProcessProduceComment(t *testing.T) {
	t.Parallel()

	const comment = "//@Produce application/json,text/csv,application/zip"

	operation := NewOperationV3(New())
	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	operation.Responses.Spec.Response = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Response]])
	operation.Responses.Spec.Response["200"] = spec.NewResponseSpec()
	operation.Responses.Spec.Response["201"] = spec.NewResponseSpec()
	operation.Responses.Spec.Response["204"] = spec.NewResponseSpec()
	operation.Responses.Spec.Response["400"] = spec.NewResponseSpec()
	operation.Responses.Spec.Response["500"] = spec.NewResponseSpec()

	err = operation.ProcessProduceComment()
	require.NoError(t, err)

	content := operation.Responses.Spec.Response["200"].Spec.Spec.Content
	assert.Equal(t, 3, len(content))
	assert.NotNil(t, content["application/json"].Spec.Schema)
	assert.NotNil(t, content["text/csv"].Spec.Schema)
	assert.NotNil(t, content["application/zip"].Spec.Schema)

	content = operation.Responses.Spec.Response["201"].Spec.Spec.Content
	assert.Equal(t, 3, len(content))
	assert.NotNil(t, content["application/json"].Spec.Schema)
	assert.NotNil(t, content["text/csv"].Spec.Schema)
	assert.NotNil(t, content["application/zip"].Spec.Schema)

	content = operation.Responses.Spec.Response["204"].Spec.Spec.Content
	assert.Nil(t, content)

	content = operation.Responses.Spec.Response["400"].Spec.Spec.Content
	assert.Nil(t, content)

	content = operation.Responses.Spec.Response["500"].Spec.Spec.Content
	assert.Nil(t, content)
}

func TestParseServerCommentV3(t *testing.T) {
	t.Parallel()

	operation := NewOperationV3(nil)

	comment := `/@servers.url https://api.example.com/v1`
	err := operation.ParseComment(comment, nil)
	require.NoError(t, err)

	comment = `/@servers.description override path 1`
	err = operation.ParseComment(comment, nil)
	require.NoError(t, err)

	comment = `/@servers.url https://api.example.com/v2`
	err = operation.ParseComment(comment, nil)
	require.NoError(t, err)

	comment = `/@servers.description override path 2`
	err = operation.ParseComment(comment, nil)
	require.NoError(t, err)

	assert.Len(t, operation.Servers, 2)
	assert.Equal(t, "https://api.example.com/v1", operation.Servers[0].Spec.URL)
	assert.Equal(t, "override path 1", operation.Servers[0].Spec.Description)
	assert.Equal(t, "https://api.example.com/v2", operation.Servers[1].Spec.URL)
	assert.Equal(t, "override path 2", operation.Servers[1].Spec.Description)
}

func TestResponseSchemaWithCustomMimeTypeV3(t *testing.T) {
	t.Parallel()

	t.Run("Schema ref is correctly associated with custom MIME type", func(t *testing.T) {
		t.Parallel()

		// Create operation with parser to handle the type reference
		parser := New()
		operation := NewOperationV3(parser)

		// Create a mock type in the parser as a stand-in for model.OrderRow
		parser.addTestType("model.OrderRow")

		// First, set the response MIME type with @Produce
		err := operation.ParseComment("/@Produce json-api", nil)
		require.NoError(t, err)

		// Then set the response with @Success
		err = operation.ParseComment("/@Success 200 {object} model.OrderRow", nil)
		require.NoError(t, err)

		// Check that we have a response for status code 200
		response, exists := operation.Responses.Spec.Response["200"]
		require.True(t, exists, "Response for status code 200 should exist")

		// Verify the correct MIME type (json-api -> application/vnd.api+json) has the schema reference
		content := response.Spec.Spec.Content
		require.NotNil(t, content, "Response content should not be nil")

		// Check that application/vnd.api+json exists in the content map
		apiJsonContent, exists := content["application/vnd.api+json"]
		require.True(t, exists, "application/vnd.api+json content should exist")

		// Verify the schema reference is correct
		require.NotNil(t, apiJsonContent.Spec.Schema, "Schema should not be nil")
		require.NotNil(t, apiJsonContent.Spec.Schema.Ref, "Schema ref should not be nil")
		require.Equal(t, "#/components/schemas/model.OrderRow", apiJsonContent.Spec.Schema.Ref.Ref)

		// Make sure the schema is NOT also defined under application/json
		_, exists = content["application/json"]
		require.False(t, exists, "application/json content should not exist when only json-api was specified")
	})

	t.Run("Default to application/json when no MIME type is specified", func(t *testing.T) {
		t.Parallel()

		// Create operation with parser to handle the type reference
		parser := New()
		operation := NewOperationV3(parser)

		// Create a mock type in the parser
		parser.addTestType("model.OrderRow")

		// Only set the response with @Success, without any @Produce
		err := operation.ParseComment("/@Success 200 {object} model.OrderRow", nil)
		require.NoError(t, err)

		// Check that we have a response for status code 200
		response, exists := operation.Responses.Spec.Response["200"]
		require.True(t, exists, "Response for status code 200 should exist")

		// Verify application/json has the schema reference
		content := response.Spec.Spec.Content
		require.NotNil(t, content, "Response content should not be nil")

		// Check that application/json exists in the content map
		jsonContent, exists := content["application/json"]
		require.True(t, exists, "application/json content should exist")

		// Verify the schema reference is correct
		require.NotNil(t, jsonContent.Spec.Schema, "Schema should not be nil")
		require.NotNil(t, jsonContent.Spec.Schema.Ref, "Schema ref should not be nil")
		require.Equal(t, "#/components/schemas/model.OrderRow", jsonContent.Spec.Schema.Ref.Ref)
	})

	t.Run("Multiple MIME types have the same schema reference", func(t *testing.T) {
		t.Parallel()

		// Create operation with parser to handle the type reference
		parser := New()
		operation := NewOperationV3(parser)

		// Create a mock type in the parser
		parser.addTestType("model.OrderRow")

		// Set multiple MIME types
		err := operation.ParseComment("/@Produce json,json-api", nil)
		require.NoError(t, err)

		// Set the response
		err = operation.ParseComment("/@Success 200 {object} model.OrderRow", nil)
		require.NoError(t, err)

		// Check that we have a response for status code 200
		response, exists := operation.Responses.Spec.Response["200"]
		require.True(t, exists, "Response for status code 200 should exist")

		// Verify both MIME types have the schema reference
		content := response.Spec.Spec.Content
		require.NotNil(t, content, "Response content should not be nil")

		// Check application/json
		jsonContent, exists := content["application/json"]
		require.True(t, exists, "application/json content should exist")
		require.NotNil(t, jsonContent.Spec.Schema, "Schema should not be nil")
		require.NotNil(t, jsonContent.Spec.Schema.Ref, "Schema ref should not be nil")
		require.Equal(t, "#/components/schemas/model.OrderRow", jsonContent.Spec.Schema.Ref.Ref)

		// Check application/vnd.api+json
		apiJsonContent, exists := content["application/vnd.api+json"]
		require.True(t, exists, "application/vnd.api+json content should exist")
		require.NotNil(t, apiJsonContent.Spec.Schema, "Schema should not be nil")
		require.NotNil(t, apiJsonContent.Spec.Schema.Ref, "Schema ref should not be nil")
		require.Equal(t, "#/components/schemas/model.OrderRow", apiJsonContent.Spec.Schema.Ref.Ref)
	})
}
