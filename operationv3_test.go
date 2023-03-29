package swag

import (
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

	response := operation.Responses.Spec.Response["200"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

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
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token2"].Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token2"].Spec.Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Response["201"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token2"].Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token2"].Spec.Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Default
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token2"].Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token2"].Spec.Spec.Schema.Spec.Type)

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
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Response["201"]
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

	response = operation.Responses.Spec.Default
	assert.NotNil(t, response)
	assert.NotNil(t, response.Spec)

	assert.Equal(t, "it's ok", response.Spec.Spec.Description)

	assert.Equal(t, "qwerty", response.Spec.Spec.Headers["Token"].Spec.Spec.Description)
	assert.Equal(t, typeString, response.Spec.Spec.Headers["Token"].Spec.Spec.Schema.Spec.Type)

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
		for _, paramType := range []string{"header", "path", "query", "formData"} {
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
											Type: typeInteger,
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
		for _, paramType := range []string{"header", "path", "query", "formData"} {
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
											Type: typeString,
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
		for _, paramType := range []string{"header", "path", "query", "formData"} {
			t.Run(paramType, func(t *testing.T) {
				assert.Error(t,
					NewOperationV3(New()).
						ParseComment(`@Param some_object `+paramType+` main.Object true "Some Object"`,
							nil))
			})
		}
	})

}

// Test ParseParamComment Query Params
func TestParseParamCommentBodyArrayV3(t *testing.T) {
	t.Parallel()

	comment := `@Param names body []string true "Users List"`
	o := NewOperationV3(New())
	err := o.ParseComment(comment, nil)
	assert.NoError(t, err)

	expected := &spec.RefOrSpec[spec.Extendable[spec.Parameter]]{
		Spec: &spec.Extendable[spec.Parameter]{
			Spec: &spec.Parameter{
				Name:        "names",
				Description: "Users List",
				In:          "body",
				Required:    true,
				Schema: &spec.RefOrSpec[spec.Schema]{
					Spec: &spec.Schema{
						JsonSchema: spec.JsonSchema{
							JsonSchemaCore: spec.JsonSchemaCore{
								Type: typeArray,
							},
							JsonSchemaTypeArray: spec.JsonSchemaTypeArray{
								Items: &spec.BoolOrSchema{
									Schema: &spec.RefOrSpec[spec.Schema]{
										Spec: &spec.Schema{
											JsonSchema: spec.JsonSchema{
												JsonSchemaCore: spec.JsonSchemaCore{
													Type: typeString,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	expectedArray := []*spec.RefOrSpec[spec.Extendable[spec.Parameter]]{expected}
	assert.Equal(t, o.Parameters, expectedArray)

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
			assert.Equal(t, typeArray, parameterSpec.Schema.Spec.Type)
			assert.Equal(t, true, parameterSpec.Required)
			assert.Equal(t, paramType, parameterSpec.In)
			assert.Equal(t, typeString, parameterSpec.Schema.Spec.Items.Schema.Spec.Type)

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
	assert.Equal(t, typeString, parameterSpec.Schema.Spec.Type)
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
	assert.Equal(t, typeArray, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, true, parameterSpec.Required)
	assert.Equal(t, "query", parameterSpec.In)
	assert.Equal(t, typeString, parameterSpec.Schema.Spec.Items.Schema.Spec.Type)
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
	assert.Equal(t, typeInteger, parameterSpec.Schema.Spec.Type)
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
	assert.Equal(t, typeInteger, parameterSpec.Schema.Spec.Type)
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

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Some ID", parameterSpec.Description)
	assert.Equal(t, "some_id", parameterSpec.Name)
	assert.Equal(t, true, parameterSpec.Required)
	assert.Equal(t, "body", parameterSpec.In)
	assert.Equal(t, "#/components/model.OrderRow", parameterSpec.Schema.Ref.Ref)
}

func TestParseParamCommentByBodyTextPlainV3(t *testing.T) {
	t.Parallel()

	comment := `@Param text body string true "Text to process"`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "Text to process", parameterSpec.Description)
	assert.Equal(t, "text", parameterSpec.Name)
	assert.Equal(t, true, parameterSpec.Required)
	assert.Equal(t, "body", parameterSpec.In)
	assert.Equal(t, typeString, parameterSpec.Schema.Spec.Type)
}

func TestParseParamCommentByBodyTypeWithDeepNestedFieldsV3(t *testing.T) {
	t.Parallel()

	comment := `@Param body body model.CommonHeader{data=string,data2=int} true "test deep"`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Len(t, operation.Parameters, 1)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "test deep", parameterSpec.Description)
	assert.Equal(t, "body", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "body", parameterSpec.In)

	assert.Equal(t, 2, len(parameterSpec.Schema.Spec.AllOf))
	assert.Equal(t, 3, len(operation.parser.openAPI.Components.Spec.Schemas))
}

func TestParseParamCommentByBodyTypeArrayOfPrimitiveGoV3(t *testing.T) {
	t.Parallel()

	comment := `@Param some_id body []int true "Some ID"`
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
	assert.Equal(t, "body", parameterSpec.In)
	assert.Equal(t, typeArray, parameterSpec.Schema.Spec.Type)
	assert.Equal(t, typeInteger, parameterSpec.Schema.Spec.Items.Schema.Spec.Type)
}

func TestParseParamCommentByBodyTypeArrayOfPrimitiveGoWithDeepNestedFieldsV3(t *testing.T) {
	t.Parallel()

	comment := `@Param body body []model.CommonHeader{data=string,data2=int} true "test deep"`
	operation := NewOperationV3(New())
	operation.parser.addTestType("model.CommonHeader")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Len(t, operation.Parameters, 1)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "test deep", parameterSpec.Description)
	assert.Equal(t, "body", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "body", parameterSpec.In)
	assert.Equal(t, typeArray, parameterSpec.Schema.Spec.Type)
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

	assert.Len(t, operation.Parameters, 1)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "this is a test file", parameterSpec.Description)
	assert.Equal(t, "file", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "formData", parameterSpec.In)
	assert.Equal(t, typeFile, parameterSpec.Schema.Spec.Type)
}

func TestParseParamCommentByFormDataTypeUint64V3(t *testing.T) {
	t.Parallel()

	comment := `@Param file formData uint64 true "this is a test file"`
	operation := NewOperationV3(New())

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Len(t, operation.Parameters, 1)

	parameters := operation.Operation.Parameters
	assert.NotNil(t, parameters)

	parameterSpec := parameters[0].Spec.Spec
	assert.NotNil(t, parameterSpec)
	assert.Equal(t, "this is a test file", parameterSpec.Description)
	assert.Equal(t, "file", parameterSpec.Name)
	assert.True(t, parameterSpec.Required)
	assert.Equal(t, "formData", parameterSpec.In)
	assert.Equal(t, typeInteger, parameterSpec.Schema.Spec.Type)
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
