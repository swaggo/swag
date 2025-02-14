package swag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"log"
	"regexp"
	"strings"

	typeSpec "github.com/go-openapi/spec"
	"github.com/swaggest/go-asyncapi/spec-2.4.0"
)

type OperationAction string
type Attribute string

const (
	Send    OperationAction = "send"
	Receive OperationAction = "receive"
)

const (
	serverAttr Attribute = "@server"
	channelAttr Attribute = "@channel"
	operationAttr Attribute = "@operation"
)

const (
	openAPISchemaPrefix = "#/definitions/"
	asyncAPISchemaPrefix = "#/components/schemas/"
)

type AsyncScope struct {
	parser *Parser
	servers map[string]*spec.ServersAdditionalProperties
	channels map[string]*spec.ChannelItem
	operations map[string]*OperationWithChannel
}

type OperationWithChannel struct {
	action OperationAction
	channel string
	spec.Operation
}

// NewAsyncOperation creates a new AsyncOperation with default properties.
func NewAsyncScope(parser *Parser) *AsyncScope {
	if parser == nil {
		parser = New()
	}

	asyncOperation := &AsyncScope{
		parser:           parser,
		servers:          make(map[string]*spec.ServersAdditionalProperties),
		channels:         make(map[string]*spec.ChannelItem),
		operations:       make(map[string]*OperationWithChannel),
	}

	return asyncOperation
}

// AttributeHandler is a map of attribute to the function that handles the attribute.
var AttributeHandler = map[Attribute]func(*AsyncScope, *string, string, *ast.File) error {
	serverAttr:  (*AsyncScope).ParseServerComment,
	channelAttr: (*AsyncScope).ParseChannelComment,
	operationAttr: (*AsyncScope).ParseOperationComment,
}

// ParseAsyncAPIComment parses the comment line and sets the AsyncAPI properties.
func (asyncScope *AsyncScope) ParseAsyncAPIComment(funcName *string, comment string, astFile *ast.File) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "/"))
	if len(commentLine) == 0 {
		return nil
	}

	fields := FieldsByAnySpace(commentLine, 2)
	attribute := fields[0]
	lowerAttribute := strings.ToLower(attribute)
	
	var lineRemainder string
	if len(fields) > 1 {
		lineRemainder = fields[1]
	}

	handler, exists := AttributeHandler[Attribute(lowerAttribute)]
	if !exists {
		return fmt.Errorf("unknown attribute '%s' in comment '%s'", attribute, comment)
	}

	return handler(asyncScope, funcName, lineRemainder, astFile)
}

var serverCommentPattern = regexp.MustCompile(`(\S+)\s+(\S+)\s+(\S+)`)

// @server {name} {protocol} {host}
func (asyncScope *AsyncScope) ParseServerComment(funcName *string, commentLine string, astFile *ast.File) error {
	matches := serverCommentPattern.FindStringSubmatch(commentLine)
	if len(matches) < 4 {
		return fmt.Errorf("missing required param comment parameters \"%s\"", commentLine)
	}

	serverName := matches[1]
	protocol := matches[2]
	host := matches[3]

	asyncScope.servers[serverName] = &spec.ServersAdditionalProperties{
		Server: &spec.Server{
			URL: host,
			Protocol: protocol,
		},
	}

	return nil
}

var channelCommentPattern = regexp.MustCompile(`(\S+)\s+(\S+)\s+"([^"]+)"`)

// @channel {name/topic} {server} "{description}"
func (asyncScope *AsyncScope) ParseChannelComment(funcName *string, commentLine string, astFile *ast.File) error {
	matches := channelCommentPattern.FindStringSubmatch(commentLine)
	if len(matches) < 4 {
		return fmt.Errorf("missing required param comment parameters \"%s\"", commentLine)
	}
	
	channelName := matches[1]
	server := matches[2]
	description := matches[3]

	asyncScope.channels[channelName] = &spec.ChannelItem{
		Servers: []string{server},
		Description: description,
	}

	return nil
}

var operationCommentPattern = regexp.MustCompile(`(\S+)\s+(\S+)\s+(\S+)\s*(.*)?`)


// @operation {operationID} {action} {channel} {message}
// @operation {action} {channel} {message}
func (asyncScope *AsyncScope) ParseOperationComment(funcName *string, commentLine string, astFile *ast.File) error {
	matches, err := asyncScope.validateOperationCommentLine(commentLine)
	if err != nil {
		return err
	}

	operationID, argsStartIndex, err := asyncScope.getOperationID(funcName, matches)
	if err != nil {
		return err
	}

	operationAction, err := asyncScope.validateOperationAction(matches[argsStartIndex], commentLine)
	if err != nil {
		return err
	}

	channel := matches[argsStartIndex+1]
	message := matches[argsStartIndex+2]

	typeSchema, err := asyncScope.parser.getTypeSchema(message, astFile, false, true)
	if err != nil {
		log.Printf("unable to get type schema for message type '%s': %v", message, err)
		return err
	}

	msg, err := asyncScope.createMessage(typeSchema, formatMessageID(message))
	if err != nil {
		return err
	}

	asyncScope.addOperation(operationID, operationAction, channel, msg)
	return nil
}

// Validates the comment line and ensures it matches the required pattern.
func (asyncScope *AsyncScope) validateOperationCommentLine(commentLine string) ([]string, error) {
	matches := operationCommentPattern.FindStringSubmatch(commentLine)
	if len(matches) < 5 {
		return nil, fmt.Errorf("missing required comment parameters: \"%s\"", commentLine)
	}
	return matches, nil
}

// Get the operation ID and the argument start index.
func (asyncScope *AsyncScope) getOperationID(funcName *string, matches []string) (string, int, error) {
	if matches[4] == "" {
		if funcName == nil {
			return "", 0, fmt.Errorf("unable to get operation ID from comment line")
		}
		return *funcName, 1, nil
	}
	return matches[1], 2, nil
}

// Validates the operation kind and ensures it is either "send" or "receive".
func (asyncScope *AsyncScope) validateOperationAction(action string, commentLine string) (OperationAction, error) {
	if action == string(Send) || action == string(Receive) {
		return OperationAction(action), nil
	}
	return "", fmt.Errorf("invalid operation action '%s' in comment line '%s'. Valid values are 'send' or 'receive'", action, commentLine)
}

// Creates a message based on the type schema and message ID.
func (asyncScope *AsyncScope) createMessage(typeSchema *typeSpec.Schema, messageID string) (spec.Message, error) {
	msg := spec.Message{}

	payload := map[string]interface{}{"type": typeSchema.Type[0]}
	if typeSchema.Type[0] == OBJECT {
		properties, err := GetAsyncAPISchemaProperties(typeSchema)
		if err != nil {
			return msg, err
		}
		payload["properties"] = properties
	}

	msg.OneOf1Ens().WithMessageEntity(spec.MessageEntity{
		MessageID: messageID,
		Payload:   payload,
	})

	return msg, nil
}

// formatMessageID extracts the message type from the package name (e.g., "package.MessageType" -> "MessageType").
func formatMessageID(messageID string) string {
	if i := strings.LastIndex(messageID, "."); i != -1 {
		return messageID[i+1:]
	}
	return messageID
}

// Marshals and processes asyncAPI type schema properties.
func GetAsyncAPISchemaProperties(typeSchema *typeSpec.Schema) (map[string]interface{}, error) {
	jsonData, err := typeSchema.Properties.MarshalJSON()
	if err != nil {
		return nil, err
	}

	jsonData, _ = replaceStringInJSON(jsonData, openAPISchemaPrefix, asyncAPISchemaPrefix)

	var properties map[string]interface{}
	if err := json.Unmarshal(jsonData, &properties); err != nil {
		return nil, err
	}

	return properties, nil
}

// Adds an operation to the async scope.
func (asyncScope *AsyncScope) addOperation(operationID string, action OperationAction, channel string, msg spec.Message) {
	operation := spec.Operation{}
	operation.WithID(operationID).WithMessage(msg)

	asyncScope.operations[operationID] = &OperationWithChannel{
		action:    action,
		channel:   channel,
		Operation: operation,
	}
}

// Replace all occurrences of oldValue with newValue
func replaceStringInJSON(originalJSON []byte, oldValue, newValue string) ([]byte, error) {
	updatedJSON := bytes.ReplaceAll(originalJSON, []byte(oldValue), []byte(newValue))
	return updatedJSON, nil
}
