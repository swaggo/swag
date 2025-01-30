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

	tests := []struct {
		name        string
		comment     string
		funcName    *string
		expectedErr string
		assertFunc  func(t *testing.T, err error, asyncScope *AsyncScope)
	}{
		{
			name:        "parses a valid @server comment",
			comment:     "@server myServer mqtt mqtt://broker.hivemq.com",
			expectedErr: "",
			assertFunc: func(t *testing.T, err error, asyncScope *AsyncScope) {
				assert.Contains(t, asyncScope.servers, "myServer")
				assert.Equal(t, "mqtt", asyncScope.servers["myServer"].Server.Protocol)
				assert.Equal(t, "mqtt://broker.hivemq.com", asyncScope.servers["myServer"].Server.URL)
			},
		},
		{
			name:        "returns error for invalid @server comment",
			comment:     "@server myServer mqtt",
			expectedErr: "missing required param comment parameters \"myServer mqtt\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asyncScope := NewAsyncScope(nil)
			err := asyncScope.ParseAsyncAPIComment(tt.funcName, tt.comment, nil)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
				tt.assertFunc(t, err, asyncScope)
			}
		})
	}
}

func TestParseChannelComment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		comment     string
		expectedErr string
		funcName    *string
		assertFunc  func(t *testing.T, err error, asyncScope *AsyncScope)
	}{
		{
			name:        "parses a valid @channel comment",
			comment:     `@channel topic1 myServer "This is a test channel"`,
			expectedErr: "",
			assertFunc: func(t *testing.T, err error, asyncScope *AsyncScope) {
				assert.Contains(t, asyncScope.channels, "topic1")
				assert.Equal(t, "myServer", asyncScope.channels["topic1"].Servers[0])
				assert.Equal(t, "This is a test channel", asyncScope.channels["topic1"].Description)
			},
		},
		{
			name:        "returns error for invalid @channel comment",
			comment:     `@channel topic1 myServer`,
			expectedErr: "missing required param comment parameters \"topic1 myServer\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asyncScope := NewAsyncScope(nil)
			err := asyncScope.ParseAsyncAPIComment(tt.funcName, tt.comment, nil)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
				tt.assertFunc(t, err, asyncScope)
			}
		})
	}
}

func TestParseOperationComment(t *testing.T) {
	t.Parallel()
	funcNameExample := "myOperation"

	tests := []struct {
		name        string
		comment     string
		expectedErr string
		funcName    *string
		assertFunc  func(t *testing.T, err error, asyncScope *AsyncScope)
	}{
		{
			name:        "parses a valid @operation comment",
			comment:     `@operation myOperation send topic1 model.OrderRow`,
			expectedErr: "",
			assertFunc: func(t *testing.T, err error, asyncScope *AsyncScope) {
				assert.Contains(t, asyncScope.operations, "myOperation")
				assert.Equal(t, Send, asyncScope.operations["myOperation"].action)
				assert.Equal(t, "topic1", asyncScope.operations["myOperation"].channel)
			},
		},
		{
			name:        "parses a valid @operation comment - funcName used as operationID",
			comment:     `@operation send topic1 model.OrderRow`,
			expectedErr: "",
			funcName: 	&funcNameExample,
			assertFunc: func(t *testing.T, err error, asyncScope *AsyncScope) {
				assert.Contains(t, asyncScope.operations, "myOperation")
				assert.Equal(t, Send, asyncScope.operations["myOperation"].action)
				assert.Equal(t, "topic1", asyncScope.operations["myOperation"].channel)
			},
		},
		{
			name:        "returns error for invalid @operation comment",
			comment:     `@operation myOperation invalid topic1 model.OrderRow`,
			expectedErr: "invalid operation action 'invalid' in comment line 'myOperation invalid topic1 model.OrderRow'. Valid values are 'send' or 'receive'",
		},
		{
			name:        "returns error for invalid @operation comment - missing params",
			comment:     `@operation model.OrderRow`,
			expectedErr: "missing required comment parameters: \"model.OrderRow\"",
		},
		{
			name:        "returns error for invalid @operation comment - unable to define operationID",
			comment:     `@operation myOperation topic1 model.OrderRow`,
			expectedErr: "unable to get operation ID from comment line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asyncScope := NewAsyncScope(nil)
			asyncScope.parser.addTestType("model.OrderRow")

			err := asyncScope.ParseAsyncAPIComment(tt.funcName, tt.comment, nil)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			}
			if tt.assertFunc != nil {
				tt.assertFunc(t, err, asyncScope)
			}
		})
	}
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
