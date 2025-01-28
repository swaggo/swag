package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEmptyAsyncComment(t *testing.T) {
	t.Parallel()

	scope := NewAsyncScope(nil)
	err := scope.ParseAsyncAPIComment(nil, "//", nil)

	assert.NoError(t, err)
}

func TestNewAsyncScope(t *testing.T) {
	t.Parallel()
	t.Run("creates a new AsyncScope instance with default properties", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)

		assert.NotNil(t, asyncScope)
		assert.NotNil(t, asyncScope.servers)
		assert.NotNil(t, asyncScope.channels)
		assert.NotNil(t, asyncScope.operations)
		assert.NotNil(t, asyncScope.parser)
	})
}

func TestParseServerComment(t *testing.T) {
	t.Parallel()
	t.Run("parses a valid @server comment", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		comment := "@server myServer mqtt mqtt://broker.hivemq.com"

		err := asyncScope.ParseAsyncAPIComment(nil, comment, nil)

		assert.NoError(t, err)
		assert.Contains(t, asyncScope.servers, "myServer")

		assert.Equal(t, "mqtt", asyncScope.servers["myServer"].Server.Protocol)
		assert.Equal(t, "mqtt://broker.hivemq.com", asyncScope.servers["myServer"].Server.URL)
	})

	t.Run("returns error for invalid @server comment", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		comment := "@server myServer mqtt"

		err := asyncScope.ParseAsyncAPIComment(nil, comment, nil)

		assert.Error(t, err)
		assert.Equal(t, "missing required param comment parameters \"myServer mqtt\"", err.Error())
	})
}

func TestParseChannelComment(t *testing.T) {
	t.Parallel()
	t.Run("parses a valid @channel comment", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		comment := `@channel topic1 myServer "This is a test channel"`

		err := asyncScope.ParseAsyncAPIComment(nil, comment, nil)

		assert.NoError(t, err)
		assert.Contains(t, asyncScope.channels, "topic1")
		assert.Equal(t, "myServer", asyncScope.channels["topic1"].Servers[0])
		assert.Equal(t, "This is a test channel", asyncScope.channels["topic1"].Description)
	})

	t.Run("returns error for invalid @channel comment", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		comment := `@channel topic1 myServer`

		err := asyncScope.ParseAsyncAPIComment(nil, comment, nil)

		assert.Error(t, err)
		assert.Equal(t, "missing required param comment parameters \"topic1 myServer\"", err.Error())
	})
}

func TestParseOperationComment(t *testing.T) {
	t.Parallel()
	t.Run("parses a valid @operation comment", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		asyncScope.parser.addTestType("model.OrderRow")

		comment := `@operation myOperation send topic1 model.OrderRow`
		err := asyncScope.ParseAsyncAPIComment(nil, comment, nil)

		assert.NoError(t, err)
		assert.Contains(t, asyncScope.operations, "myOperation")
		assert.Equal(t, Send, asyncScope.operations["myOperation"].action)
		assert.Equal(t, "topic1", asyncScope.operations["myOperation"].channel)
	})

	t.Run("parses a valid @operation comment - funcName used as operationID", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		asyncScope.parser.addTestType("model.OrderRow")
		funcName := "myOperation"

		comment := `@operation send topic1 model.OrderRow`
		err := asyncScope.ParseAsyncAPIComment(&funcName, comment, nil)

		assert.NoError(t, err)
		assert.Contains(t, asyncScope.operations, "myOperation")
		assert.Equal(t, Send, asyncScope.operations["myOperation"].action)
		assert.Equal(t, "topic1", asyncScope.operations["myOperation"].channel)
	})

	t.Run("returns error for invalid @operation comment", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		comment := `@operation myOperation invalid topic1 model.OrderRow`

		err := asyncScope.ParseAsyncAPIComment(nil, comment, nil)

		assert.Error(t, err)
		assert.Equal(t, "invalid operation action 'invalid' in comment line 'myOperation invalid topic1 model.OrderRow'. Valid values are 'send' or 'receive'", err.Error())
	})

	t.Run("returns error for invalid @operation comment - missing params", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		comment := `@operation model.OrderRow`

		err := asyncScope.ParseAsyncAPIComment(nil, comment, nil)

		assert.Error(t, err)
		assert.Equal(t, "missing required comment parameters: \"model.OrderRow\"", err.Error())
	})

	t.Run("returns error for invalid @operation comment - unable to define operationID", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		comment := `@operation myOperation topic1 model.OrderRow`

		err := asyncScope.ParseAsyncAPIComment(nil, comment, nil)

		assert.Error(t, err)
		assert.Equal(t, "unable to get operation ID from comment line", err.Error())
	})
}

func TestReplaceStringInJSON(t *testing.T) {
	t.Parallel()
	t.Run("replaces occurrences of a string in JSON", func(t *testing.T) {
		input := []byte(`{"$ref": "#/definitions/MyType"}`)
		expected := []byte(`{"$ref": "#/components/schemas/MyType"}`)

		result, err := replaceStringInJSON(input, "#/definitions/", "#/components/schemas/")

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func TestInvalidCommentAttr(t *testing.T) {
	t.Parallel()
	t.Run("returns error if unknown attr", func(t *testing.T) {
		asyncScope := NewAsyncScope(nil)
		comment := "@unknownAttr somevalue"

		err := asyncScope.ParseAsyncAPIComment(nil, comment, nil)
		assert.Error(t, err)
		assert.Equal(t, "unknown attribute '@unknownAttr' in comment '@unknownAttr somevalue'", err.Error())
	})
}
