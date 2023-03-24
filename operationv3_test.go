package swag

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	b, _ := json.MarshalIndent(operation, "", "    ")
	expected := `{
    "responses": {
        "200": {
            "description": "An empty response"
        }
    }
}`
	assert.Equal(t, expected, string(b))
}

func TestParseResponseFailureCommentWithEmptyResponseV3(t *testing.T) {
	t.Parallel()

	comment := `@Failure 500 {object} nil`
	operation := NewOperationV3(nil)

	err := operation.ParseComment(comment, nil)
	assert.NoError(t, err)

	b, _ := json.MarshalIndent(operation, "", "    ")
	expected := `{
    "responses": {
        "500": {
            "description": "Internal Server Error"
        }
    }
}`
	assert.Equal(t, expected, string(b))
}
