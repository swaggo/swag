package swag

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/sv-tools/openapi/spec"
)

// Operation describes a single API operation on a path.
// For more information: https://github.com/swaggo/swag#api-operation
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
	case "query", "formData":
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

			if len(schema.Properties) == 0 {
				return nil
			}

			for name, item := range schema.Properties {
				prop := item.Spec
				if len(prop.Type) == 0 {
					continue
				}

				switch {
				case prop.Type[0] == ARRAY &&
					prop.Items.Schema != nil &&
					len(prop.Items.Schema.Spec.Type) > 0 &&
					IsSimplePrimitiveType(prop.Items.Schema.Spec.Type[0]):

					param = createParameterV3(paramType, prop.Description, name, prop.Type[0], prop.Items.Schema.Spec.Type[0], findInSlice(schema.Required, name), enums, o.parser.collectionFormatInQuery)

				case IsSimplePrimitiveType(prop.Type[0]):
					param = createParameterV3(paramType, prop.Description, name, PRIMITIVE, prop.Type[0], findInSlice(schema.Required, name), enums, o.parser.collectionFormatInQuery)
				default:
					o.parser.debug.Printf("skip field [%s] in %s is not supported type for %s", name, refType, paramType)

					continue
				}

				param.Schema.Spec = prop

				listItem := &spec.RefOrSpec[spec.Extendable[spec.Parameter]]{
					Spec: &spec.Extendable[spec.Parameter]{
						Spec: &param,
					},
				}

				o.Operation.Parameters = append(o.Operation.Parameters, listItem)
			}

			return nil
		}
	case "body":
		if objectType == PRIMITIVE {
			param.Schema = &spec.RefOrSpec[spec.Schema]{
				Spec: PrimitiveSchemaV3(refType),
			}
		} else {
			schema, err := o.parseAPIObjectSchema(commentLine, objectType, refType, astFile)
			if err != nil {
				return err
			}

			param.Schema = &spec.RefOrSpec[spec.Schema]{
				Spec: schema,
			}
		}
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

func (o *OperationV3) parseParamAttribute(comment, objectType, schemaType string, param *spec.Parameter) error {
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
			param.Schema.Spec.Format = attr
		case exampleTag:
			err = setExampleV3(param, schemaType, attr)
		case schemaExampleTag:
			err = setSchemaExampleV3(param, schemaType, attr)
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

func setCollectionFormatParamV3(param *spec.Parameter, name, schemaType, attr, commentLine string) error {
	if schemaType == ARRAY {
		// param.Schema.Spec.JsonSchema.
		// param.Schema.Spec.CollectionFormat = TransToValidCollectionFormat(attr)
		// TODO ich hab kein plan bruder
		return nil
	}

	return fmt.Errorf("%s is attribute to set to an array. comment=%s got=%s", name, commentLine, schemaType)
}

func setSchemaExampleV3(param *spec.Parameter, schemaType string, value string) error {
	val, err := defineType(schemaType, value)
	if err != nil {
		return nil // Don't set a example value if it's not valid
	}
	// skip schema
	if param.Schema == nil {
		return nil
	}

	switch v := val.(type) {
	case string:
		//  replaces \r \n \t in example string values.
		param.Schema.Spec.Example = strings.NewReplacer(`\r`, "\r", `\n`, "\n", `\t`, "\t").Replace(v)
	default:
		param.Schema.Spec.Example = val
	}

	return nil
}

func setExampleV3(param *spec.Parameter, schemaType string, value string) error {
	val, err := defineType(schemaType, value)
	if err != nil {
		return nil // Don't set a example value if it's not valid
	}

	param.Example = val

	return nil
}

func setStringParamV3(param *spec.Parameter, name, schemaType, attr, commentLine string) error {
	if schemaType != STRING {
		return fmt.Errorf("%s is attribute to set to a number. comment=%s got=%s", name, commentLine, schemaType)
	}

	n, err := strconv.Atoi(attr)
	if err != nil {
		return fmt.Errorf("%s is allow only a number got=%s", name, attr)
	}

	switch name {
	case minLengthTag:
		param.Schema.Spec.MinLength = &n
	case maxLengthTag:
		param.Schema.Spec.MaxLength = &n
	}

	return nil
}

func setDefaultV3(param *spec.Parameter, schemaType string, value string) error {
	val, err := defineType(schemaType, value)
	if err != nil {
		return nil // Don't set a default value if it's not valid
	}

	param.Schema.Spec.Default = val

	return nil
}

func setEnumParamV3(param *spec.Parameter, attr, objectType, schemaType string) error {
	for _, e := range strings.Split(attr, ",") {
		e = strings.TrimSpace(e)

		value, err := defineType(schemaType, e)
		if err != nil {
			return err
		}

		switch objectType {
		case ARRAY:
			param.Schema.Spec.Items.Schema.Spec.Enum = append(param.Schema.Spec.Items.Schema.Spec.Enum, value)
		default:
			param.Schema.Spec.Enum = append(param.Schema.Spec.Enum, value)
		}
	}

	return nil
}

func setNumberParamV3(param *spec.Parameter, name, schemaType, attr, commentLine string) error {
	switch schemaType {
	case INTEGER, NUMBER:
		n, err := strconv.Atoi(attr)
		if err != nil {
			return fmt.Errorf("maximum is allow only a number. comment=%s got=%s", commentLine, attr)
		}

		switch name {
		case minimumTag:
			param.Schema.Spec.Minimum = &n
		case maximumTag:
			param.Schema.Spec.Maximum = &n
		}

		return nil
	default:
		return fmt.Errorf("%s is attribute to set to a number. comment=%s got=%s", name, commentLine, schemaType)
	}
}

func (o *OperationV3) parseAPIObjectSchema(commentLine, schemaType, refType string, astFile *ast.File) (*spec.Schema, error) {
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
		// schema, err := o.parseObjectSchema(refType, astFile)
		// if err != nil {
		// 	return nil, err
		// }

		// return spec.ArrayProperty(schema), nil
	default:
		return PrimitiveSchemaV3(schemaType), nil
	}

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
		// result.Schema.Spec.Type = objectType
		// result.Schema.Spec.CollectionFormat = collectionFormat
		// result.Schema.Spec.Items = spec.NewBoolOrSchema(true, spec.NewRefOrSpec(nil, spec *spec.T))

		// &spec.Items{
		// 	CommonValidations: spec.CommonValidations{
		// 		Enum: enums,
		// 	},
		// 	SimpleSchema: spec.SimpleSchema{
		// 		Type: schemaType,
		// 	},
		// }
	case PRIMITIVE, OBJECT:
		result.Schema.Spec.Type = spec.NewSingleOrArray(schemaType)
		result.Schema.Spec.Enum = enums
	}

	return result
}

func (o *OperationV3) parseObjectSchema(refType string, astFile *ast.File) (*spec.Schema, error) {
	return parseObjectSchemaV3(o.parser, refType, astFile)
}

func parseObjectSchemaV3(parser *Parser, refType string, astFile *ast.File) (*spec.Schema, error) {
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
		// schema, err := parseObjectSchema(parser, refType[2:], astFile)
		// if err != nil {
		// 	return nil, err
		// }

		// return spec.ArrayProperty(schema), nil
	case strings.HasPrefix(refType, "map["):
		// // ignore key type
		// idx := strings.Index(refType, "]")
		// if idx < 0 {
		// 	return nil, fmt.Errorf("invalid type: %s", refType)
		// }

		// refType = refType[idx+1:]
		// if refType == INTERFACE || refType == ANY {
		// 	return spec.MapProperty(nil), nil
		// }

		// schema, err := parseObjectSchema(parser, refType, astFile)
		// if err != nil {
		// 	return nil, err
		// }

		// return spec.MapProperty(schema), nil
	case strings.Contains(refType, "{"):
		// return parseCombinedObjectSchema(parser, refType, astFile)
	default:
		if parser != nil { // checking refType has existing in 'TypeDefinitions'
			schema, err := parser.getTypeSchemaV3(refType, astFile, true)
			if err != nil {
				return nil, err
			}

			return schema, nil
		}

		// return spec.NewRef("#/definitions/" + refType), nil

		// return RefSchema(refType), nil
	}

	return nil, nil
}
