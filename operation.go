package swag

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
)

// Operation describes a single API operation on a path.
// For more information: https://github.com/swaggo/swag#api-operation
type Operation struct {
	HTTPMethod string
	Path       string
	spec.Operation

	parser *Parser // TODO: we don't need it
}

// NewOperation creates a new Operation with default properties.
// map[int]Response
func NewOperation() *Operation {
	return &Operation{
		HTTPMethod: "get",
		Operation: spec.Operation{
			OperationProps: spec.OperationProps{},
		},
	}
}

// ParseComment parses comment for gived comment string and returns error if error occurs.
func (operation *Operation) ParseComment(comment string) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "//"))
	if len(commentLine) == 0 {
		return nil
	}

	attribute := strings.Fields(commentLine)[0]
	switch strings.ToLower(attribute) {
	case "@description":
		operation.Description = strings.TrimSpace(commentLine[len(attribute):])
	case "@summary":
		operation.Summary = strings.TrimSpace(commentLine[len(attribute):])
	case "@id":
		operation.ID = strings.TrimSpace(commentLine[len(attribute):])
	case "@tags":
		operation.ParseTagsComment(strings.TrimSpace(commentLine[len(attribute):]))
	case "@accept":
		if err := operation.ParseAcceptComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	case "@produce":
		if err := operation.ParseProduceComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	case "@param":
		if err := operation.ParseParamComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	case "@success", "@failure":
		if err := operation.ParseResponseComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {

			if errWhenEmpty := operation.ParseEmptyResponseComment(strings.TrimSpace(commentLine[len(attribute):])); errWhenEmpty != nil {
				var errs []string
				errs = append(errs, err.Error())
				errs = append(errs, errWhenEmpty.Error())

				return fmt.Errorf(strings.Join(errs, "\n"))

			}
		}

	case "@router":
		if err := operation.ParseRouterComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	case "@security":
		if err := operation.ParseSecurityComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	}
	return nil
}

// ParseParamComment Parse params return []string of param properties
// @Param	queryText		form	      string	  true		        "The email for login"
// 			[param name]    [paramType] [data type]  [is mandatory?]   [Comment]
// @Param   some_id     path    int     true        "Some ID"
func (operation *Operation) ParseParamComment(commentLine string) error {
	re := regexp.MustCompile(`([-\w]+)[\s]+([\w]+)[\s]+([\S.]+)[\s]+([\w]+)[\s]+"([^"]+)"`)
	matches := re.FindStringSubmatch(commentLine)
	if len(matches) != 6 {
		return fmt.Errorf("can not parse param comment \"%s\"", commentLine)
	}
	name := matches[1]
	paramType := matches[2]

	schemaType := matches[3]

	requiredText := strings.ToLower(matches[4])
	required := requiredText == "true" || requiredText == "required"
	description := matches[5]

	var param spec.Parameter

	//five possible parameter types.
	switch paramType {
	case "query", "path", "header":
		param = createParameter(paramType, description, name, TransToValidSchemeType(schemaType), required)
	case "body":
		param = createParameter(paramType, description, name, "object", required) // TODO: if Parameter types can be objects, but also primitives and arrays
		// TODO: this snippets have to extract out
		refSplit := strings.Split(schemaType, ".")
		if len(refSplit) == 2 {
			pkgName := refSplit[0]
			typeName := refSplit[1]
			if typeSpec, ok := operation.parser.TypeDefinitions[pkgName][typeName]; ok {
				operation.parser.registerTypes[schemaType] = typeSpec
			} else {
				return fmt.Errorf("can not find ref type:\"%s\"", schemaType)
			}
			param.Schema.Ref = spec.Ref{
				Ref: jsonreference.MustCreateRef("#/definitions/" + schemaType),
			}
		}
	case "formData":
		param = createParameter(paramType, description, name, TransToValidSchemeType(schemaType), required)
	}
	param = operation.parseAndExtractionParamAttribute(commentLine, schemaType, param)
	operation.Operation.Parameters = append(operation.Operation.Parameters, param)
	return nil
}

var regexAttributes = map[string]*regexp.Regexp{
	// for Enums(A, B)
	"enums": regexp.MustCompile(`(?i)enums\(.*\)`),
	// for Minimum(0)
	"maxinum": regexp.MustCompile(`(?i)maxinum\(.*\)`),
	// for Maximum(0)
	"mininum": regexp.MustCompile(`(?i)mininum\(.*\)`),
	// for Maximum(0)
	"default": regexp.MustCompile(`(?i)default\(.*\)`),
	// for minlength(0)
	"minlength": regexp.MustCompile(`(?i)minlength\(.*\)`),
	// for maxlength(0)
	"maxlength": regexp.MustCompile(`(?i)maxlength\(.*\)`),
	// for format(email)
	"format": regexp.MustCompile(`(?i)format\(.*\)`),
}

func (operation *Operation) parseAndExtractionParamAttribute(commentLine, schemaType string, param spec.Parameter) spec.Parameter {
	schemaType = TransToValidSchemeType(schemaType)
	for attrKey, re := range regexAttributes {
		switch attrKey {
		case "enums":
			attr := re.FindString(commentLine)
			l := strings.Index(attr, "(")
			r := strings.Index(attr, ")")
			if !(l == -1 && r == -1) {
				enums := strings.Split(attr[l+1:r], ",")
				for _, e := range enums {
					e = strings.TrimSpace(e)
					param.Enum = append(param.Enum, defineType(schemaType, e))
				}
			}
		case "maxinum":
			attr := re.FindString(commentLine)
			l := strings.Index(attr, "(")
			r := strings.Index(attr, ")")
			if !(l == -1 && r == -1) {
				if schemaType != "integer" && schemaType != "number" {
					log.Panicf("maxinum is attribute to set to a number. comment=%s got=%s", commentLine, schemaType)
				}
				attr = strings.TrimSpace(attr[l+1 : r])
				n, err := strconv.ParseFloat(attr, 64)
				if err != nil {
					log.Panicf("maximum is allow only a number. comment=%s got=%s", commentLine, attr)
				}
				param.Maximum = &n
			}
		case "mininum":
			attr := re.FindString(commentLine)
			l := strings.Index(attr, "(")
			r := strings.Index(attr, ")")
			if !(l == -1 && r == -1) {
				if schemaType != "integer" && schemaType != "number" {
					log.Panicf("mininum is attribute to set to a number. comment=%s got=%s", commentLine, schemaType)
				}
				attr = strings.TrimSpace(attr[l+1 : r])
				n, err := strconv.ParseFloat(attr, 64)
				if err != nil {
					log.Panicf("mininum is allow only a number got=%s", attr)
				}
				param.Minimum = &n
			}
		case "default":
			attr := re.FindString(commentLine)
			l := strings.Index(attr, "(")
			r := strings.Index(attr, ")")
			if !(l == -1 && r == -1) {
				attr = strings.TrimSpace(attr[l+1 : r])
				param.Default = defineType(schemaType, attr)
			}
		case "maxlength":
			attr := re.FindString(commentLine)
			l := strings.Index(attr, "(")
			r := strings.Index(attr, ")")
			if !(l == -1 && r == -1) {
				if schemaType != "string" {
					log.Panicf("maxlength is attribute to set to a number. comment=%s got=%s", commentLine, schemaType)
				}
				attr = strings.TrimSpace(attr[l+1 : r])
				n, err := strconv.ParseInt(attr, 10, 64)
				if err != nil {
					log.Panicf("maxlength is allow only a number got=%s", attr)
				}
				param.MaxLength = &n
			}
		case "minlength":
			attr := re.FindString(commentLine)
			l := strings.Index(attr, "(")
			r := strings.Index(attr, ")")
			if !(l == -1 && r == -1) {
				if schemaType != "string" {
					log.Panicf("maxlength is attribute to set to a number. comment=%s got=%s", commentLine, schemaType)
				}
				attr = strings.TrimSpace(attr[l+1 : r])
				n, err := strconv.ParseInt(attr, 10, 64)
				if err != nil {
					log.Panicf("minlength is allow only a number got=%s", attr)
				}
				param.MinLength = &n
			}
		case "format":
			attr := re.FindString(commentLine)
			l := strings.Index(attr, "(")
			r := strings.Index(attr, ")")
			if !(l == -1 && r == -1) {
				param.Format = strings.TrimSpace(attr[l+1 : r])
			}
		}
	}
	return param
}

// defineType enum value define the type (object and array unsupported)
func defineType(schemaType string, value string) interface{} {
	schemaType = TransToValidSchemeType(schemaType)
	switch schemaType {
	case "string":
		return value
	case "number":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err))
		}
		return v
	case "integer":
		v, err := strconv.Atoi(value)
		if err != nil {
			panic(fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err))
		}
		return v
	case "boolean":
		v, err := strconv.ParseBool(value)
		if err != nil {
			panic(fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err))
		}
		return v
	default:
		panic(fmt.Errorf("%s is unsupported type in enum value", schemaType))
	}
}

// ParseTagsComment parses comment for gived `tag` comment string.
func (operation *Operation) ParseTagsComment(commentLine string) {
	tags := strings.Split(commentLine, ",")
	for _, tag := range tags {
		operation.Tags = append(operation.Tags, strings.TrimSpace(tag))
	}
}

// ParseAcceptComment parses comment for gived `accept` comment string.
func (operation *Operation) ParseAcceptComment(commentLine string) error {
	accepts := strings.Split(commentLine, ",")
	for _, a := range accepts {
		switch a {
		case "json", "application/json":
			operation.Consumes = append(operation.Consumes, "application/json")
		case "xml", "text/xml":
			operation.Consumes = append(operation.Consumes, "text/xml")
		case "plain", "text/plain":
			operation.Consumes = append(operation.Consumes, "text/plain")
		case "html", "text/html":
			operation.Consumes = append(operation.Consumes, "text/html")
		case "mpfd", "multipart/form-data":
			operation.Consumes = append(operation.Consumes, "multipart/form-data")
		case "x-www-form-urlencoded", "application/x-www-form-urlencoded":
			operation.Consumes = append(operation.Consumes, "application/x-www-form-urlencoded")
		case "json-api", "application/vnd.api+json":
			operation.Consumes = append(operation.Consumes, "application/vnd.api+json")
		case "json-stream", "application/x-json-stream":
			operation.Consumes = append(operation.Consumes, "application/x-json-stream")
		case "octet-stream", "application/octet-stream":
			operation.Consumes = append(operation.Consumes, "application/octet-stream")
		case "png", "image/png":
			operation.Consumes = append(operation.Consumes, "image/png")
		case "jpeg", "image/jpeg":
			operation.Consumes = append(operation.Consumes, "image/jpeg")
		case "gif", "image/gif":
			operation.Consumes = append(operation.Consumes, "image/gif")
		default:
			return fmt.Errorf("%v accept type can't accepted", a)
		}
	}
	return nil
}

// ParseProduceComment parses comment for gived `produce` comment string.
func (operation *Operation) ParseProduceComment(commentLine string) error {
	produces := strings.Split(commentLine, ",")
	for _, a := range produces {
		switch a {
		case "json", "application/json":
			operation.Produces = append(operation.Produces, "application/json")
		case "xml", "text/xml":
			operation.Produces = append(operation.Produces, "text/xml")
		case "plain", "text/plain":
			operation.Produces = append(operation.Produces, "text/plain")
		case "html", "text/html":
			operation.Produces = append(operation.Produces, "text/html")
		case "mpfd", "multipart/form-data":
			operation.Produces = append(operation.Produces, "multipart/form-data")
		case "x-www-form-urlencoded", "application/x-www-form-urlencoded":
			operation.Produces = append(operation.Produces, "application/x-www-form-urlencoded")
		case "json-api", "application/vnd.api+json":
			operation.Produces = append(operation.Produces, "application/vnd.api+json")
		case "json-stream", "application/x-json-stream":
			operation.Produces = append(operation.Produces, "application/x-json-stream")
		case "octet-stream", "application/octet-stream":
			operation.Produces = append(operation.Produces, "application/octet-stream")
		case "png", "image/png":
			operation.Produces = append(operation.Produces, "image/png")
		case "jpeg", "image/jpeg":
			operation.Produces = append(operation.Produces, "image/jpeg")
		case "gif", "image/gif":
			operation.Produces = append(operation.Produces, "image/gif")
		default:
			return fmt.Errorf("%v produce type can't accepted", a)
		}
	}
	return nil
}

// ParseRouterComment parses comment for gived `router` comment string.
func (operation *Operation) ParseRouterComment(commentLine string) error {
	re := regexp.MustCompile(`([\w\.\/\-{}]+)[^\[]+\[([^\]]+)`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 3 {
		return fmt.Errorf("can not parse router comment \"%s\"", commentLine)
	}
	path := matches[1]
	httpMethod := matches[2]

	operation.Path = path
	operation.HTTPMethod = strings.ToUpper(httpMethod)

	return nil
}

// ParseSecurityComment parses comment for gived `security` comment string.
func (operation *Operation) ParseSecurityComment(commentLine string) error {
	securitySource := commentLine[strings.Index(commentLine, "@Security")+1:]
	l := strings.Index(securitySource, "[")
	r := strings.Index(securitySource, "]")
	// exists scope
	if !(l == -1 && r == -1) {
		scopes := securitySource[l+1 : r]
		s := []string{}
		for _, scope := range strings.Split(scopes, ",") {
			scope = strings.TrimSpace(scope)
			s = append(s, scope)
		}
		securityKey := securitySource[0:l]
		securityMap := map[string][]string{}
		securityMap[securityKey] = append(securityMap[securityKey], s...)
		operation.Security = append(operation.Security, securityMap)
	} else {
		securityKey := strings.TrimSpace(securitySource)
		securityMap := map[string][]string{}
		securityMap[securityKey] = []string{}
		operation.Security = append(operation.Security, securityMap)
	}
	return nil
}

// ParseResponseComment parses comment for gived `response` comment string.
func (operation *Operation) ParseResponseComment(commentLine string) error {
	re := regexp.MustCompile(`([\d]+)[\s]+([\w\{\}]+)[\s]+([\w\-\.\/]+)[^"]*(.*)?`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 5 {
		return fmt.Errorf("can not parse response comment \"%s\"", commentLine)
	}

	response := spec.Response{}

	code, _ := strconv.Atoi(matches[1])

	responseDescription := strings.Trim(matches[4], "\"")
	if responseDescription == "" {
		responseDescription = http.StatusText(code)
	}
	response.Description = responseDescription

	schemaType := strings.Trim(matches[2], "{}")
	refType := matches[3]

	if operation.parser != nil { // checking refType has existing in 'TypeDefinitions'
		refSplit := strings.Split(refType, ".")
		if len(refSplit) == 2 {
			pkgName := refSplit[0]
			typeName := refSplit[1]
			if typeSpec, ok := operation.parser.TypeDefinitions[pkgName][typeName]; ok {
				operation.parser.registerTypes[refType] = typeSpec
			} else {
				return fmt.Errorf("can not find ref type:\"%s\"", refType)
			}

		}
	}

	// so we have to know all type in app
	//TODO: we might omitted schema.type if schemaType equals 'object'
	response.Schema = &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{schemaType}}}

	if schemaType == "object" {
		response.Schema.Ref = spec.Ref{
			Ref: jsonreference.MustCreateRef("#/definitions/" + refType),
		}
	}

	if schemaType == "array" {
		response.Schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Ref: spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/" + refType)},
				},
			},
		}

	}

	if operation.Responses == nil {
		operation.Responses = &spec.Responses{
			ResponsesProps: spec.ResponsesProps{
				StatusCodeResponses: make(map[int]spec.Response),
			},
		}
	}

	operation.Responses.StatusCodeResponses[code] = response

	return nil
}

// ParseEmptyResponseComment TODO: NEEDS COMMENT INFO
func (operation *Operation) ParseEmptyResponseComment(commentLine string) error {
	re := regexp.MustCompile(`([\d]+)[\s]+"(.*)"`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 3 {
		return fmt.Errorf("can not parse empty response comment \"%s\"", commentLine)
	}

	response := spec.Response{}

	code, _ := strconv.Atoi(matches[1])

	response.Description = strings.Trim(matches[2], "")

	if operation.Responses == nil {
		operation.Responses = &spec.Responses{
			ResponsesProps: spec.ResponsesProps{
				StatusCodeResponses: make(map[int]spec.Response),
			},
		}
	}

	operation.Responses.StatusCodeResponses[code] = response

	return nil
}

// createParameter returns swagger spec.Parameter for gived  paramType, description, paramName, schemaType, required
func createParameter(paramType, description, paramName, schemaType string, required bool) spec.Parameter {
	// //five possible parameter types. 	query, path, body, header, form
	paramProps := spec.ParamProps{
		Name:        paramName,
		Description: description,
		Required:    required,
		In:          paramType,
	}
	if paramType == "body" {
		paramProps.Schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{schemaType},
			},
		}
		parameter := spec.Parameter{
			ParamProps: paramProps,
		}
		return parameter
	}
	parameter := spec.Parameter{
		ParamProps: paramProps,
		SimpleSchema: spec.SimpleSchema{
			Type: schemaType,
		},
	}
	return parameter
}
