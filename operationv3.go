package swag

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sv-tools/openapi/spec"
	"gopkg.in/yaml.v3"
)

// OperationV3 describes a single API operation on a path.
// For more information: https://github.com/swaggo/swag#api-operation
type OperationV3 struct {
	parser              *Parser
	codeExampleFilesDir string
	spec.Operation
	RouterProperties  []RouteProperties
	responseMimeTypes []string
}

// NewOperationV3 returns a new instance of OperationV3.
func NewOperationV3(parser *Parser, options ...func(*OperationV3)) *OperationV3 {
	op := *spec.NewOperation().Spec
	op.Responses = spec.NewResponses()

	operation := &OperationV3{
		parser:    parser,
		Operation: op,
	}

	for _, option := range options {
		option(operation)
	}

	return operation
}

// SetCodeExampleFilesDirectoryV3 sets the directory to search for codeExamples.
func SetCodeExampleFilesDirectoryV3(directoryPath string) func(*OperationV3) {
	return func(o *OperationV3) {
		o.codeExampleFilesDir = directoryPath
	}
}

// ParseComment parses comment for given comment string and returns error if error occurs.
func (o *OperationV3) ParseComment(comment string, astFile *ast.File) error {
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
	switch lowerAttribute {
	case descriptionAttr:
		o.ParseDescriptionComment(lineRemainder)
	case descriptionMarkdownAttr:
		commentInfo, err := getMarkdownForTag(lineRemainder, o.parser.markdownFileDir)
		if err != nil {
			return err
		}

		o.ParseDescriptionComment(string(commentInfo))
	case summaryAttr:
		o.Summary = lineRemainder
	case idAttr:
		o.OperationID = lineRemainder
	case tagsAttr:
		o.ParseTagsComment(lineRemainder)
	case acceptAttr:
		return o.ParseAcceptComment(lineRemainder)
	case produceAttr:
		return o.ParseProduceComment(lineRemainder)
	case paramAttr:
		return o.ParseParamComment(lineRemainder, astFile)
	case successAttr, failureAttr, responseAttr:
		return o.ParseResponseComment(lineRemainder, astFile)
	case headerAttr:
		return o.ParseResponseHeaderComment(lineRemainder, astFile)
	case routerAttr:
		return o.ParseRouterComment(lineRemainder)
	case securityAttr:
		return o.ParseSecurityComment(lineRemainder)
	case deprecatedAttr:
		o.Deprecated = true
	case xCodeSamplesAttr, xCodeSamplesAttrOriginal:
		return o.ParseCodeSample(attribute, commentLine, lineRemainder)
	case "@servers.url":
		return o.ParseServerURLComment(lineRemainder)
	case "@servers.description":
		return o.ParseServerDescriptionComment(lineRemainder)
	default:
		return o.ParseMetadata(attribute, lowerAttribute, lineRemainder)
	}

	return nil
}

// ParseDescriptionComment parses the description comment and sets it to the operation.
func (o *OperationV3) ParseDescriptionComment(lineRemainder string) {
	if o.Description == "" {
		o.Description = lineRemainder

		return
	}

	o.Description += "\n" + lineRemainder
}

// ParseMetadata godoc.
func (o *OperationV3) ParseMetadata(attribute, lowerAttribute, lineRemainder string) error {
	// parsing specific meta data extensions
	if strings.HasPrefix(lowerAttribute, "@x-") {
		if len(lineRemainder) == 0 {
			return fmt.Errorf("annotation %s need a value", attribute)
		}

		var valueJSON any

		err := json.Unmarshal([]byte(lineRemainder), &valueJSON)
		if err != nil {
			return fmt.Errorf("annotation %s need a valid json value. error: %s", attribute, err.Error())
		}

		o.Responses.Extensions[attribute[1:]] = valueJSON
		return nil
	}

	return nil
}

// ParseTagsComment parses comment for given `tag` comment string.
func (o *OperationV3) ParseTagsComment(commentLine string) {
	for _, tag := range strings.Split(commentLine, ",") {
		o.Tags = append(o.Tags, strings.TrimSpace(tag))
	}
}

// ParseAcceptComment parses comment for given `accept` comment string.
func (o *OperationV3) ParseAcceptComment(commentLine string) error {
	const errMessage = "could not parse accept comment"

	validTypes, err := parseMimeTypeListV3(commentLine, "%v accept type can't be accepted")
	if err != nil {
		return fmt.Errorf("%s: %w", errMessage, err)
	}

	if o.RequestBody == nil {
		o.RequestBody = spec.NewRequestBodySpec()
	}

	if o.RequestBody.Spec.Spec.Content == nil {
		o.RequestBody.Spec.Spec.Content = make(map[string]*spec.Extendable[spec.MediaType], len(validTypes))
	}

	for _, value := range validTypes {
		// skip correctly setup types like application/json
		if o.RequestBody.Spec.Spec.Content[value] != nil {
			continue
		}

		mediaType := spec.NewMediaType()
		schema := spec.NewSchemaSpec()

		switch value {
		case "application/json", "multipart/form-data", "text/xml":
			schema.Spec.Type = &spec.SingleOrArray[string]{OBJECT}
		case "image/png",
			"image/jpeg",
			"image/gif",
			"application/octet-stream",
			"application/pdf",
			"application/msexcel",
			"application/zip",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation":
			schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
			schema.Spec.Format = "binary"
		default:
			schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
		}

		mediaType.Spec.Schema = schema
		o.RequestBody.Spec.Spec.Content[value] = mediaType
	}

	return nil
}

// ParseProduceComment parses comment for given `produce` comment string.
func (o *OperationV3) ParseProduceComment(commentLine string) error {
	const errMessage = "could not parse produce comment"

	validTypes, err := parseMimeTypeListV3(commentLine, "%v produce type can't be accepted")
	if err != nil {
		return fmt.Errorf("%s: %w", errMessage, err)
	}

	o.responseMimeTypes = validTypes

	return nil
}

// ProcessProduceComment processes the previously parsed produce comment.
func (o *OperationV3) ProcessProduceComment() error {
	const errMessage = "could not process produce comment"

	if o.Responses == nil {
		return nil
	}

	for _, value := range o.responseMimeTypes {
		if o.Responses.Spec.Response == nil {
			o.Responses.Spec.Response = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Response]], len(o.responseMimeTypes))
		}

		for key, response := range o.Responses.Spec.Response {
			code, err := strconv.Atoi(key)
			if err != nil {
				return fmt.Errorf("%s: %w", errMessage, err)
			}

			// Status 204 is no content. So we do not need to add content.
			if code == 204 {
				continue
			}

			// As this is a workaround, we need to check if the code is in range.
			// The Produce comment is being deprecated soon.
			if code < 200 || code > 299 {
				continue
			}

			// skip correctly setup types like application/json
			if response.Spec.Spec.Content[value] != nil {
				continue
			}

			mediaType := spec.NewMediaType()
			schema := spec.NewSchemaSpec()

			switch value {
			case "application/json", "multipart/form-data", "text/xml":
				schema.Spec.Type = &spec.SingleOrArray[string]{OBJECT}
			case "image/png",
				"image/jpeg",
				"image/gif",
				"application/octet-stream",
				"application/pdf",
				"application/msexcel",
				"application/zip",
				"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
				"application/vnd.openxmlformats-officedocument.presentationml.presentation":
				schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
				schema.Spec.Format = "binary"
			default:
				schema.Spec.Type = &spec.SingleOrArray[string]{STRING}
			}

			mediaType.Spec.Schema = schema

			if response.Spec.Spec.Content == nil {
				response.Spec.Spec.Content = make(map[string]*spec.Extendable[spec.MediaType])
			}

			response.Spec.Spec.Content[value] = mediaType

		}
	}

	return nil
}

// parseMimeTypeList parses a list of MIME Types for a comment like
// `produce` (`Content-Type:` response header) or
// `accept` (`Accept:` request header).
func parseMimeTypeListV3(mimeTypeList string, format string) ([]string, error) {
	var result []string
	for _, typeName := range strings.Split(mimeTypeList, ",") {
		typeName = strings.TrimSpace(typeName)

		if mimeTypePattern.MatchString(typeName) {
			result = append(result, typeName)

			continue
		}

		aliasMimeType, ok := mimeTypeAliases[typeName]
		if !ok {
			return nil, fmt.Errorf(format, typeName)
		}

		result = append(result, aliasMimeType)
	}

	return result, nil
}

// ParseParamComment parses params return []string of param properties
// E.g. @Param	queryText		formData	      string	  true		        "The email for login"
//
//	[param name]    [paramType] [data type]  [is mandatory?]   [Comment]
//
// E.g. @Param   some_id     path    int     true        "Some ID".
func (o *OperationV3) ParseParamComment(commentLine string, astFile *ast.File) error {
	matches := paramPattern.FindStringSubmatch(commentLine)
	if len(matches) != 6 {
		return fmt.Errorf("missing required param comment parameters \"%s\"", commentLine)
	}

	name := matches[1]
	paramType := matches[2]
	refType := TransToValidSchemeType(matches[3])

	// Detect refType
	objectType := OBJECT

	if strings.HasPrefix(refType, "[]") {
		objectType = ARRAY
		refType = strings.TrimPrefix(refType, "[]")
		refType = TransToValidSchemeType(refType)
	} else if IsPrimitiveType(refType) ||
		paramType == "formData" && refType == "file" {
		objectType = PRIMITIVE
	}

	var enums []interface{}
	if !IsPrimitiveType(refType) {
		schema, _ := o.parser.getTypeSchemaV3(refType, astFile, false)
		if schema != nil && schema.Spec != nil && schema.Spec.Enum != nil {
			// schema.Spec.Type != ARRAY
			fmt.Println(schema.Spec.Type)

			if objectType == OBJECT {
				objectType = PRIMITIVE
			}
			refType = TransToValidSchemeType((*schema.Spec.Type)[0])
			enums = schema.Spec.Enum
		}
	}

	requiredText := strings.ToLower(matches[4])
	required := requiredText == "true" || requiredText == requiredLabel
	description := matches[5]

	param := createParameterV3(paramType, description, name, objectType, refType, required, enums, o.parser.collectionFormatInQuery)

	switch paramType {
	case "path", "header":
		switch objectType {
		case ARRAY:
			if !IsPrimitiveType(refType) {
				return fmt.Errorf("%s is not supported array type for %s", refType, paramType)
			}
		case OBJECT:
			return fmt.Errorf("%s is not supported type for %s", refType, paramType)
		}
	case "query":
		switch objectType {
		case ARRAY:
			if !IsPrimitiveType(refType) && !(refType == "file" && paramType == "formData") {
				return fmt.Errorf("%s is not supported array type for %s", refType, paramType)
			}
		case PRIMITIVE:
			break
		case OBJECT:
			schema, err := o.parser.getTypeSchemaV3(refType, astFile, false)
			if err != nil {
				return err
			}

			if len(schema.Spec.Properties) == 0 {
				return nil
			}

			for name, item := range schema.Spec.Properties {
				prop := item.Spec
				if len(*prop.Type) == 0 {
					continue
				}

				itemParam := param // Avoid shadowed variable which could cause side effects to o.Operation.Parameters

				switch {
				case (*prop.Type)[0] == ARRAY &&
					prop.Items.Schema != nil &&
					len(*prop.Items.Schema.Spec.Type) > 0 &&
					IsSimplePrimitiveType((*prop.Items.Schema.Spec.Type)[0]):

					itemParam = createParameterV3(paramType, prop.Description, name, (*prop.Type)[0], (*prop.Items.Schema.Spec.Type)[0], findInSlice(schema.Spec.Required, name), enums, o.parser.collectionFormatInQuery)

				case IsSimplePrimitiveType((*prop.Type)[0]):
					itemParam = createParameterV3(paramType, prop.Description, name, PRIMITIVE, (*prop.Type)[0], findInSlice(schema.Spec.Required, name), enums, o.parser.collectionFormatInQuery)
				default:
					o.parser.debug.Printf("skip field [%s] in %s is not supported type for %s", name, refType, paramType)

					continue
				}

				itemParam.Schema.Spec = prop

				listItem := &spec.RefOrSpec[spec.Extendable[spec.Parameter]]{
					Spec: &spec.Extendable[spec.Parameter]{
						Spec: &itemParam,
					},
				}

				o.Operation.Parameters = append(o.Operation.Parameters, listItem)
			}

			return nil
		}
	case "body", "formData":
		if objectType == PRIMITIVE {
			schema := PrimitiveSchemaV3(refType)

			err := o.parseParamAttributeForBody(commentLine, objectType, refType, schema.Spec)
			if err != nil {
				return err
			}

			o.fillRequestBody(name, schema, required, description, true, paramType == "formData")

			return nil

		}

		schema, err := o.parseAPIObjectSchema(commentLine, objectType, refType, astFile)
		if err != nil {
			return err
		}

		err = o.parseParamAttributeForBody(commentLine, objectType, refType, schema.Spec)
		if err != nil {
			return err
		}
		o.fillRequestBody(name, schema, required, description, false, paramType == "formData")

		return nil

	default:
		return fmt.Errorf("%s is not supported paramType", paramType)
	}

	err := o.parseParamAttribute(commentLine, objectType, refType, &param)
	if err != nil {
		return err
	}

	item := spec.NewRefOrSpec(nil, &spec.Extendable[spec.Parameter]{
		Spec: &param,
	})

	o.Operation.Parameters = append(o.Operation.Parameters, item)

	return nil
}

func (o *OperationV3) fillRequestBody(name string, schema *spec.RefOrSpec[spec.Schema], required bool, description string, primitive, formData bool) {
	if o.RequestBody == nil {
		o.RequestBody = spec.NewRequestBodySpec()
		o.RequestBody.Spec.Spec.Content = make(map[string]*spec.Extendable[spec.MediaType])

		if primitive && !formData {
			o.RequestBody.Spec.Spec.Content["text/plain"] = spec.NewMediaType()
		} else if formData {
			o.RequestBody.Spec.Spec.Content["application/x-www-form-urlencoded"] = spec.NewMediaType()
		} else {
			o.RequestBody.Spec.Spec.Content["application/json"] = spec.NewMediaType()
		}
	}

	o.RequestBody.Spec.Spec.Required = required

	// Append description to existing description if this is not the first body
	if o.RequestBody.Spec.Spec.Description != "" && description != "" {
		o.RequestBody.Spec.Spec.Description += " | " + description
	} else if description != "" {
		o.RequestBody.Spec.Spec.Description = description
	}

	// Handle oneOf merging for request body schemas
	contentType := "application/json"
	if primitive && !formData {
		contentType = "text/plain"
	} else if formData {
		contentType = "application/x-www-form-urlencoded"
	}

	mediaType := o.RequestBody.Spec.Spec.Content[contentType]
	if mediaType == nil {
		mediaType = spec.NewMediaType()
		o.RequestBody.Spec.Spec.Content[contentType] = mediaType
	}
	if schema.Ref != nil {
		schema.Ref.Summary = name
		schema.Ref.Description = description
	}
	if schema.Spec != nil {
		schema.Spec.Title = name
	}
	if mediaType.Spec.Schema == nil {
		mediaType.Spec.Schema = schema
	} else if mediaType.Spec.Schema.Ref != nil || mediaType.Spec.Schema.Spec.OneOf == nil {
		// If there's an existing schema that doesn't have oneOf, create a oneOf schema
		oneOfSchema := spec.NewSchemaSpec()
		oneOfSchema.Spec.OneOf = []*spec.RefOrSpec[spec.Schema]{mediaType.Spec.Schema, schema}
		mediaType.Spec.Schema = oneOfSchema
	} else {
		// If there's already a oneOf schema, append to it
		mediaType.Spec.Schema.Spec.OneOf = append(mediaType.Spec.Schema.Spec.OneOf, schema)
	}
}

func (o *OperationV3) parseParamAttribute(comment, objectType, schemaType string, param *spec.Parameter) error {
	if param == nil {
		return fmt.Errorf("cannot parse empty parameter for comment: %s", comment)
	}

	schemaType = TransToValidSchemeType(schemaType)

	for attrKey, re := range regexAttributes {
		attr, err := findAttr(re, comment)
		if err != nil {
			continue
		}

		switch attrKey {
		case enumsTag:
			err = setEnumParamV3(param.Schema.Spec, attr, objectType, schemaType)
		case minimumTag, maximumTag:
			err = setNumberParamV3(param.Schema.Spec, attrKey, schemaType, attr, comment)
		case defaultTag:
			err = setDefaultV3(param.Schema.Spec, schemaType, attr)
		case minLengthTag, maxLengthTag:
			err = setStringParamV3(param.Schema.Spec, attrKey, schemaType, attr, comment)
		case formatTag:
			param.Schema.Spec.Format = attr
		case exampleTag:
			val, err := defineType(schemaType, attr)
			if err != nil {
				continue // Don't set a example value if it's not valid
			}

			param.Example = val
		case schemaExampleTag:
			err = setSchemaExampleV3(param.Schema.Spec, schemaType, attr)
		case extensionsTag:
			param.Schema.Spec.Extensions = setExtensionParam(attr)
		case collectionFormatTag:
			err = setCollectionFormatParamV3(param, attrKey, objectType, attr, comment)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (o *OperationV3) parseParamAttributeForBody(comment, objectType, schemaType string, param *spec.Schema) error {
	schemaType = TransToValidSchemeType(schemaType)

	for attrKey, re := range regexAttributes {
		attr, err := findAttr(re, comment)
		if err != nil {
			continue
		}

		switch attrKey {
		case enumsTag:
			err = setEnumParamV3(param, attr, objectType, schemaType)
		case minimumTag, maximumTag:
			err = setNumberParamV3(param, attrKey, schemaType, attr, comment)
		case defaultTag:
			err = setDefaultV3(param, schemaType, attr)
		case minLengthTag, maxLengthTag:
			err = setStringParamV3(param, attrKey, schemaType, attr, comment)
		case formatTag:
			param.Format = attr
		case exampleTag:
			err = setSchemaExampleV3(param, schemaType, attr)
		case schemaExampleTag:
			err = setSchemaExampleV3(param, schemaType, attr)
		case extensionsTag:
			param.Extensions = setExtensionParam(attr)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func setCollectionFormatParamV3(param *spec.Parameter, name, schemaType, attr, commentLine string) error {
	if schemaType == ARRAY {
		param.Style = TransToValidCollectionFormatV3(attr, param.In)
		return nil
	}

	return fmt.Errorf("%s is attribute to set to an array. comment=%s got=%s", name, commentLine, schemaType)
}

func setSchemaExampleV3(param *spec.Schema, schemaType string, value string) error {
	val, err := defineType(schemaType, value)
	if err != nil {
		return nil // Don't set a example value if it's not valid
	}

	// skip schema
	if param == nil {
		return nil
	}

	switch v := val.(type) {
	case string:
		//  replaces \r \n \t in example string values.
		param.Example = strings.NewReplacer(`\r`, "\r", `\n`, "\n", `\t`, "\t").Replace(v)
	default:
		param.Example = val
	}

	return nil
}

func setExampleParameterV3(param *spec.Parameter, schemaType string, value string) error {
	val, err := defineType(schemaType, value)
	if err != nil {
		return nil // Don't set a example value if it's not valid
	}

	param.Example = val

	return nil
}

func setStringParamV3(param *spec.Schema, name, schemaType, attr, commentLine string) error {
	if schemaType != STRING {
		return fmt.Errorf("%s is attribute to set to a number. comment=%s got=%s", name, commentLine, schemaType)
	}

	n, err := strconv.Atoi(attr)
	if err != nil {
		return fmt.Errorf("%s is allow only a number got=%s", name, attr)
	}

	switch name {
	case minLengthTag:
		param.MinLength = &n
	case maxLengthTag:
		param.MaxLength = &n
	}

	return nil
}

func setDefaultV3(param *spec.Schema, schemaType string, value string) error {
	val, err := defineType(schemaType, value)
	if err != nil {
		return nil // Don't set a default value if it's not valid
	}

	param.Default = val

	return nil
}

func setEnumParamV3(param *spec.Schema, attr, objectType, schemaType string) error {
	for _, e := range strings.Split(attr, ",") {
		e = strings.TrimSpace(e)

		value, err := defineType(schemaType, e)
		if err != nil {
			return err
		}

		switch objectType {
		case ARRAY:
			param.Items.Schema.Spec.Enum = append(param.Items.Schema.Spec.Enum, value)
		default:
			param.Enum = append(param.Enum, value)
		}
	}

	return nil
}

func setNumberParamV3(param *spec.Schema, name, schemaType, attr, commentLine string) error {
	switch schemaType {
	case INTEGER, NUMBER:
		n, err := strconv.Atoi(attr)
		if err != nil {
			return fmt.Errorf("maximum is allow only a number. comment=%s got=%s", commentLine, attr)
		}

		switch name {
		case minimumTag:
			param.Minimum = &n
		case maximumTag:
			param.Maximum = &n
		}

		return nil
	default:
		return fmt.Errorf("%s is attribute to set to a number. comment=%s got=%s", name, commentLine, schemaType)
	}
}

func (o *OperationV3) parseAPIObjectSchema(commentLine, schemaType, refType string, astFile *ast.File) (*spec.RefOrSpec[spec.Schema], error) {
	if strings.HasSuffix(refType, ",") && strings.Contains(refType, "[") {
		// regexp may have broken generic syntax. find closing bracket and add it back
		allMatchesLenOffset := strings.Index(commentLine, refType) + len(refType)
		lostPartEndIdx := strings.Index(commentLine[allMatchesLenOffset:], "]")
		if lostPartEndIdx >= 0 {
			refType += commentLine[allMatchesLenOffset : allMatchesLenOffset+lostPartEndIdx+1]
		}
	}

	switch schemaType {
	case OBJECT:
		if !strings.HasPrefix(refType, "[]") {
			return o.parseObjectSchema(refType, astFile)
		}

		refType = refType[2:]

		fallthrough
	case ARRAY:
		schema, err := o.parseObjectSchema(refType, astFile)
		if err != nil {
			return nil, err
		}

		result := spec.NewSchemaSpec()
		result.Spec.Type = &spec.SingleOrArray[string]{ARRAY}
		result.Spec.Items = spec.NewBoolOrSchema(false, schema) // TODO: allowed?
		return result, nil

	default:
		return PrimitiveSchemaV3(schemaType), nil
	}
}

// ParseRouterComment parses comment for given `router` comment string.
func (o *OperationV3) ParseRouterComment(commentLine string) error {
	matches := routerPattern.FindStringSubmatch(commentLine)
	if len(matches) != 3 {
		return fmt.Errorf("can not parse router comment \"%s\"", commentLine)
	}

	signature := RouteProperties{
		Path:       matches[1],
		HTTPMethod: strings.ToUpper(matches[2]),
	}

	if _, ok := allMethod[signature.HTTPMethod]; !ok {
		return fmt.Errorf("invalid method: %s", signature.HTTPMethod)
	}

	o.RouterProperties = append(o.RouterProperties, signature)

	return nil
}

func (o *OperationV3) ParseServerURLComment(commentLine string) error {
	server := spec.NewServer()
	server.Spec.URL = commentLine
	o.Servers = append(o.Servers, server)
	return nil
}

func (o *OperationV3) ParseServerDescriptionComment(commentLine string) error {
	lastAddedServer := o.Servers[len(o.Servers)-1]
	lastAddedServer.Spec.Description = commentLine
	return nil
}

// createParameter returns swagger spec.Parameter for given  paramType, description, paramName, schemaType, required.
func createParameterV3(in, description, paramName, objectType, schemaType string, required bool, enums []interface{}, collectionFormat string) spec.Parameter {
	// //five possible parameter types. 	query, path, body, header, form
	result := spec.Parameter{
		Description: description,
		Required:    required,
		Name:        paramName,
		In:          in,
		Schema:      spec.NewRefOrSpec(nil, &spec.Schema{}),
	}

	if in == "body" {
		return result
	}

	switch objectType {
	case ARRAY:
		result.Schema.Spec.Type = &spec.SingleOrArray[string]{objectType}
		result.Schema.Spec.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
		result.Schema.Spec.Items.Schema.Spec.Type = &spec.SingleOrArray[string]{schemaType}
		result.Schema.Spec.Enum = enums
	case PRIMITIVE, OBJECT:
		result.Schema.Spec.Type = &spec.SingleOrArray[string]{schemaType}
		result.Schema.Spec.Enum = enums
	}

	return result
}

func (o *OperationV3) parseObjectSchema(refType string, astFile *ast.File) (*spec.RefOrSpec[spec.Schema], error) {
	return parseObjectSchemaV3(o.parser, refType, astFile)
}

func parseObjectSchemaV3(parser *Parser, refType string, astFile *ast.File) (*spec.RefOrSpec[spec.Schema], error) {
	switch {
	case refType == NIL:
		return nil, nil
	case refType == INTERFACE:
		return PrimitiveSchemaV3(OBJECT), nil
	case refType == ANY:
		return PrimitiveSchemaV3(OBJECT), nil
	case IsGolangPrimitiveType(refType):
		refType = TransToValidSchemeType(refType)

		return PrimitiveSchemaV3(refType), nil
	case IsPrimitiveType(refType):
		return PrimitiveSchemaV3(refType), nil
	case strings.HasPrefix(refType, "[]"):
		schema, err := parseObjectSchemaV3(parser, refType[2:], astFile)
		if err != nil {
			return nil, err
		}

		result := spec.NewSchemaSpec()
		result.Spec.Type = &spec.SingleOrArray[string]{ARRAY}
		result.Spec.Items = spec.NewBoolOrSchema(false, schema)

		return result, nil
	case strings.HasPrefix(refType, "map["):
		// ignore key type
		idx := strings.Index(refType, "]")
		if idx < 0 {
			return nil, fmt.Errorf("invalid type: %s", refType)
		}

		refType = refType[idx+1:]
		if refType == INTERFACE || refType == ANY {
			schema := &spec.Schema{}
			schema.AdditionalProperties = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
			schema.Type = &spec.SingleOrArray[string]{OBJECT}
			refOrSpec := spec.NewRefOrSpec(nil, schema)
			return refOrSpec, nil
		}

		schema, err := parseObjectSchemaV3(parser, refType, astFile)
		if err != nil {
			return nil, err
		}

		result := &spec.Schema{}
		result.AdditionalProperties = spec.NewBoolOrSchema(false, schema)
		result.Type = &spec.SingleOrArray[string]{OBJECT}
		refOrSpec := spec.NewSchemaSpec()
		refOrSpec.Spec = result

		return refOrSpec, nil
	case strings.Contains(refType, "{"):
		return parseCombinedObjectSchemaV3(parser, refType, astFile)
	default:
		if parser != nil { // checking refType has existing in 'TypeDefinitions'
			schema, err := parser.getTypeSchemaV3(refType, astFile, true)
			if err != nil {
				return nil, err
			}

			return schema, nil
		}

		return spec.NewSchemaRef(spec.NewRef("#/components/schemas/" + refType)), nil
	}
}

// ParseResponseHeaderComment parses comment for given `response header` comment string.
func (o *OperationV3) ParseResponseHeaderComment(commentLine string, _ *ast.File) error {
	matches := responsePattern.FindStringSubmatch(commentLine)
	if len(matches) != 5 {
		return fmt.Errorf("can not parse response comment \"%s\"", commentLine)
	}

	header := newHeaderSpecV3(strings.Trim(matches[2], "{}"), strings.Trim(matches[4], "\""))

	headerKey := strings.TrimSpace(matches[3])

	if strings.EqualFold(matches[1], "all") {
		if o.Responses.Spec.Default != nil {
			o.Responses.Spec.Default.Spec.Spec.Headers[headerKey] = header
		}

		if o.Responses.Spec.Response != nil {
			for _, v := range o.Responses.Spec.Response {
				v.Spec.Spec.Headers[headerKey] = header

			}
		}

		return nil
	}

	for _, codeStr := range strings.Split(matches[1], ",") {
		if strings.EqualFold(codeStr, defaultTag) {
			if o.Responses.Spec.Default != nil {
				o.Responses.Spec.Default.Spec.Spec.Headers[headerKey] = header
			}

			continue
		}

		_, err := strconv.Atoi(codeStr)
		if err != nil {
			return fmt.Errorf("can not parse response comment \"%s\"", commentLine)
		}

		// TODO check condition
		if o.Responses != nil && o.Responses.Spec != nil && o.Responses.Spec.Response != nil {
			response, responseExist := o.Responses.Spec.Response[codeStr]
			if responseExist {
				response.Spec.Spec.Headers[headerKey] = header
				o.Responses.Spec.Response[codeStr] = response
			}
		}
	}

	return nil
}

func newHeaderSpecV3(schemaType, description string) *spec.RefOrSpec[spec.Extendable[spec.Header]] {
	result := spec.NewHeaderSpec()
	result.Spec.Spec.Description = description
	result.Spec.Spec.Schema = spec.NewSchemaSpec()
	result.Spec.Spec.Schema.Spec.Type = &spec.SingleOrArray[string]{schemaType}

	return result
}

// ParseResponseComment parses comment for given `response` comment string.
func (o *OperationV3) ParseResponseComment(commentLine string, astFile *ast.File) error {
	matches := responsePattern.FindStringSubmatch(commentLine)
	if len(matches) != 5 {
		err := o.ParseEmptyResponseComment(commentLine)
		if err != nil {
			return o.ParseEmptyResponseOnly(commentLine)
		}

		return err
	}

	description := strings.Trim(matches[4], "\"")

	schema, err := o.parseAPIObjectSchema(commentLine, strings.Trim(matches[2], "{}"), strings.TrimSpace(matches[3]), astFile)
	if err != nil {
		return err
	}

	for _, codeStr := range strings.Split(matches[1], ",") {
		if strings.EqualFold(codeStr, defaultTag) {
			codeStr = ""
		} else {
			code, err := strconv.Atoi(codeStr)
			if err != nil {
				return fmt.Errorf("can not parse response comment \"%s\"", commentLine)
			}
			if description == "" {
				description = http.StatusText(code)
			}
		}

		response := spec.NewResponseSpec()
		response.Spec.Spec.Description = description

		mimeType := "application/json" // TODO: set correct mimeType
		setResponseSchema(response.Spec.Spec, mimeType, schema)

		o.AddResponse(codeStr, response)
	}

	return nil
}

// setResponseSchema sets response schema for given response.
func setResponseSchema(response *spec.Response, mimeType string, schema *spec.RefOrSpec[spec.Schema]) {
	mediaType := spec.NewMediaType()
	mediaType.Spec.Schema = schema

	if response.Content == nil {
		response.Content = make(map[string]*spec.Extendable[spec.MediaType])
	}

	response.Content[mimeType] = mediaType
}

// ParseEmptyResponseComment parse only comment out status code and description,eg: @Success 200 "it's ok".
func (o *OperationV3) ParseEmptyResponseComment(commentLine string) error {
	matches := emptyResponsePattern.FindStringSubmatch(commentLine)
	if len(matches) != 3 {
		return fmt.Errorf("can not parse response comment \"%s\"", commentLine)
	}

	description := strings.Trim(matches[2], "\"")

	for _, codeStr := range strings.Split(matches[1], ",") {
		if strings.EqualFold(codeStr, defaultTag) {
			codeStr = ""
		} else {
			_, err := strconv.Atoi(codeStr)
			if err != nil {
				return fmt.Errorf("can not parse response comment \"%s\"", commentLine)
			}
		}

		o.AddResponse(codeStr, newResponseWithDescription(description))
	}

	return nil
}

// AddResponse add a response for a code.
// If the code is already exist, it will merge with the old one:
// 1. The description will be replaced by the new one if the new one is not empty.
// 2. The content schema will be merged using `oneOf` if the new one is not empty.
func (o *OperationV3) AddResponse(code string, response *spec.RefOrSpec[spec.Extendable[spec.Response]]) {
	if response.Spec.Spec.Headers == nil {
		response.Spec.Spec.Headers = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Header]])
	}

	if o.Responses.Spec.Response == nil {
		o.Responses.Spec.Response = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Response]])
	}

	res := response
	var prev *spec.RefOrSpec[spec.Extendable[spec.Response]]
	if code != "" {
		prev = o.Responses.Spec.Response[code]
	} else {
		prev = o.Responses.Spec.Default
	}
	if prev != nil { // merge into prev
		res = prev
		if response.Spec.Spec.Description != "" {
			prev.Spec.Spec.Description = response.Spec.Spec.Description
		}
		if len(response.Spec.Spec.Content) > 0 {
			// responses should only have one content type
			singleKey := ""
			for k := range response.Spec.Spec.Content {
				singleKey = k
				break
			}
			if prevMediaType := prev.Spec.Spec.Content[singleKey]; prevMediaType == nil {
				prev.Spec.Spec.Content = response.Spec.Spec.Content
			} else {
				newMediaType := response.Spec.Spec.Content[singleKey]
				if len(newMediaType.Extensions) > 0 {
					if prevMediaType.Extensions == nil {
						prevMediaType.Extensions = make(map[string]interface{})
					}
					for k, v := range newMediaType.Extensions {
						prevMediaType.Extensions[k] = v
					}
				}
				if len(newMediaType.Spec.Examples) > 0 {
					if prevMediaType.Spec.Examples == nil {
						prevMediaType.Spec.Examples = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Example]])
					}
					for k, v := range newMediaType.Spec.Examples {
						prevMediaType.Spec.Examples[k] = v
					}
				}
				if prevSchema := prevMediaType.Spec.Schema; prevSchema.Ref != nil || prevSchema.Spec.OneOf == nil {
					oneOfSchema := spec.NewSchemaSpec()
					oneOfSchema.Spec.OneOf = []*spec.RefOrSpec[spec.Schema]{prevSchema, newMediaType.Spec.Schema}
					prevMediaType.Spec.Schema = oneOfSchema
				} else {
					prevSchema.Spec.OneOf = append(prevSchema.Spec.OneOf, newMediaType.Spec.Schema)
				}
			}
		}
	}

	if code != "" {
		o.Responses.Spec.Response[code] = res
	} else {
		o.Responses.Spec.Default = res
	}
}

// ParseEmptyResponseOnly parse only comment out status code ,eg: @Success 200.
func (o *OperationV3) ParseEmptyResponseOnly(commentLine string) error {
	for _, codeStr := range strings.Split(commentLine, ",") {
		var description string
		if strings.EqualFold(codeStr, defaultTag) {
			codeStr = ""
		} else {
			code, err := strconv.Atoi(codeStr)
			if err != nil {
				return fmt.Errorf("can not parse response comment \"%s\"", commentLine)
			}
			description = http.StatusText(code)
		}

		o.AddResponse(codeStr, newResponseWithDescription(description))
	}

	return nil
}

func newResponseWithDescription(description string) *spec.RefOrSpec[spec.Extendable[spec.Response]] {
	response := spec.NewResponseSpec()
	response.Spec.Spec.Description = description
	return response
}

func parseCombinedObjectSchemaV3(parser *Parser, refType string, astFile *ast.File) (*spec.RefOrSpec[spec.Schema], error) {
	matches := combinedPattern.FindStringSubmatch(refType)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid type: %s", refType)
	}

	schema, err := parseObjectSchemaV3(parser, matches[1], astFile)
	if err != nil {
		return nil, err
	}

	fields, props := parseFields(matches[2]), map[string]*spec.RefOrSpec[spec.Schema]{}

	for _, field := range fields {
		keyVal := strings.SplitN(field, "=", 2)
		if len(keyVal) != 2 {
			continue
		}

		schema, err := parseObjectSchemaV3(parser, keyVal[1], astFile)
		if err != nil {
			return nil, err
		}

		props[keyVal[0]] = schema
	}

	if len(props) == 0 {
		return schema, nil
	}

	if schema.Ref == nil &&
		len(*schema.Spec.Type) > 0 &&
		(*schema.Spec.Type)[0] == OBJECT &&
		len(schema.Spec.Properties) == 0 &&
		schema.Spec.AdditionalProperties == nil {
		schema.Spec.Properties = props
		return schema, nil
	}

	schemaRefPath := strings.Replace(schema.Ref.Ref, "#/components/schemas/", "", 1)
	schemaSpec := parser.openAPI.Components.Spec.Schemas[schemaRefPath]
	schemaSpec.Spec.JsonSchemaComposition.AllOf = make([]*spec.RefOrSpec[spec.Schema], len(props))

	i := 0
	for name, prop := range props {
		wrapperSpec := spec.NewSchemaSpec()
		wrapperSpec.Spec = &spec.Schema{}
		wrapperSpec.Spec.Type = &spec.SingleOrArray[string]{OBJECT}
		wrapperSpec.Spec.Properties = map[string]*spec.RefOrSpec[spec.Schema]{
			name: prop,
		}

		parser.openAPI.Components.Spec.Schemas[name] = wrapperSpec

		ref := spec.NewRefOrSpec[spec.Schema](spec.NewRef("#/components/schemas/"+name), nil)

		schemaSpec.Spec.JsonSchemaComposition.AllOf[i] = ref
		i++
	}

	return schemaSpec, nil
}

// ParseSecurityComment parses comment for given `security` comment string.
func (o *OperationV3) ParseSecurityComment(commentLine string) error {
	var (
		securityMap    = make(map[string][]string)
		securitySource = commentLine[strings.Index(commentLine, "@Security")+1:]
	)

	for _, securityOption := range strings.Split(securitySource, "||") {
		securityOption = strings.TrimSpace(securityOption)

		left, right := strings.Index(securityOption, "["), strings.Index(securityOption, "]")

		if !(left == -1 && right == -1) {
			scopes := securityOption[left+1 : right]

			var options []string

			for _, scope := range strings.Split(scopes, ",") {
				options = append(options, strings.TrimSpace(scope))
			}

			securityKey := securityOption[0:left]
			securityMap[securityKey] = append(securityMap[securityKey], options...)
		} else {
			securityKey := strings.TrimSpace(securityOption)
			securityMap[securityKey] = []string{}
		}
	}

	o.Security = append(o.Security, securityMap)

	return nil
}

// ParseCodeSample godoc.
func (o *OperationV3) ParseCodeSample(attribute, _, lineRemainder string) error {
	log.Println("line remainder:", lineRemainder)

	if lineRemainder == "file" {
		log.Println("line remainder is file")

		data, isJSON, err := getCodeExampleForSummary(o.Summary, o.codeExampleFilesDir)
		if err != nil {
			return err
		}

		// using custom type, as json marshaller has problems with []map[interface{}]map[interface{}]interface{}
		var valueJSON CodeSamples

		if isJSON {
			err = json.Unmarshal(data, &valueJSON)
			if err != nil {
				return fmt.Errorf("annotation %s need a valid json value. error: %s", attribute, err.Error())
			}
		} else {
			err = yaml.Unmarshal(data, &valueJSON)
			if err != nil {
				return fmt.Errorf("annotation %s need a valid yaml value. error: %s", attribute, err.Error())
			}
		}

		o.Responses.Extensions[attribute[1:]] = valueJSON

		return nil
	}

	// Fallback into existing logic
	return o.ParseMetadata(attribute, strings.ToLower(attribute), lineRemainder)
}
