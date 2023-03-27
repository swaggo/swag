package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sv-tools/openapi/spec"
)

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
	assert.NoError(t, err)

	err = operation.ParseComment(anotherComment, nil)
	assert.NoError(t, err)

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
	assert.NoError(t, err)
	assert.Len(t, operation.RouterProperties, 1)
	assert.Equal(t, "/", operation.RouterProperties[0].Path)
	assert.Equal(t, "GET", operation.RouterProperties[0].HTTPMethod)
}

func TestParseRouterCommentWithPlusSignV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router /customer/get-wishlist/{proxy+} [post]`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
	assert.Len(t, operation.RouterProperties, 1)
	assert.Equal(t, "/customer/get-wishlist/{proxy+}", operation.RouterProperties[0].Path)
	assert.Equal(t, "POST", operation.RouterProperties[0].HTTPMethod)
}

func TestParseRouterCommentWithDollarSignV3(t *testing.T) {
	t.Parallel()

	comment := `/@Router /customer/get-wishlist/{wishlist_id}$move [post]`
	operation := NewOperationV3(nil)
	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)

	assert.Equal(t, "An empty response", operation.Responses.Spec.Default.Spec.Spec.Description)

	comment = `@Success 200,default {string} Response "A response"`
	operation = NewOperationV3(nil)

	err = operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Equal(t, "A response", operation.Responses.Spec.Default.Spec.Spec.Description)
	assert.Equal(t, "A response", operation.Responses.Spec.Response["200"].Spec.Spec.Description)
}

func TestParseResponseSuccessCommentWithEmptyResponseV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} nil "An empty response"`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `An empty response`, response.Spec.Spec.Description)
}

func TestParseResponseFailureCommentWithEmptyResponseV3(t *testing.T) {
	t.Parallel()

	comment := `@Failure 500 {object} nil`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	assert.Equal(t, "Internal Server Error", operation.Responses.Spec.Response["500"].Spec.Spec.Description)
}

func TestParseResponseCommentWithObjectTypeV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.OrderRow "Error message, if code != 200`
	parser := New()
	operation := NewOperationV3(parser)
	operation.parser.addTestType("model.OrderRow")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

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
	assert.NoError(t, err)

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
	assert.NoError(t, err)

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
	assert.NoError(t, err)

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
	assert.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)

	allOf := response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.AllOf
	assert.Equal(t, 2, len(allOf))

	assert.Equal(t, "#/components/model.Payload", operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Spec.Items.Schema.Ref.Ref)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"array"}), operation.parser.openAPI.Components.Spec.Schemas["data"].Spec.Properties["data"].Spec.Type)

	assert.Equal(t, "#/components/model.Payload2", operation.parser.openAPI.Components.Spec.Schemas["data2"].Spec.Properties["data2"].Spec.Items.Schema.Ref.Ref)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"array"}), operation.parser.openAPI.Components.Spec.Schemas["data2"].Spec.Properties["data2"].Spec.Type)
}

func TestParseResponseCommentWithNestedFieldsV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data1=int,data2=[]int,data3=model.Payload,data4=[]model.Payload} "Error message, if code != 200`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)

	allOf := response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.AllOf
	assert.Equal(t, 4, len(allOf))

	schemas := operation.parser.openAPI.Components.Spec.Schemas

	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"integer"}), schemas["data1"].Spec.Properties["data1"].Spec.Type)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"object"}), schemas["data1"].Spec.Type)

	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"array"}), schemas["data2"].Spec.Properties["data2"].Spec.Type)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"integer"}), schemas["data2"].Spec.Properties["data2"].Spec.Items.Schema.Spec.Type)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"object"}), schemas["data2"].Spec.Type)

	assert.Equal(t, "#/components/model.Payload", schemas["data3"].Spec.Properties["data3"].Ref.Ref)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"object"}), schemas["data3"].Spec.Type)

	assert.Equal(t, "#/components/model.Payload", schemas["data4"].Spec.Properties["data4"].Spec.Items.Schema.Ref.Ref)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"array"}), schemas["data4"].Spec.Properties["data4"].Spec.Type)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"object"}), schemas["data4"].Spec.Type)
}

func TestParseResponseCommentWithDeepNestedFieldsV3(t *testing.T) {
	t.Parallel()

	comment := `@Success 200 {object} model.CommonHeader{data1=int,data2=[]int,data3=model.Payload{data1=int,data2=model.DeepPayload},data4=[]model.Payload{data1=[]int,data2=[]model.DeepPayload}} "Error message, if code != 200`
	operation := NewOperationV3(New())

	operation.parser.addTestType("model.CommonHeader")
	operation.parser.addTestType("model.Payload")
	operation.parser.addTestType("model.DeepPayload")

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	response := operation.Responses.Spec.Response["200"]
	assert.Equal(t, `Error message, if code != 200`, response.Spec.Spec.Description)

	allOf := response.Spec.Spec.Content["application/json"].Spec.Schema.Spec.AllOf
	assert.Equal(t, 4, len(allOf))

	schemas := operation.parser.openAPI.Components.Spec.Schemas

	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"integer"}), schemas["data1"].Spec.Properties["data1"].Spec.Type)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"object"}), schemas["data1"].Spec.Type)

	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"array"}), schemas["data2"].Spec.Properties["data2"].Spec.Type)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"integer"}), schemas["data2"].Spec.Properties["data2"].Spec.Items.Schema.Spec.Type)
	assert.Equal(t, spec.SingleOrArray[string](spec.SingleOrArray[string]{"object"}), schemas["data2"].Spec.Type)

}
