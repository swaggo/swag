package swag

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/loader"
)

// Operation describes a single API operation on a path.
// For more information: https://github.com/swaggo/swag#api-operation
type Operation struct {
	HTTPMethod string
	Path       string
	spec.Operation

	parser *Parser
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

// ParseComment parses comment for given comment string and returns error if error occurs.
func (operation *Operation) ParseComment(comment string, astFile *ast.File) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "//"))
	if len(commentLine) == 0 {
		return nil
	}

	attribute := strings.Fields(commentLine)[0]
	lineRemainder := strings.TrimSpace(commentLine[len(attribute):])
	switch strings.ToLower(attribute) {
	case "@description":
		operation.Description = lineRemainder
	case "@summary":
		operation.Summary = lineRemainder
	case "@id":
		operation.ID = lineRemainder
	case "@tags":
		operation.ParseTagsComment(lineRemainder)
	case "@accept":
		if err := operation.ParseAcceptComment(lineRemainder); err != nil {
			return err
		}
	case "@produce":
		if err := operation.ParseProduceComment(lineRemainder); err != nil {
			return err
		}
	case "@param":
		if err := operation.ParseParamComment(lineRemainder, astFile); err != nil {
			return err
		}
	case "@success", "@failure":
		if err := operation.ParseResponseComment(lineRemainder, astFile); err != nil {
			if err := operation.ParseEmptyResponseComment(lineRemainder); err != nil {
				if err := operation.ParseEmptyResponseOnly(lineRemainder); err != nil {
					return err
				}
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

// ParseParamComment parses params return []string of param properties
// @Param	queryText		form	      string	  true		        "The email for login"
// 			[param name]    [paramType] [data type]  [is mandatory?]   [Comment]
// @Param   some_id     path    int     true        "Some ID"
func (operation *Operation) ParseParamComment(commentLine string, astFile *ast.File) error {
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
				var typeSpec *ast.TypeSpec
				if astFile != nil {
					for _, imp := range astFile.Imports {
						if imp.Name != nil && imp.Name.Name == pkgName { // the import had an alias that matched
							break
						}
						impPath := strings.Replace(imp.Path.Value, `"`, ``, -1)
						if strings.HasSuffix(impPath, "/"+pkgName) {
							var err error
							typeSpec, err = findTypeDef(impPath, typeName)
							if err != nil {
								return errors.Wrapf(err, "can not find ref type: %q", schemaType)
							}
							break
						}
					}
				}

				if typeSpec == nil {
					return fmt.Errorf("can not find ref type:\"%s\"", schemaType)
				}

				operation.parser.TypeDefinitions[pkgName][typeName] = typeSpec
				operation.parser.registerTypes[schemaType] = typeSpec

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

// ParseTagsComment parses comment for given `tag` comment string.
func (operation *Operation) ParseTagsComment(commentLine string) {
	tags := strings.Split(commentLine, ",")
	for _, tag := range tags {
		operation.Tags = append(operation.Tags, strings.TrimSpace(tag))
	}
}

// ParseAcceptComment parses comment for given `accept` comment string.
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

// findTypeDef attempts to find the *ast.TypeSpec for a specific type given the
// type's name and the package's import path
func findTypeDef(importPath, typeName string) (*ast.TypeSpec, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	conf := loader.Config{
		ParserMode: goparser.SpuriousErrors,
		Cwd:        cwd,
	}

	conf.Import(importPath)

	lprog, err := conf.Load()
	if err != nil {
		return nil, err
	}

	// If the pkg is vendored, the actual pkg path is going to resemble
	// something like "{importPath}/vendor/{importPath}"
	for k := range lprog.AllPackages {
		realPkgPath := k.Path()

		if strings.Contains(realPkgPath, "vendor/"+importPath) {
			importPath = realPkgPath
		}
	}

	pkgInfo := lprog.Package(importPath)

	if pkgInfo == nil {
		return nil, errors.New("package was nil")
	}

	// TODO: possibly cache pkgInfo since it's an expensive operation

	for i := range pkgInfo.Files {
		for _, astDeclaration := range pkgInfo.Files[i].Decls {
			if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
				for _, astSpec := range generalDeclaration.Specs {
					if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
						if typeSpec.Name.String() == typeName {
							return typeSpec, nil
						}
					}
				}
			}
		}
	}
	return nil, errors.New("type spec not found")
}

// ParseResponseComment parses comment for gived `response` comment string.
func (operation *Operation) ParseResponseComment(commentLine string, astFile *ast.File) error {
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
				var typeSpec *ast.TypeSpec
				if astFile != nil {
					for _, imp := range astFile.Imports {
						if imp.Name != nil && imp.Name.Name == pkgName { // the import had an alias that matched
							break
						}
						impPath := strings.Replace(imp.Path.Value, `"`, ``, -1)

						if strings.HasSuffix(impPath, "/"+pkgName) {
							var err error

							typeSpec, err = findTypeDef(impPath, typeName)
							if err != nil {
								return errors.Wrapf(err, "can not find ref type: %q", refType)
							}
							break
						}
					}
				}

				if typeSpec == nil {
					return fmt.Errorf("can not find ref type: %q", refType)
				}

				if _, ok := operation.parser.TypeDefinitions[pkgName]; !ok {
					operation.parser.TypeDefinitions[pkgName] = make(map[string]*ast.TypeSpec)

				}
				operation.parser.TypeDefinitions[pkgName][typeName] = typeSpec
				operation.parser.registerTypes[refType] = typeSpec
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
		refType = TransToValidSchemeType(refType)
		if IsPrimitiveType(refType) {
			response.Schema.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type: spec.StringOrArray{refType},
					},
				},
			}
		} else {
			response.Schema.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Ref: spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/" + refType)},
					},
				},
			}
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

// ParseEmptyResponseComment parse only comment out status code and description,eg: @Success 200 "it's ok"
func (operation *Operation) ParseEmptyResponseComment(commentLine string) error {
	re := regexp.MustCompile(`([\d]+)[\s]+"(.*)"`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 3 {
		return fmt.Errorf("can not parse response comment \"%s\"", commentLine)
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

//ParseEmptyResponseOnly parse only comment out status code ,eg: @Success 200
func (operation *Operation) ParseEmptyResponseOnly(commentLine string) error {
	response := spec.Response{}

	code, err := strconv.Atoi(commentLine)
	if err != nil {
		return fmt.Errorf("can not parse response comment \"%s\"", commentLine)
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
