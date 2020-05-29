package swag

import (
	"encoding/json"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
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

var mimeTypeAliases = map[string]string{
	"json":                  "application/json",
	"xml":                   "text/xml",
	"plain":                 "text/plain",
	"html":                  "text/html",
	"mpfd":                  "multipart/form-data",
	"x-www-form-urlencoded": "application/x-www-form-urlencoded",
	"json-api":              "application/vnd.api+json",
	"json-stream":           "application/x-json-stream",
	"octet-stream":          "application/octet-stream",
	"png":                   "image/png",
	"jpeg":                  "image/jpeg",
	"gif":                   "image/gif",
}

var mimeTypePattern = regexp.MustCompile("^[^/]+/[^/]+$")

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
	lowerAttribute := strings.ToLower(attribute)

	var err error
	switch lowerAttribute {
	case "@description":
		operation.ParseDescriptionComment(lineRemainder)
	case "@description.markdown":
		commentInfo, err := getMarkdownForTag(lineRemainder, operation.parser.markdownFileDir)
		if err != nil {
			return err
		}
		operation.ParseDescriptionComment(string(commentInfo))
	case "@summary":
		operation.Summary = lineRemainder
	case "@id":
		operation.ID = lineRemainder
	case "@tags":
		operation.ParseTagsComment(lineRemainder)
	case "@accept":
		err = operation.ParseAcceptComment(lineRemainder)
	case "@produce":
		err = operation.ParseProduceComment(lineRemainder)
	case "@param":
		err = operation.ParseParamComment(lineRemainder, astFile)
	case "@success", "@failure":
		err = operation.ParseResponseComment(lineRemainder, astFile)
	case "@header":
		err = operation.ParseResponseHeaderComment(lineRemainder, astFile)
	case "@router":
		err = operation.ParseRouterComment(lineRemainder)
	case "@security":
		err = operation.ParseSecurityComment(lineRemainder)
	case "@deprecated":
		operation.Deprecate()
	default:
		err = operation.ParseMetadata(attribute, lowerAttribute, lineRemainder)
	}

	return err
}

// ParseDescriptionComment godoc
func (operation *Operation) ParseDescriptionComment(lineRemainder string) {
	if operation.Description == "" {
		operation.Description = lineRemainder
		return
	}
	operation.Description += "\n" + lineRemainder
}

// ParseMetadata godoc
func (operation *Operation) ParseMetadata(attribute, lowerAttribute, lineRemainder string) error {
	// parsing specific meta data extensions
	if strings.HasPrefix(lowerAttribute, "@x-") {
		if len(lineRemainder) == 0 {
			return fmt.Errorf("annotation %s need a value", attribute)
		}

		var valueJSON interface{}
		if err := json.Unmarshal([]byte(lineRemainder), &valueJSON); err != nil {
			return fmt.Errorf("annotation %s need a valid json value", attribute)
		}
		operation.Operation.AddExtension(attribute[1:], valueJSON) // Trim "@" at head
	}
	return nil
}

var paramPattern = regexp.MustCompile(`(\S+)[\s]+([\w]+)[\s]+([\S.]+)[\s]+([\w]+)[\s]+"([^"]+)"`)

// ParseParamComment parses params return []string of param properties
// E.g. @Param	queryText		formData	      string	  true		        "The email for login"
//              [param name]    [paramType] [data type]  [is mandatory?]   [Comment]
// E.g. @Param   some_id     path    int     true        "Some ID"
func (operation *Operation) ParseParamComment(commentLine string, astFile *ast.File) error {
	matches := paramPattern.FindStringSubmatch(commentLine)
	if len(matches) != 6 {
		return fmt.Errorf("missing required param comment parameters \"%s\"", commentLine)
	}
	name := matches[1]
	paramType := matches[2]
	refType := TransToValidSchemeType(matches[3])

	// Detect refType
	objectType := "object"
	if strings.HasPrefix(refType, "[]") {
		objectType = "array"
		refType = strings.TrimPrefix(refType, "[]")
		refType = TransToValidSchemeType(refType)
	} else if IsPrimitiveType(refType) ||
		paramType == "formData" && refType == "file" {
		objectType = "primitive"
	}

	requiredText := strings.ToLower(matches[4])
	required := requiredText == "true" || requiredText == "required"
	description := matches[5]

	param := createParameter(paramType, description, name, refType, required)

	switch paramType {
	case "path", "header", "formData":
		switch objectType {
		case "array", "object":
			return fmt.Errorf("%s is not supported type for %s", refType, paramType)
		}
	case "query":
		switch objectType {
		case "array":
			if !IsPrimitiveType(refType) {
				return fmt.Errorf("%s is not supported array type for %s", refType, paramType)
			}
			param.SimpleSchema.Type = "array"
			if operation.parser != nil {
				param.CollectionFormat = TransToValidCollectionFormat(operation.parser.collectionFormatInQuery)
			}
			param.SimpleSchema.Items = &spec.Items{
				SimpleSchema: spec.SimpleSchema{
					Type: refType,
				},
			}
		case "object":
			refType, typeSpec, err := operation.registerSchemaType(refType, astFile)
			if err != nil {
				return err
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return fmt.Errorf("%s is not supported type for %s", refType, paramType)
			}
			refSplit := strings.Split(refType, ".")
			schema, err := operation.parser.parseStruct(refSplit[0], structType.Fields)
			if err != nil {
				return err
			}
			if len(schema.Properties) == 0 {
				return nil
			}
			find := func(arr []string, target string) bool {
				for _, str := range arr {
					if str == target {
						return true
					}
				}
				return false
			}
			orderedNames := make([]string, 0, len(schema.Properties))
			for k := range schema.Properties {
				orderedNames = append(orderedNames, k)
			}
			sort.Strings(orderedNames)
			for _, name := range orderedNames {
				prop := schema.Properties[name]
				if len(prop.Type) == 0 {
					continue
				}
				if prop.Type[0] == "array" &&
					prop.Items.Schema != nil &&
					len(prop.Items.Schema.Type) > 0 &&
					IsSimplePrimitiveType(prop.Items.Schema.Type[0]) {
					param = createParameter(paramType, prop.Description, name, prop.Type[0], find(schema.Required, name))
					param.SimpleSchema.Type = prop.Type[0]
					if operation.parser != nil && operation.parser.collectionFormatInQuery != "" && param.CollectionFormat == "" {
						param.CollectionFormat = TransToValidCollectionFormat(operation.parser.collectionFormatInQuery)
					}
					param.SimpleSchema.Items = &spec.Items{
						SimpleSchema: spec.SimpleSchema{
							Type: prop.Items.Schema.Type[0],
						},
					}
				} else if IsSimplePrimitiveType(prop.Type[0]) {
					param = createParameter(paramType, prop.Description, name, prop.Type[0], find(schema.Required, name))
				} else {
					Println(fmt.Sprintf("skip field [%s] in %s is not supported type for %s", name, refType, paramType))
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
				operation.Operation.Parameters = append(operation.Operation.Parameters, param)
			}
			return nil
		}
	case "body":
		switch objectType {
		case "primitive":
			param.Schema.Type = spec.StringOrArray{refType}
		case "array":
			refType = "[]" + refType
			fallthrough
		case "object":
			schema, err := operation.parseObjectSchema(refType, astFile)
			if err != nil {
				return err
			}
			param.Schema = schema
		}
	default:
		return fmt.Errorf("%s is not supported paramType", paramType)
	}

	if err := operation.parseAndExtractionParamAttribute(commentLine, objectType, refType, &param); err != nil {
		return err
	}
	operation.Operation.Parameters = append(operation.Operation.Parameters, param)
	return nil
}

func (operation *Operation) registerSchemaType(schemaType string, astFile *ast.File) (string, *ast.TypeSpec, error) {
	if !strings.ContainsRune(schemaType, '.') {
		if astFile == nil {
			return schemaType, nil, fmt.Errorf("no package name for type %s", schemaType)
		}
		schemaType = fullTypeName(astFile.Name.String(), schemaType)
	}
	refSplit := strings.Split(schemaType, ".")
	pkgName := refSplit[0]
	typeName := refSplit[1]
	if typeSpec, ok := operation.parser.TypeDefinitions[pkgName][typeName]; ok {
		operation.parser.registerTypes[schemaType] = typeSpec
		return schemaType, typeSpec, nil
	}
	var typeSpec *ast.TypeSpec
	if astFile == nil {
		return schemaType, nil, fmt.Errorf("can not register schema type: %q reason: astFile == nil", schemaType)
	}
	for _, imp := range astFile.Imports {
		if imp.Name != nil && imp.Name.Name == pkgName { // the import had an alias that matched
			break
		}
		impPath := strings.Replace(imp.Path.Value, `"`, ``, -1)
		if strings.HasSuffix(impPath, "/"+pkgName) {
			var err error
			typeSpec, err = findTypeDef(impPath, typeName)
			if err != nil {
				return schemaType, nil, fmt.Errorf("can not find type def: %q error: %s", schemaType, err)
			}
			break
		}
	}

	if typeSpec == nil {
		return schemaType, nil, fmt.Errorf("can not find schema type: %q", schemaType)
	}

	if _, ok := operation.parser.TypeDefinitions[pkgName]; !ok {
		operation.parser.TypeDefinitions[pkgName] = make(map[string]*ast.TypeSpec)
	}

	operation.parser.TypeDefinitions[pkgName][typeName] = typeSpec
	operation.parser.registerTypes[schemaType] = typeSpec
	return schemaType, typeSpec, nil
}

var regexAttributes = map[string]*regexp.Regexp{
	// for Enums(A, B)
	"enums": regexp.MustCompile(`(?i)\s+enums\(.*\)`),
	// for Minimum(0)
	"maxinum": regexp.MustCompile(`(?i)\s+maxinum\(.*\)`),
	// for Maximum(0)
	"mininum": regexp.MustCompile(`(?i)\s+mininum\(.*\)`),
	// for Maximum(0)
	"default": regexp.MustCompile(`(?i)\s+default\(.*\)`),
	// for minlength(0)
	"minlength": regexp.MustCompile(`(?i)\s+minlength\(.*\)`),
	// for maxlength(0)
	"maxlength": regexp.MustCompile(`(?i)\s+maxlength\(.*\)`),
	// for format(email)
	"format": regexp.MustCompile(`(?i)\s+format\(.*\)`),
	// for collectionFormat(csv)
	"collectionFormat": regexp.MustCompile(`(?i)\s+collectionFormat\(.*\)`),
}

func (operation *Operation) parseAndExtractionParamAttribute(commentLine, objectType, schemaType string, param *spec.Parameter) error {
	schemaType = TransToValidSchemeType(schemaType)
	for attrKey, re := range regexAttributes {
		attr, err := findAttr(re, commentLine)
		if err != nil {
			continue
		}
		switch attrKey {
		case "enums":
			err := setEnumParam(attr, schemaType, param)
			if err != nil {
				return err
			}
		case "maxinum":
			n, err := setNumberParam(attrKey, schemaType, attr, commentLine)
			if err != nil {
				return err
			}
			param.Maximum = &n
		case "mininum":
			n, err := setNumberParam(attrKey, schemaType, attr, commentLine)
			if err != nil {
				return err
			}
			param.Minimum = &n
		case "default":
			value, err := defineType(schemaType, attr)
			if err != nil {
				return nil
			}
			param.Default = value
		case "maxlength":
			n, err := setStringParam(attrKey, schemaType, attr, commentLine)
			if err != nil {
				return err
			}
			param.MaxLength = &n
		case "minlength":
			n, err := setStringParam(attrKey, schemaType, attr, commentLine)
			if err != nil {
				return err
			}
			param.MinLength = &n
		case "format":
			param.Format = attr
		case "collectionFormat":
			n, err := setCollectionFormatParam(attrKey, objectType, attr, commentLine)
			if err != nil {
				return err
			}
			param.CollectionFormat = n
		}
	}
	return nil
}

func findAttr(re *regexp.Regexp, commentLine string) (string, error) {
	attr := re.FindString(commentLine)
	l := strings.Index(attr, "(")
	r := strings.Index(attr, ")")
	if l == -1 || r == -1 {
		return "", fmt.Errorf("can not find regex=%s, comment=%s", re.String(), commentLine)
	}
	return strings.TrimSpace(attr[l+1 : r]), nil
}

func setStringParam(name, schemaType, attr, commentLine string) (int64, error) {
	if schemaType != "string" {
		return 0, fmt.Errorf("%s is attribute to set to a number. comment=%s got=%s", name, commentLine, schemaType)
	}
	n, err := strconv.ParseInt(attr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is allow only a number got=%s", name, attr)
	}
	return n, nil
}

func setNumberParam(name, schemaType, attr, commentLine string) (float64, error) {
	if schemaType != "integer" && schemaType != "number" {
		return 0, fmt.Errorf("%s is attribute to set to a number. comment=%s got=%s", name, commentLine, schemaType)
	}
	n, err := strconv.ParseFloat(attr, 64)
	if err != nil {
		return 0, fmt.Errorf("maximum is allow only a number. comment=%s got=%s", commentLine, attr)
	}
	return n, nil
}

func setEnumParam(attr, schemaType string, param *spec.Parameter) error {
	for _, e := range strings.Split(attr, ",") {
		e = strings.TrimSpace(e)

		value, err := defineType(schemaType, e)
		if err != nil {
			return err
		}
		param.Enum = append(param.Enum, value)
	}
	return nil
}

func setCollectionFormatParam(name, schemaType, attr, commentLine string) (string, error) {
	if schemaType != "array" {
		return "", fmt.Errorf("%s is attribute to set to an array. comment=%s got=%s", name, commentLine, schemaType)
	}
	return TransToValidCollectionFormat(attr), nil
}

// defineType enum value define the type (object and array unsupported)
func defineType(schemaType string, value string) (interface{}, error) {
	schemaType = TransToValidSchemeType(schemaType)
	switch schemaType {
	case "string":
		return value, nil
	case "number":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err)
		}
		return v, nil
	case "integer":
		v, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err)
		}
		return v, nil
	case "boolean":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("%s is unsupported type in enum value", schemaType)
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
	return parseMimeTypeList(commentLine, &operation.Consumes, "%v accept type can't be accepted")
}

// ParseProduceComment parses comment for given `produce` comment string.
func (operation *Operation) ParseProduceComment(commentLine string) error {
	return parseMimeTypeList(commentLine, &operation.Produces, "%v produce type can't be accepted")
}

// parseMimeTypeList parses a list of MIME Types for a comment like
// `produce` (`Content-Type:` response header) or
// `accept` (`Accept:` request header)
func parseMimeTypeList(mimeTypeList string, typeList *[]string, format string) error {
	mimeTypes := strings.Split(mimeTypeList, ",")
	for _, typeName := range mimeTypes {
		if mimeTypePattern.MatchString(typeName) {
			*typeList = append(*typeList, typeName)
			continue
		}
		if aliasMimeType, ok := mimeTypeAliases[typeName]; ok {
			*typeList = append(*typeList, aliasMimeType)
			continue
		}
		return fmt.Errorf(format, typeName)
	}
	return nil
}

var routerPattern = regexp.MustCompile(`^(/[\w\.\/\-{}\+:]*)[[:blank:]]+\[(\w+)]`)

// ParseRouterComment parses comment for gived `router` comment string.
func (operation *Operation) ParseRouterComment(commentLine string) error {
	var matches []string

	if matches = routerPattern.FindStringSubmatch(commentLine); len(matches) != 3 {
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
// TODO: improve finding external pkg
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
		return nil, fmt.Errorf("package was nil")
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
	return nil, fmt.Errorf("type spec not found")
}

var responsePattern = regexp.MustCompile(`([\d]+)[\s]+([\w\{\}]+)[\s]+([\w\-\.\/\{\}=,\[\]]+)[^"]*(.*)?`)

//RepsonseType{data1=Type1,data2=Type2}
var combinedPattern = regexp.MustCompile(`^([\w\-\.\/\[\]]+)\{(.*)\}$`)

func (operation *Operation) parseObjectSchema(refType string, astFile *ast.File) (*spec.Schema, error) {
	switch {
	case refType == "interface{}":
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"object"}}}, nil
	case IsGolangPrimitiveType(refType):
		refType = TransToValidSchemeType(refType)
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{refType}}}, nil
	case IsPrimitiveType(refType):
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{refType}}}, nil
	case strings.HasPrefix(refType, "[]"):
		schema, err := operation.parseObjectSchema(refType[2:], astFile)
		if err != nil {
			return nil, err
		}
		return &spec.Schema{SchemaProps: spec.SchemaProps{
			Type:  []string{"array"},
			Items: &spec.SchemaOrArray{Schema: schema}},
		}, nil
	case strings.HasPrefix(refType, "map["):
		//ignore key type
		idx := strings.Index(refType, "]")
		if idx < 0 {
			return nil, fmt.Errorf("invalid type: %s", refType)
		}
		refType = refType[idx+1:]
		var valueSchema spec.SchemaOrBool
		if refType == "interface{}" {
			valueSchema.Allows = true
		} else {
			schema, err := operation.parseObjectSchema(refType, astFile)
			if err != nil {
				return &spec.Schema{}, err
			}
			valueSchema.Schema = schema
		}
		return &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:                 []string{"object"},
				AdditionalProperties: &valueSchema,
			},
		}, nil
	case strings.Contains(refType, "{"):
		return operation.parseResponseCombinedObjectSchema(refType, astFile)
	default:
		if operation.parser != nil { // checking refType has existing in 'TypeDefinitions'
			refNewType, typeSpec, err := operation.registerSchemaType(refType, astFile)
			if err != nil {
				return nil, err
			}
			refType = TypeDocName(refNewType, typeSpec)
		}
		return &spec.Schema{SchemaProps: spec.SchemaProps{Ref: spec.Ref{
			Ref: jsonreference.MustCreateRef("#/definitions/" + refType),
		}}}, nil
	}
}

func (operation *Operation) parseResponseCombinedObjectSchema(refType string, astFile *ast.File) (*spec.Schema, error) {
	matches := combinedPattern.FindStringSubmatch(refType)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid type: %s", refType)
	}
	refType = matches[1]
	schema, err := operation.parseObjectSchema(refType, astFile)
	if err != nil {
		return nil, err
	}

	parseFields := func(s string) []string {
		n := 0
		return strings.FieldsFunc(s, func(r rune) bool {
			if r == '{' {
				n++
				return false
			} else if r == '}' {
				n--
				return false
			}
			return r == ',' && n == 0
		})
	}

	fields := parseFields(matches[2])
	props := map[string]spec.Schema{}
	for _, field := range fields {
		if matches := strings.SplitN(field, "=", 2); len(matches) == 2 {
			if strings.HasPrefix(matches[1], "[]") {
				itemSchema, err := operation.parseObjectSchema(matches[1][2:], astFile)
				if err != nil {
					return nil, err
				}
				props[matches[0]] = spec.Schema{SchemaProps: spec.SchemaProps{
					Type:  []string{"array"},
					Items: &spec.SchemaOrArray{Schema: itemSchema}},
				}
			} else {
				schema, err := operation.parseObjectSchema(matches[1], astFile)
				if err != nil {
					return nil, err
				}
				props[matches[0]] = *schema
			}
		}
	}

	if len(props) == 0 {
		return schema, nil
	}
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			AllOf: []spec.Schema{
				*schema,
				{
					SchemaProps: spec.SchemaProps{
						Type:       []string{"object"},
						Properties: props,
					},
				},
			},
		},
	}, nil
}

func (operation *Operation) parseResponseSchema(schemaType, refType string, astFile *ast.File) (*spec.Schema, error) {
	switch schemaType {
	case "object":
		if !strings.HasPrefix(refType, "[]") {
			return operation.parseObjectSchema(refType, astFile)
		}
		refType = refType[2:]
		fallthrough
	case "array":
		schema, err := operation.parseObjectSchema(refType, astFile)
		if err != nil {
			return nil, err
		}
		return &spec.Schema{SchemaProps: spec.SchemaProps{
			Type:  []string{"array"},
			Items: &spec.SchemaOrArray{Schema: schema}},
		}, nil
	default:
		return &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{schemaType}}}, nil
	}
}

// ParseResponseComment parses comment for given `response` comment string.
func (operation *Operation) ParseResponseComment(commentLine string, astFile *ast.File) error {
	var matches []string

	if matches = responsePattern.FindStringSubmatch(commentLine); len(matches) != 5 {
		err := operation.ParseEmptyResponseComment(commentLine)
		if err != nil {
			return operation.ParseEmptyResponseOnly(commentLine)
		}
		return err
	}

	code, _ := strconv.Atoi(matches[1])

	responseDescription := strings.Trim(matches[4], "\"")
	if responseDescription == "" {
		responseDescription = http.StatusText(code)
	}

	schemaType := strings.Trim(matches[2], "{}")
	refType := matches[3]
	schema, err := operation.parseResponseSchema(schemaType, refType, astFile)
	if err != nil {
		return err
	}

	if operation.Responses == nil {
		operation.Responses = &spec.Responses{
			ResponsesProps: spec.ResponsesProps{
				StatusCodeResponses: make(map[int]spec.Response),
			},
		}
	}

	operation.Responses.StatusCodeResponses[code] = spec.Response{
		ResponseProps: spec.ResponseProps{Schema: schema, Description: responseDescription},
	}
	return nil
}

// ParseResponseHeaderComment parses comment for gived `response header` comment string.
func (operation *Operation) ParseResponseHeaderComment(commentLine string, astFile *ast.File) error {
	var matches []string

	if matches = responsePattern.FindStringSubmatch(commentLine); len(matches) != 5 {
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

	if operation.Responses == nil {
		operation.Responses = &spec.Responses{
			ResponsesProps: spec.ResponsesProps{
				StatusCodeResponses: make(map[int]spec.Response),
			},
		}
	}

	response, responseExist := operation.Responses.StatusCodeResponses[code]
	if responseExist {
		header := spec.Header{}
		header.Description = responseDescription
		header.Type = schemaType

		if response.Headers == nil {
			response.Headers = make(map[string]spec.Header)
		}
		response.Headers[refType] = header

		operation.Responses.StatusCodeResponses[code] = response
	}

	return nil
}

var emptyResponsePattern = regexp.MustCompile(`([\d]+)[\s]+"(.*)"`)

// ParseEmptyResponseComment parse only comment out status code and description,eg: @Success 200 "it's ok"
func (operation *Operation) ParseEmptyResponseComment(commentLine string) error {
	var matches []string

	if matches = emptyResponsePattern.FindStringSubmatch(commentLine); len(matches) != 3 {
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
