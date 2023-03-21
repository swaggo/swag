package swag

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"strings"

	"github.com/sv-tools/openapi/spec"
)

// Operation describes a single API operation on a path.
// For more information: https://github.com/Nerzal/swag#api-operation
type OperationV3 struct {
	parser              *Parser
	codeExampleFilesDir string
	spec.Operation
	RouterProperties []RouteProperties
}

// NewOperationV3 returns a new instance of OperationV3.
func NewOperationV3(parser *Parser, options ...func(*OperationV3)) *OperationV3 {
	operation := &OperationV3{
		parser:    parser,
		Operation: *spec.NewOperation().Spec,
	}

	for _, option := range options {
		option(operation)
	}

	return operation
}

// SetCodeExampleFilesDirectory sets the directory to search for codeExamples.
func SetCodeExampleFilesDirectoryV3(directoryPath string) func(*OperationV3) {
	return func(o *OperationV3) {
		o.codeExampleFilesDir = directoryPath
	}
}

// ParseComment parses comment for given comment string and returns error if error occurs.
func (o *OperationV3) ParseCommentV3(comment string, astFile *ast.File) error {
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
		// return o.ParseResponseComment(lineRemainder, astFile)
	case headerAttr:
		// return o.ParseResponseHeaderComment(lineRemainder, astFile)
	case routerAttr:
		return o.ParseRouterComment(lineRemainder)
	case securityAttr:
		// return o.ParseSecurityComment(lineRemainder)
	case deprecatedAttr:
		o.Deprecated = true
	case xCodeSamplesAttr:
		// return o.ParseCodeSample(attribute, commentLine, lineRemainder)
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

		var valueJSON interface{}

		err := json.Unmarshal([]byte(lineRemainder), &valueJSON)
		if err != nil {
			return fmt.Errorf("annotation %s need a valid json value. error: %s", attribute, err.Error())
		}

		// don't use the method provided by spec lib, because it will call toLower() on attribute names, which is wrongly
		// o.Extensions[attribute[1:]] = valueJSON
		// TODO: vendor extensions must be placed under the http method. not sure how to get that at this place
		// return errors.New("not implemented yet")
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

	// TODO this must be moved into another comment
	// return parseMimeTypeList(commentLine, &o.RequestBody.Spec.Spec.Content, )
	// result, err := parseMimeTypeListV3(commentLine, "%v accept type can't be accepted")
	// if err != nil {
	// 	return errors.Wrap(err, errMessage)
	// }

	// for _, value := range result {
	// 	o.RequestBody.Spec.Spec.Content[value] = spec.NewMediaType()
	// }

	return nil
}

// ParseProduceComment parses comment for given `produce` comment string.
func (o *OperationV3) ParseProduceComment(commentLine string) error {
	const errMessage = "could not parse produce comment"
	// return parseMimeTypeList(commentLine, &o.Responses, "%v produce type can't be accepted")

	// result, err := parseMimeTypeListV3(commentLine, "%v accept type can't be accepted")
	// if err != nil {
	// 	return errors.Wrap(err, errMessage)
	// }

	// for _, value := range result {
	// 	o.Responses.Spec.Response
	// }

	// TODO the format of the comment needs to be changed in order to work
	// The produce can be different per response code, so the produce mimetype needs to be included in the response comment

	return nil
}

// parseMimeTypeList parses a list of MIME Types for a comment like
// `produce` (`Content-Type:` response header) or
// `accept` (`Accept:` request header).
func parseMimeTypeListV3(mimeTypeList string, format string) ([]string, error) {
	var result []string
	for _, typeName := range strings.Split(mimeTypeList, ",") {
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
		schema, _ := o.parser.getTypeSchema(refType, astFile, false)
		if schema != nil && len(schema.Type) == 1 && schema.Enum != nil {
			if objectType == OBJECT {
				objectType = PRIMITIVE
			}
			refType = TransToValidSchemeType(schema.Type[0])
			enums = schema.Enum
		}
	}

	requiredText := strings.ToLower(matches[4])
	required := requiredText == "true" || requiredText == requiredLabel
	description := matches[5]

	param := createParameter(paramType, description, name, objectType, refType, required, enums, o.parser.collectionFormatInQuery)

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
	case "query", "formData":
		switch objectType {
		case ARRAY:
			if !IsPrimitiveType(refType) && !(refType == "file" && paramType == "formData") {
				return fmt.Errorf("%s is not supported array type for %s", refType, paramType)
			}
		case PRIMITIVE:
			break
		case OBJECT:
			schema, err := o.parser.getTypeSchema(refType, astFile, false)
			if err != nil {
				return err
			}

			if len(schema.Properties) == 0 {
				return nil
			}

			items := schema.Properties.ToOrderedSchemaItems()

			for _, item := range items {
				name, prop := item.Name, item.Schema
				if len(prop.Type) == 0 {
					continue
				}

				switch {
				case prop.Type[0] == ARRAY && prop.Items.Schema != nil &&
					len(prop.Items.Schema.Type) > 0 && IsSimplePrimitiveType(prop.Items.Schema.Type[0]):

					param = createParameter(paramType, prop.Description, name, prop.Type[0], prop.Items.Schema.Type[0], findInSlice(schema.Required, name), enums, o.parser.collectionFormatInQuery)

				case IsSimplePrimitiveType(prop.Type[0]):
					param = createParameter(paramType, prop.Description, name, PRIMITIVE, prop.Type[0], findInSlice(schema.Required, name), enums, o.parser.collectionFormatInQuery)
				default:
					o.parser.debug.Printf("skip field [%s] in %s is not supported type for %s", name, refType, paramType)

					continue
				}

				param.Nullable = prop.Nullable
				param.Format = prop.Format
				param.Default = prop.Default
				param.Example = prop.Example
				param.Extensions = prop.Extensions
				param.CommonValidations.Maximum = prop.Maximum
				param.CommonValidations.Minimum = prop.Minimum
				param.CommonValidations.ExclusiveMaximum = prop.ExclusiveMaximum
				param.CommonValidations.ExclusiveMinimum = prop.ExclusiveMinimum
				param.CommonValidations.MaxLength = prop.MaxLength
				param.CommonValidations.MinLength = prop.MinLength
				param.CommonValidations.Pattern = prop.Pattern
				param.CommonValidations.MaxItems = prop.MaxItems
				param.CommonValidations.MinItems = prop.MinItems
				param.CommonValidations.UniqueItems = prop.UniqueItems
				param.CommonValidations.MultipleOf = prop.MultipleOf
				param.CommonValidations.Enum = prop.Enum
				// o.Operation.Parameters = append(o.Operation.Parameters, param)
			}

			return nil
		}
	case "body":
		if objectType == PRIMITIVE {
			param.Schema = PrimitiveSchema(refType)
		} else {
			// schema, err := o.parseAPIObjectSchema(commentLine, objectType, refType, astFile)
			// if err != nil {
			// 	return err
			// }

			// param.Schema = schema
		}
	default:
		return fmt.Errorf("%s is not supported paramType", paramType)
	}

	// err := o.parseParamAttribute(commentLine, objectType, refType, &param)
	// if err != nil {
	// 	return err
	// }

	// o.Operation.Parameters = append(o.Operation.Parameters, param)

	return nil
}

func (o *OperationV3) parseParamAttribute(comment, objectType, schemaType string, param *spec.Parameter) error {
	schemaType = TransToValidSchemeType(schemaType)

	// for attrKey, re := range regexAttributes {
	// 	attr, err := findAttr(re, comment)
	// 	if err != nil {
	// 		continue
	// 	}

	// 	switch attrKey {
	// 	case enumsTag:
	// 		err = setEnumParam(param, attr, objectType, schemaType)
	// 	case minimumTag, maximumTag:
	// 		err = setNumberParam(param, attrKey, schemaType, attr, comment)
	// 	case defaultTag:
	// 		err = setDefault(param, schemaType, attr)
	// 	case minLengthTag, maxLengthTag:
	// 		err = setStringParam(param, attrKey, schemaType, attr, comment)
	// 	case formatTag:
	// 		param.Format = attr
	// 	case exampleTag:
	// 		err = setExample(param, schemaType, attr)
	// 	case schemaExampleTag:
	// 		err = setSchemaExample(param, schemaType, attr)
	// 	case extensionsTag:
	// 		param.Extensions = setExtensionParam(attr)
	// 	case collectionFormatTag:
	// 		err = setCollectionFormatParam(param, attrKey, objectType, attr, comment)
	// 	}

	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (o *OperationV3) parseAPIObjectSchema(commentLine, schemaType, refType string, astFile *ast.File) (*spec.Schema, error) {
	// if strings.HasSuffix(refType, ",") && strings.Contains(refType, "[") {
	// 	// regexp may have broken generic syntax. find closing bracket and add it back
	// 	allMatchesLenOffset := strings.Index(commentLine, refType) + len(refType)
	// 	lostPartEndIdx := strings.Index(commentLine[allMatchesLenOffset:], "]")
	// 	if lostPartEndIdx >= 0 {
	// 		refType += commentLine[allMatchesLenOffset : allMatchesLenOffset+lostPartEndIdx+1]
	// 	}
	// }

	// switch schemaType {
	// case OBJECT:
	// 	if !strings.HasPrefix(refType, "[]") {
	// 		return o.parseObjectSchema(refType, astFile)
	// 	}

	// 	refType = refType[2:]

	// 	fallthrough
	// case ARRAY:
	// 	schema, err := o.parseObjectSchema(refType, astFile)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return spec.ArrayProperty(schema), nil
	// default:
	// 	return PrimitiveSchema(schemaType), nil
	// }
	return nil, nil
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
