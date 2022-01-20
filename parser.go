package swag

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/KyleBanks/depth"
	"github.com/go-openapi/spec"
)

const (
	// CamelCase indicates using CamelCase strategy for struct field.
	CamelCase = "camelcase"

	// PascalCase indicates using PascalCase strategy for struct field.
	PascalCase = "pascalcase"

	// SnakeCase indicates using SnakeCase strategy for struct field.
	SnakeCase = "snakecase"

	idAttr                  = "@id"
	acceptAttr              = "@accept"
	produceAttr             = "@produce"
	paramAttr               = "@param"
	successAttr             = "@success"
	failureAttr             = "@failure"
	responseAttr            = "@response"
	headerAttr              = "@header"
	tagsAttr                = "@tags"
	routerAttr              = "@router"
	summaryAttr             = "@summary"
	deprecatedAttr          = "@deprecated"
	securityAttr            = "@security"
	titleAttr               = "@title"
	versionAttr             = "@version"
	descriptionAttr         = "@description"
	descriptionMarkdownAttr = "@description.markdown"
	xCodeSamplesAttr        = "@x-codesamples"
	scopeAttrPrefix         = "@scope."
)

var (
	// ErrRecursiveParseStruct recursively parsing struct.
	ErrRecursiveParseStruct = errors.New("recursively parsing struct")

	// ErrFuncTypeField field type is func.
	ErrFuncTypeField = errors.New("field type is func")

	// ErrFailedConvertPrimitiveType Failed to convert for swag to interpretable type.
	ErrFailedConvertPrimitiveType = errors.New("swag property: failed convert primitive type")

	// ErrSkippedField .swaggo specifies field should be skipped
	ErrSkippedField = errors.New("field is skipped by global overrides")
)

var allMethod = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodPut:     {},
	http.MethodPost:    {},
	http.MethodDelete:  {},
	http.MethodOptions: {},
	http.MethodHead:    {},
	http.MethodPatch:   {},
}

// Parser implements a parser for Go source files.
type Parser struct {
	// swagger represents the root document object for the API specification
	swagger *spec.Swagger

	// packages store entities of APIs, definitions, file, package path etc.  and their relations
	packages *PackagesDefinitions

	// parsedSchemas store schemas which have been parsed from ast.TypeSpec
	parsedSchemas map[*TypeSpecDef]*Schema

	// outputSchemas store schemas which will be export to swagger
	outputSchemas map[*TypeSpecDef]*Schema

	// existSchemaNames store names of models for conflict determination
	existSchemaNames map[string]*Schema

	// toBeRenamedSchemas names of models to be renamed
	toBeRenamedSchemas map[string]string

	// toBeRenamedSchemas URLs of ref models to be renamed
	toBeRenamedRefURLs []*url.URL

	// PropNamingStrategy naming strategy
	PropNamingStrategy string

	// ParseVendor parse vendor folder
	ParseVendor bool

	// ParseDependencies whether swag should be parse outside dependency folder
	ParseDependency bool

	// ParseInternal whether swag should parse internal packages
	ParseInternal bool

	// Strict whether swag should error or warn when it detects cases which are most likely user errors
	Strict bool

	// structStack stores full names of the structures that were already parsed or are being parsed now
	structStack []*TypeSpecDef

	// markdownFileDir holds the path to the folder, where markdown files are stored
	markdownFileDir string

	// codeExampleFilesDir holds path to the folder, where code example files are stored
	codeExampleFilesDir string

	// collectionFormatInQuery set the default collectionFormat otherwise then 'csv' for array in query params
	collectionFormatInQuery string

	// excludes excludes dirs and files in SearchDir
	excludes map[string]struct{}

	// debugging output goes here
	debug Debugger

	// fieldParserFactory create FieldParser
	fieldParserFactory FieldParserFactory

	// Overrides allows global replacements of types. A blank replacement will be skipped.
	Overrides map[string]string
}

// FieldParserFactory create FieldParser
type FieldParserFactory func(ps *Parser, field *ast.Field) FieldParser

// FieldParser parse struct field
type FieldParser interface {
	ShouldSkip() (bool, error)
	FieldName() (string, error)
	CustomSchema() (*spec.Schema, error)
	ComplementSchema(schema *spec.Schema) error
	IsRequired() (bool, error)
}

// Debugger is the interface that wraps the basic Printf method.
type Debugger interface {
	Printf(format string, v ...interface{})
}

// New creates a new Parser with default properties.
func New(options ...func(*Parser)) *Parser {
	// parser.swagger.SecurityDefinitions =

	parser := &Parser{
		swagger: &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Info: &spec.Info{
					InfoProps: spec.InfoProps{
						Contact: &spec.ContactInfo{},
						License: nil,
					},
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{},
					},
				},
				Paths: &spec.Paths{
					Paths: make(map[string]spec.PathItem),
				},
				Definitions:         make(map[string]spec.Schema),
				SecurityDefinitions: make(map[string]*spec.SecurityScheme),
			},
		},
		packages:           NewPackagesDefinitions(),
		debug:              log.New(os.Stdout, "", log.LstdFlags),
		parsedSchemas:      make(map[*TypeSpecDef]*Schema),
		outputSchemas:      make(map[*TypeSpecDef]*Schema),
		existSchemaNames:   make(map[string]*Schema),
		toBeRenamedSchemas: make(map[string]string),
		excludes:           make(map[string]struct{}),
		fieldParserFactory: newTagBaseFieldParser,
		Overrides:          make(map[string]string),
	}

	for _, option := range options {
		option(parser)
	}

	return parser
}

// SetMarkdownFileDirectory sets the directory to search for markdown files.
func SetMarkdownFileDirectory(directoryPath string) func(*Parser) {
	return func(p *Parser) {
		p.markdownFileDir = directoryPath
	}
}

// SetCodeExamplesDirectory sets the directory to search for code example files.
func SetCodeExamplesDirectory(directoryPath string) func(*Parser) {
	return func(p *Parser) {
		p.codeExampleFilesDir = directoryPath
	}
}

// SetExcludedDirsAndFiles sets directories and files to be excluded when searching.
func SetExcludedDirsAndFiles(excludes string) func(*Parser) {
	return func(p *Parser) {
		for _, f := range strings.Split(excludes, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				f = filepath.Clean(f)
				p.excludes[f] = struct{}{}
			}
		}
	}
}

// SetStrict sets whether swag should error or warn when it detects cases which are most likely user errors.
func SetStrict(strict bool) func(*Parser) {
	return func(p *Parser) {
		p.Strict = strict
	}
}

// SetDebugger allows the use of user-defined implementations.
func SetDebugger(logger Debugger) func(parser *Parser) {
	return func(p *Parser) {
		p.debug = logger
	}
}

// SetFieldParserFactory allows the use of user-defined implementations.
func SetFieldParserFactory(factory FieldParserFactory) func(parser *Parser) {
	return func(p *Parser) {
		p.fieldParserFactory = factory
	}
}

// SetOverrides allows the use of user-defined global type overrides.
func SetOverrides(overrides map[string]string) func(parser *Parser) {
	return func(p *Parser) {
		for k, v := range overrides {
			p.Overrides[k] = v
		}
	}
}

// ParseAPI parses general api info for given searchDir and mainAPIFile.
func (parser *Parser) ParseAPI(searchDir string, mainAPIFile string, parseDepth int) error {
	return parser.ParseAPIMultiSearchDir([]string{searchDir}, mainAPIFile, parseDepth)
}

// ParseAPIMultiSearchDir is like ParseAPI but for multiple search dirs.
func (parser *Parser) ParseAPIMultiSearchDir(searchDirs []string, mainAPIFile string, parseDepth int) error {
	for _, searchDir := range searchDirs {
		parser.debug.Printf("Generate general API Info, search dir:%s", searchDir)

		packageDir, err := getPkgName(searchDir)
		if err != nil {
			parser.debug.Printf("warning: failed to get package name in dir: %s, error: %s", searchDir, err.Error())
		}

		err = parser.getAllGoFileInfo(packageDir, searchDir)
		if err != nil {
			return err
		}
	}

	absMainAPIFilePath, err := filepath.Abs(filepath.Join(searchDirs[0], mainAPIFile))
	if err != nil {
		return err
	}

	if parser.ParseDependency {
		var t depth.Tree
		t.ResolveInternal = true
		t.MaxDepth = parseDepth

		pkgName, err := getPkgName(filepath.Dir(absMainAPIFilePath))
		if err != nil {
			return err
		}

		err = t.Resolve(pkgName)
		if err != nil {
			return fmt.Errorf("pkg %s cannot find all dependencies, %s", pkgName, err)
		}

		for i := 0; i < len(t.Root.Deps); i++ {
			err := parser.getAllGoFileInfoFromDeps(&t.Root.Deps[i])
			if err != nil {
				return err
			}
		}
	}

	err = parser.ParseGeneralAPIInfo(absMainAPIFilePath)
	if err != nil {
		return err
	}

	parser.parsedSchemas, err = parser.packages.ParseTypes()
	if err != nil {
		return err
	}

	err = rangeFiles(parser.packages.files, parser.ParseRouterAPIInfo)
	if err != nil {
		return err
	}

	parser.renameRefSchemas()

	return parser.checkOperationIDUniqueness()
}

func getPkgName(searchDir string) (string, error) {
	cmd := exec.Command("go", "list", "-f={{.ImportPath}}")
	cmd.Dir = searchDir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("execute go list command, %s, stdout:%s, stderr:%s", err, stdout.String(), stderr.String())
	}

	outStr, _ := stdout.String(), stderr.String()

	if outStr[0] == '_' { // will shown like _/{GOPATH}/src/{YOUR_PACKAGE} when NOT enable GO MODULE.
		outStr = strings.TrimPrefix(outStr, "_"+build.Default.GOPATH+"/src/")
	}
	f := strings.Split(outStr, "\n")
	outStr = f[0]

	return outStr, nil
}

func initIfEmpty(license *spec.License) *spec.License {
	if license == nil {
		return new(spec.License)
	}

	return license
}

// ParseGeneralAPIInfo parses general api info for given mainAPIFile path.
func (parser *Parser) ParseGeneralAPIInfo(mainAPIFile string) error {
	fileTree, err := goparser.ParseFile(token.NewFileSet(), mainAPIFile, nil, goparser.ParseComments)
	if err != nil {
		return fmt.Errorf("cannot parse source files %s: %s", mainAPIFile, err)
	}

	parser.swagger.Swagger = "2.0"

	for _, comment := range fileTree.Comments {
		comments := strings.Split(comment.Text(), "\n")
		if !isGeneralAPIComment(comments) {
			continue
		}
		err := parseGeneralAPIInfo(parser, comments)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseGeneralAPIInfo(parser *Parser, comments []string) error {
	previousAttribute := ""

	// parsing classic meta data model
	for i, commentLine := range comments {
		attribute := strings.Split(commentLine, " ")[0]
		value := strings.TrimSpace(commentLine[len(attribute):])
		multilineBlock := false
		if previousAttribute == attribute {
			multilineBlock = true
		}
		switch strings.ToLower(attribute) {
		case versionAttr:
			parser.swagger.Info.Version = value
		case titleAttr:
			parser.swagger.Info.Title = value
		case descriptionAttr:
			if multilineBlock {
				parser.swagger.Info.Description += "\n" + value

				continue
			}
			parser.swagger.Info.Description = value
		case "@description.markdown":
			commentInfo, err := getMarkdownForTag("api", parser.markdownFileDir)
			if err != nil {
				return err
			}
			parser.swagger.Info.Description = string(commentInfo)
		case "@termsofservice":
			parser.swagger.Info.TermsOfService = value
		case "@contact.name":
			parser.swagger.Info.Contact.Name = value
		case "@contact.email":
			parser.swagger.Info.Contact.Email = value
		case "@contact.url":
			parser.swagger.Info.Contact.URL = value
		case "@license.name":
			parser.swagger.Info.License = initIfEmpty(parser.swagger.Info.License)
			parser.swagger.Info.License.Name = value
		case "@license.url":
			parser.swagger.Info.License = initIfEmpty(parser.swagger.Info.License)
			parser.swagger.Info.License.URL = value
		case "@host":
			parser.swagger.Host = value
		case "@basepath":
			parser.swagger.BasePath = value
		case acceptAttr:
			err := parser.ParseAcceptComment(value)
			if err != nil {
				return err
			}
		case produceAttr:
			err := parser.ParseProduceComment(value)
			if err != nil {
				return err
			}
		case "@schemes":
			parser.swagger.Schemes = getSchemes(commentLine)
		case "@tag.name":
			parser.swagger.Tags = append(parser.swagger.Tags, spec.Tag{
				TagProps: spec.TagProps{
					Name: value,
				},
			})
		case "@tag.description":
			tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
			tag.TagProps.Description = value
			replaceLastTag(parser.swagger.Tags, tag)
		case "@tag.description.markdown":
			tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
			commentInfo, err := getMarkdownForTag(tag.TagProps.Name, parser.markdownFileDir)
			if err != nil {
				return err
			}
			tag.TagProps.Description = string(commentInfo)
			replaceLastTag(parser.swagger.Tags, tag)
		case "@tag.docs.url":
			tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
			tag.TagProps.ExternalDocs = &spec.ExternalDocumentation{
				URL: value,
			}
			replaceLastTag(parser.swagger.Tags, tag)
		case "@tag.docs.description":
			tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
			if tag.TagProps.ExternalDocs == nil {
				return fmt.Errorf("%s needs to come after a @tags.docs.url", attribute)
			}
			tag.TagProps.ExternalDocs.Description = value
			replaceLastTag(parser.swagger.Tags, tag)
		case "@securitydefinitions.basic":
			parser.swagger.SecurityDefinitions[value] = spec.BasicAuth()
		case "@securitydefinitions.apikey":
			attrMap, _, _, err := parseSecAttr(attribute, []string{"@in", "@name"}, comments[i+1:])
			if err != nil {
				return err
			}
			parser.swagger.SecurityDefinitions[value] = spec.APIKeyAuth(attrMap["@name"], attrMap["@in"])
		case "@securitydefinitions.oauth2.application":
			attrMap, scopes, extensions, err := parseSecAttr(attribute, []string{"@tokenurl"}, comments[i+1:])
			if err != nil {
				return err
			}
			parser.swagger.SecurityDefinitions[value] = secOAuth2Application(attrMap["@tokenurl"], scopes, extensions)
		case "@securitydefinitions.oauth2.implicit":
			attrs, scopes, ext, err := parseSecAttr(attribute, []string{"@authorizationurl"}, comments[i+1:])
			if err != nil {
				return err
			}
			parser.swagger.SecurityDefinitions[value] = secOAuth2Implicit(attrs["@authorizationurl"], scopes, ext)
		case "@securitydefinitions.oauth2.password":
			attrs, scopes, ext, err := parseSecAttr(attribute, []string{"@tokenurl"}, comments[i+1:])
			if err != nil {
				return err
			}
			parser.swagger.SecurityDefinitions[value] = secOAuth2Password(attrs["@tokenurl"], scopes, ext)
		case "@securitydefinitions.oauth2.accesscode":
			attrs, scopes, ext, err := parseSecAttr(attribute, []string{"@tokenurl", "@authorizationurl"}, comments[i+1:])
			if err != nil {
				return err
			}
			parser.swagger.SecurityDefinitions[value] = secOAuth2AccessToken(attrs["@authorizationurl"], attrs["@tokenurl"], scopes, ext)
		case "@query.collection.format":
			parser.collectionFormatInQuery = value
		default:
			prefixExtension := "@x-"
			// Prefix extension + 1 char + 1 space  + 1 char
			if len(attribute) > 5 && attribute[:len(prefixExtension)] == prefixExtension {
				extExistsInSecurityDef := false
				// for each security definition
				for _, v := range parser.swagger.SecurityDefinitions {
					// check if extension exists
					_, extExistsInSecurityDef = v.VendorExtensible.Extensions.GetString(attribute[1:])
					// if it exists in at least one, then we stop iterating
					if extExistsInSecurityDef {
						break
					}
				}
				// if it is present on security def, don't add it again
				if extExistsInSecurityDef {
					break
				}

				var valueJSON interface{}
				split := strings.SplitAfter(commentLine, attribute+" ")
				if len(split) < 2 {
					return fmt.Errorf("annotation %s need a value", attribute)
				}
				extensionName := "x-" + strings.SplitAfter(attribute, prefixExtension)[1]
				err := json.Unmarshal([]byte(split[1]), &valueJSON)
				if err != nil {
					return fmt.Errorf("annotation %s need a valid json value", attribute)
				}

				if strings.Contains(extensionName, "logo") {
					parser.swagger.Info.Extensions.Add(extensionName, valueJSON)
				} else {
					if parser.swagger.Extensions == nil {
						parser.swagger.Extensions = make(map[string]interface{})
					}
					parser.swagger.Extensions[attribute[1:]] = valueJSON
				}
			}
		}
		previousAttribute = attribute
	}

	return nil
}

// ParseAcceptComment parses comment for given `accept` comment string.
func (parser *Parser) ParseAcceptComment(commentLine string) error {
	return parseMimeTypeList(commentLine, &parser.swagger.Consumes, "%v accept type can't be accepted")
}

// ParseProduceComment parses comment for given `produce` comment string.
func (parser *Parser) ParseProduceComment(commentLine string) error {
	return parseMimeTypeList(commentLine, &parser.swagger.Produces, "%v produce type can't be accepted")
}

func isGeneralAPIComment(comments []string) bool {
	for _, commentLine := range comments {
		attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
		switch attribute {
		// The @summary, @router, @success, @failure annotation belongs to Operation
		case summaryAttr, routerAttr, successAttr, failureAttr, responseAttr:
			return false
		}
	}

	return true
}

func parseSecAttr(context string, search []string, lines []string) (map[string]string, map[string]string, map[string]interface{}, error) {
	attrMap := map[string]string{}
	scopes := map[string]string{}
	extensions := map[string]interface{}{}
	for _, v := range lines {
		securityAttr := strings.ToLower(strings.Split(v, " ")[0])
		for _, findterm := range search {
			if securityAttr == findterm {
				attrMap[securityAttr] = strings.TrimSpace(v[len(securityAttr):])

				continue
			}
		}
		isExists, err := isExistsScope(securityAttr)
		if err != nil {
			return nil, nil, nil, err
		}
		if isExists {
			scopes[securityAttr[len(scopeAttrPrefix):]] = v[len(securityAttr):]
		}
		if strings.HasPrefix(securityAttr, "@x-") {
			// Add the custom attribute without the @
			extensions[securityAttr[1:]] = strings.TrimSpace(v[len(securityAttr):])
		}
		// next securityDefinitions
		if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
			break
		}
	}

	if len(attrMap) != len(search) {
		return nil, nil, nil, fmt.Errorf("%s is %v required", context, search)
	}

	return attrMap, scopes, extensions, nil
}

func secOAuth2Application(tokenURL string, scopes map[string]string,
	extensions map[string]interface{}) *spec.SecurityScheme {
	securityScheme := spec.OAuth2Application(tokenURL)
	securityScheme.VendorExtensible.Extensions = handleSecuritySchemaExtensions(extensions)
	for scope, description := range scopes {
		securityScheme.AddScope(scope, description)
	}

	return securityScheme
}

func secOAuth2Implicit(authorizationURL string, scopes map[string]string,
	extensions map[string]interface{}) *spec.SecurityScheme {
	securityScheme := spec.OAuth2Implicit(authorizationURL)
	securityScheme.VendorExtensible.Extensions = handleSecuritySchemaExtensions(extensions)
	for scope, description := range scopes {
		securityScheme.AddScope(scope, description)
	}

	return securityScheme
}

func secOAuth2Password(tokenURL string, scopes map[string]string,
	extensions map[string]interface{}) *spec.SecurityScheme {
	securityScheme := spec.OAuth2Password(tokenURL)
	securityScheme.VendorExtensible.Extensions = handleSecuritySchemaExtensions(extensions)
	for scope, description := range scopes {
		securityScheme.AddScope(scope, description)
	}

	return securityScheme
}

func secOAuth2AccessToken(authorizationURL, tokenURL string,
	scopes map[string]string, extensions map[string]interface{}) *spec.SecurityScheme {
	securityScheme := spec.OAuth2AccessToken(authorizationURL, tokenURL)
	securityScheme.VendorExtensible.Extensions = handleSecuritySchemaExtensions(extensions)
	for scope, description := range scopes {
		securityScheme.AddScope(scope, description)
	}

	return securityScheme
}

func handleSecuritySchemaExtensions(providedExtensions map[string]interface{}) spec.Extensions {
	var extensions spec.Extensions
	if len(providedExtensions) > 0 {
		extensions = make(map[string]interface{}, len(providedExtensions))
		for key, value := range providedExtensions {
			extensions[key] = value
		}
	}

	return extensions
}

func getMarkdownForTag(tagName string, dirPath string) ([]byte, error) {
	filesInfos, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range filesInfos {
		if fileInfo.IsDir() {
			continue
		}
		fileName := fileInfo.Name()

		if !strings.Contains(fileName, ".md") {
			continue
		}

		if strings.Contains(fileName, tagName) {
			fullPath := filepath.Join(dirPath, fileName)
			commentInfo, err := ioutil.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("Failed to read markdown file %s error: %s ", fullPath, err)
			}

			return commentInfo, nil
		}
	}

	return nil, fmt.Errorf("Unable to find markdown file for tag %s in the given directory", tagName)
}

func isExistsScope(scope string) (bool, error) {
	s := strings.Fields(scope)
	for _, v := range s {
		if strings.Contains(v, scopeAttrPrefix) {
			if strings.Contains(v, ",") {
				return false, fmt.Errorf("@scope can't use comma(,) get=" + v)
			}
		}
	}

	return strings.Contains(scope, scopeAttrPrefix), nil
}

// getSchemes parses swagger schemes for given commentLine.
func getSchemes(commentLine string) []string {
	attribute := strings.ToLower(strings.Split(commentLine, " ")[0])

	return strings.Split(strings.TrimSpace(commentLine[len(attribute):]), " ")
}

// ParseRouterAPIInfo parses router api info for given astFile.
func (parser *Parser) ParseRouterAPIInfo(fileName string, astFile *ast.File) error {
	for _, astDescription := range astFile.Decls {
		astDeclaration, ok := astDescription.(*ast.FuncDecl)
		if ok && astDeclaration.Doc != nil && astDeclaration.Doc.List != nil {
			// for per 'function' comment, create a new 'Operation' object
			operation := NewOperation(parser, SetCodeExampleFilesDirectory(parser.codeExampleFilesDir))
			for _, comment := range astDeclaration.Doc.List {
				err := operation.ParseComment(comment.Text, astFile)
				if err != nil {
					return fmt.Errorf("ParseComment error in file %s :%+v", fileName, err)
				}
			}

			for _, routeProperties := range operation.RouterProperties {
				var pathItem spec.PathItem
				var ok bool

				pathItem, ok = parser.swagger.Paths.Paths[routeProperties.Path]
				if !ok {
					pathItem = spec.PathItem{}
				}

				op := refRouteMethodOp(&pathItem, routeProperties.HTTPMethod)

				// check if we already have a operation for this path and method
				if *op != nil {
					err := fmt.Errorf("route %s %s is declared multiple times", routeProperties.HTTPMethod, routeProperties.Path)
					if parser.Strict {
						return err
					}
					parser.debug.Printf("warning: %s\n", err)
				}

				*op = &operation.Operation

				parser.swagger.Paths.Paths[routeProperties.Path] = pathItem
			}
		}
	}

	return nil
}

func refRouteMethodOp(item *spec.PathItem, method string) (op **spec.Operation) {
	switch method {
	case http.MethodGet:
		op = &item.Get
	case http.MethodPost:
		op = &item.Post
	case http.MethodDelete:
		op = &item.Delete
	case http.MethodPut:
		op = &item.Put
	case http.MethodPatch:
		op = &item.Patch
	case http.MethodHead:
		op = &item.Head
	case http.MethodOptions:
		op = &item.Options
	}
	return
}

func convertFromSpecificToPrimitive(typeName string) (string, error) {
	name := typeName
	if strings.ContainsRune(name, '.') {
		name = strings.Split(name, ".")[1]
	}
	switch strings.ToUpper(name) {
	case "TIME", "OBJECTID", "UUID":
		return STRING, nil
	case "DECIMAL":
		return NUMBER, nil
	}

	return typeName, ErrFailedConvertPrimitiveType
}

func (parser *Parser) getTypeSchema(typeName string, file *ast.File, ref bool) (*spec.Schema, error) {
	if IsGolangPrimitiveType(typeName) {
		return PrimitiveSchema(TransToValidSchemeType(typeName)), nil
	}

	schemaType, err := convertFromSpecificToPrimitive(typeName)
	if err == nil {
		return PrimitiveSchema(schemaType), nil
	}

	typeSpecDef := parser.packages.FindTypeSpec(typeName, file, parser.ParseDependency)
	if typeSpecDef == nil {
		return nil, fmt.Errorf("cannot find type definition: %s", typeName)
	}

	if override, ok := parser.Overrides[typeSpecDef.FullPath()]; ok {
		if override == "" {
			parser.debug.Printf("Override detected for %s: ignoring", typeSpecDef.FullPath())
			return nil, ErrSkippedField
		}

		parser.debug.Printf("Override detected for %s: using %s instead", typeSpecDef.FullPath(), override)

		separator := strings.LastIndex(override, ".")
		if separator == -1 {
			// treat as a swaggertype tag
			parts := strings.Split(override, ",")
			return BuildCustomSchema(parts)
		}

		typeSpecDef = parser.packages.findTypeSpec(override[0:separator], override[separator+1:])
	}

	schema, ok := parser.parsedSchemas[typeSpecDef]
	if !ok {
		var err error
		schema, err = parser.ParseDefinition(typeSpecDef)
		if err != nil {
			if err == ErrRecursiveParseStruct && ref {
				return parser.getRefTypeSchema(typeSpecDef, schema), nil
			}

			return nil, err
		}
	}

	if ref && len(schema.Schema.Type) > 0 && schema.Schema.Type[0] == OBJECT {
		return parser.getRefTypeSchema(typeSpecDef, schema), nil
	}

	return schema.Schema, nil
}

func (parser *Parser) renameRefSchemas() {
	if len(parser.toBeRenamedSchemas) == 0 {
		return
	}

	// rename schemas in swagger.Definitions
	for name, pkgPath := range parser.toBeRenamedSchemas {
		if schema, ok := parser.swagger.Definitions[name]; ok {
			delete(parser.swagger.Definitions, name)
			name = parser.renameSchema(name, pkgPath)
			parser.swagger.Definitions[name] = schema
		}
	}

	// rename URLs if match
	for _, refURL := range parser.toBeRenamedRefURLs {
		parts := strings.Split(refURL.Fragment, "/")
		name := parts[len(parts)-1]
		if pkgPath, ok := parser.toBeRenamedSchemas[name]; ok {
			parts[len(parts)-1] = parser.renameSchema(name, pkgPath)
			refURL.Fragment = strings.Join(parts, "/")
		}
	}
}

func (parser *Parser) renameSchema(name, pkgPath string) string {
	parts := strings.Split(name, ".")
	name = fullTypeName(pkgPath, parts[len(parts)-1])
	name = strings.ReplaceAll(name, "/", "_")

	return name
}

func (parser *Parser) getRefTypeSchema(typeSpecDef *TypeSpecDef, schema *Schema) *spec.Schema {
	_, ok := parser.outputSchemas[typeSpecDef]
	if !ok {
		existSchema, ok := parser.existSchemaNames[schema.Name]
		if ok {
			// store the first one to be renamed after parsing over
			_, ok = parser.toBeRenamedSchemas[existSchema.Name]
			if !ok {
				parser.toBeRenamedSchemas[existSchema.Name] = existSchema.PkgPath
			}
			// rename not the first one
			schema.Name = parser.renameSchema(schema.Name, schema.PkgPath)
		} else {
			parser.existSchemaNames[schema.Name] = schema
		}
		parser.swagger.Definitions[schema.Name] = spec.Schema{}

		if schema.Schema != nil {
			parser.swagger.Definitions[schema.Name] = *schema.Schema
		}

		parser.outputSchemas[typeSpecDef] = schema
	}

	refSchema := RefSchema(schema.Name)
	// store every URL
	parser.toBeRenamedRefURLs = append(parser.toBeRenamedRefURLs, refSchema.Ref.GetURL())

	return refSchema
}

func (parser *Parser) isInStructStack(typeSpecDef *TypeSpecDef) bool {
	for _, specDef := range parser.structStack {
		if typeSpecDef == specDef {
			return true
		}
	}

	return false
}

// ParseDefinition parses given type spec that corresponds to the type under
// given name and package, and populates swagger schema definitions registry
// with a schema for the given type
func (parser *Parser) ParseDefinition(typeSpecDef *TypeSpecDef) (*Schema, error) {
	typeName := typeSpecDef.FullName()
	refTypeName := TypeDocName(typeName, typeSpecDef.TypeSpec)

	schema, ok := parser.parsedSchemas[typeSpecDef]
	if ok {
		parser.debug.Printf("Skipping '%s', already parsed.", typeName)

		return schema, nil
	}

	if parser.isInStructStack(typeSpecDef) {
		parser.debug.Printf("Skipping '%s', recursion detected.", typeName)

		return &Schema{
				Name:    refTypeName,
				PkgPath: typeSpecDef.PkgPath,
				Schema:  PrimitiveSchema(OBJECT),
			},
			ErrRecursiveParseStruct
	}
	parser.structStack = append(parser.structStack, typeSpecDef)

	parser.debug.Printf("Generating %s", typeName)

	definition, err := parser.parseTypeExpr(typeSpecDef.File, typeSpecDef.TypeSpec.Type, false)
	if err != nil {
		return nil, err
	}

	if definition.Description == "" {
		fillDefinitionDescription(definition, typeSpecDef.File, typeSpecDef)
	}

	s := Schema{
		Name:    refTypeName,
		PkgPath: typeSpecDef.PkgPath,
		Schema:  definition,
	}
	parser.parsedSchemas[typeSpecDef] = &s

	// update an empty schema as a result of recursion
	s2, ok := parser.outputSchemas[typeSpecDef]
	if ok {
		parser.swagger.Definitions[s2.Name] = *definition
	}

	return &s, nil
}

func fullTypeName(pkgName, typeName string) string {
	if pkgName != "" {
		return pkgName + "." + typeName
	}

	return typeName
}

// fillDefinitionDescription additionally fills fields in definition (spec.Schema)
// TODO: If .go file contains many types, it may work for a long time
func fillDefinitionDescription(definition *spec.Schema, file *ast.File, typeSpecDef *TypeSpecDef) {
	for _, astDeclaration := range file.Decls {
		generalDeclaration, ok := astDeclaration.(*ast.GenDecl)
		if !ok || generalDeclaration.Tok != token.TYPE {
			continue
		}

		for _, astSpec := range generalDeclaration.Specs {
			typeSpec, ok := astSpec.(*ast.TypeSpec)
			if !ok || typeSpec != typeSpecDef.TypeSpec {
				continue
			}

			definition.Description =
				extractDeclarationDescription(typeSpec.Doc, typeSpec.Comment, generalDeclaration.Doc)
		}
	}
}

// extractDeclarationDescription gets first description
// from attribute descriptionAttr in commentGroups (ast.CommentGroup)
func extractDeclarationDescription(commentGroups ...*ast.CommentGroup) string {
	var description string

	for _, commentGroup := range commentGroups {
		if commentGroup == nil {
			continue
		}

		isHandlingDescription := false
		for _, comment := range commentGroup.List {
			commentText := strings.TrimSpace(strings.TrimLeft(comment.Text, "/"))
			attribute := strings.Split(commentText, " ")[0]
			if strings.ToLower(attribute) != descriptionAttr {
				if !isHandlingDescription {
					continue
				}
				break
			}

			isHandlingDescription = true
			description += " " + strings.TrimSpace(commentText[len(attribute):])
		}
	}

	return strings.TrimLeft(description, " ")
}

// parseTypeExpr parses given type expression that corresponds to the type under
// given name and package, and returns swagger schema for it.
func (parser *Parser) parseTypeExpr(file *ast.File, typeExpr ast.Expr, ref bool) (*spec.Schema, error) {
	switch expr := typeExpr.(type) {
	// type Foo interface{}
	case *ast.InterfaceType:
		return &spec.Schema{}, nil

	// type Foo struct {...}
	case *ast.StructType:
		return parser.parseStruct(file, expr.Fields)

	// type Foo Baz
	case *ast.Ident:
		return parser.getTypeSchema(expr.Name, file, ref)

	// type Foo *Baz
	case *ast.StarExpr:
		return parser.parseTypeExpr(file, expr.X, ref)

	// type Foo pkg.Bar
	case *ast.SelectorExpr:
		if xIdent, ok := expr.X.(*ast.Ident); ok {
			return parser.getTypeSchema(fullTypeName(xIdent.Name, expr.Sel.Name), file, ref)
		}
	// type Foo []Baz
	case *ast.ArrayType:
		itemSchema, err := parser.parseTypeExpr(file, expr.Elt, true)
		if err != nil {
			return nil, err
		}

		return spec.ArrayProperty(itemSchema), nil
	// type Foo map[string]Bar
	case *ast.MapType:
		if _, ok := expr.Value.(*ast.InterfaceType); ok {
			return spec.MapProperty(nil), nil
		}
		schema, err := parser.parseTypeExpr(file, expr.Value, true)
		if err != nil {
			return nil, err
		}

		return spec.MapProperty(schema), nil

	case *ast.FuncType:
		return nil, ErrFuncTypeField
	// ...
	default:
		parser.debug.Printf("Type definition of type '%T' is not supported yet. Using 'object' instead.\n", typeExpr)
	}

	return PrimitiveSchema(OBJECT), nil
}

func (parser *Parser) parseStruct(file *ast.File, fields *ast.FieldList) (*spec.Schema, error) {
	required := make([]string, 0)
	properties := make(map[string]spec.Schema)
	for _, field := range fields.List {
		fieldProps, requiredFromAnon, err := parser.parseStructField(file, field)
		if err != nil {
			if err == ErrFuncTypeField || err == ErrSkippedField {
				continue
			}

			return nil, err
		}
		if len(fieldProps) == 0 {
			continue
		}
		required = append(required, requiredFromAnon...)
		for k, v := range fieldProps {
			properties[k] = v
		}
	}

	sort.Strings(required)

	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       []string{OBJECT},
			Properties: properties,
			Required:   required,
		},
	}, nil
}

func (parser *Parser) parseStructField(file *ast.File, field *ast.Field) (map[string]spec.Schema, []string, error) {
	if field.Names == nil {
		if field.Tag != nil {
			skip, ok := reflect.StructTag(strings.ReplaceAll(field.Tag.Value, "`", "")).Lookup("swaggerignore")
			if ok && strings.EqualFold(skip, "true") {
				return nil, nil, nil
			}
		}

		typeName, err := getFieldType(field.Type)
		if err != nil {
			return nil, nil, err
		}
		schema, err := parser.getTypeSchema(typeName, file, false)
		if err != nil {
			return nil, nil, err
		}
		if len(schema.Type) > 0 && schema.Type[0] == OBJECT {
			if len(schema.Properties) == 0 {
				return nil, nil, nil
			}

			properties := map[string]spec.Schema{}
			for k, v := range schema.Properties {
				properties[k] = v
			}

			return properties, schema.SchemaProps.Required, nil
		}

		// for alias type of non-struct types ,such as array,map, etc. ignore field tag.
		return map[string]spec.Schema{typeName: *schema}, nil, nil
	}

	ps := parser.fieldParserFactory(parser, field)

	ok, err := ps.ShouldSkip()
	if err != nil {
		return nil, nil, err
	}
	if ok {
		return nil, nil, nil
	}

	fieldName, err := ps.FieldName()
	if err != nil {
		return nil, nil, err
	}

	schema, err := ps.CustomSchema()
	if err != nil {
		return nil, nil, err
	}
	if schema == nil {
		typeName, err := getFieldType(field.Type)
		if err == nil {
			// named type
			schema, err = parser.getTypeSchema(typeName, file, true)
		} else {
			// unnamed type
			schema, err = parser.parseTypeExpr(file, field.Type, false)
		}
		if err != nil {
			return nil, nil, err
		}
	}

	err = ps.ComplementSchema(schema)
	if err != nil {
		return nil, nil, err
	}

	var tagRequired []string
	required, err := ps.IsRequired()
	if err != nil {
		return nil, nil, err
	}
	if required {
		tagRequired = append(tagRequired, fieldName)
	}

	return map[string]spec.Schema{fieldName: *schema}, tagRequired, nil
}

func getFieldType(field ast.Expr) (string, error) {
	switch fieldType := field.(type) {
	case *ast.Ident:
		return fieldType.Name, nil
	case *ast.SelectorExpr:
		packageName, err := getFieldType(fieldType.X)
		if err != nil {
			return "", err
		}

		return fullTypeName(packageName, fieldType.Sel.Name), nil

	case *ast.StarExpr:
		fullName, err := getFieldType(fieldType.X)
		if err != nil {
			return "", err
		}

		return fullName, nil
	default:
		return "", fmt.Errorf("unknown field type %#v", field)
	}
}

// GetSchemaTypePath get path of schema type.
func (parser *Parser) GetSchemaTypePath(schema *spec.Schema, depth int) []string {
	if schema == nil || depth == 0 {
		return nil
	}
	name := schema.Ref.String()
	if name != "" {
		if pos := strings.LastIndexByte(name, '/'); pos >= 0 {
			name = name[pos+1:]
			if schema, ok := parser.swagger.Definitions[name]; ok {
				return parser.GetSchemaTypePath(&schema, depth)
			}
		}

		return nil
	}
	if len(schema.Type) > 0 {
		switch schema.Type[0] {
		case ARRAY:
			depth--
			s := []string{schema.Type[0]}

			return append(s, parser.GetSchemaTypePath(schema.Items.Schema, depth)...)
		case OBJECT:
			if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
				// for map
				depth--
				s := []string{schema.Type[0]}

				return append(s, parser.GetSchemaTypePath(schema.AdditionalProperties.Schema, depth)...)
			}
		}

		return []string{schema.Type[0]}
	}

	return []string{ANY}
}

func replaceLastTag(slice []spec.Tag, element spec.Tag) {
	slice = append(slice[:len(slice)-1], element)
}

// defineTypeOfExample example value define the type (object and array unsupported)
func defineTypeOfExample(schemaType, arrayType, exampleValue string) (interface{}, error) {
	switch schemaType {
	case STRING:
		return exampleValue, nil
	case NUMBER:
		v, err := strconv.ParseFloat(exampleValue, 64)
		if err != nil {
			return nil, fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err)
		}

		return v, nil
	case INTEGER:
		v, err := strconv.Atoi(exampleValue)
		if err != nil {
			return nil, fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err)
		}

		return v, nil
	case BOOLEAN:
		v, err := strconv.ParseBool(exampleValue)
		if err != nil {
			return nil, fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err)
		}

		return v, nil
	case ARRAY:
		values := strings.Split(exampleValue, ",")
		result := make([]interface{}, 0)
		for _, value := range values {
			v, err := defineTypeOfExample(arrayType, "", value)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}

		return result, nil
	case OBJECT:
		if arrayType == "" {
			return nil, fmt.Errorf("%s is unsupported type in example value `%s`", schemaType, exampleValue)
		}

		values := strings.Split(exampleValue, ",")
		result := map[string]interface{}{}
		for _, value := range values {
			mapData := strings.Split(value, ":")

			if len(mapData) == 2 {
				v, err := defineTypeOfExample(arrayType, "", mapData[1])
				if err != nil {
					return nil, err
				}
				result[mapData[0]] = v
			} else {
				return nil, fmt.Errorf("example value %s should format: key:value", exampleValue)
			}
		}

		return result, nil
	}

	return nil, fmt.Errorf("%s is unsupported type in example value %s", schemaType, exampleValue)
}

// GetAllGoFileInfo gets all Go source files information for given searchDir.
func (parser *Parser) getAllGoFileInfo(packageDir, searchDir string) error {
	return filepath.Walk(searchDir, func(path string, f os.FileInfo, _ error) error {
		if err := parser.Skip(path, f); err != nil {
			return err
		} else if f.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(searchDir, path)
		if err != nil {
			return err
		}

		return parser.parseFile(filepath.ToSlash(filepath.Dir(filepath.Clean(filepath.Join(packageDir, relPath)))), path, nil)
	})
}

func (parser *Parser) getAllGoFileInfoFromDeps(pkg *depth.Pkg) error {
	ignoreInternal := pkg.Internal && !parser.ParseInternal
	if ignoreInternal || !pkg.Resolved { // ignored internal and not resolved dependencies
		return nil
	}

	// Skip cgo
	if pkg.Raw == nil && pkg.Name == "C" {
		return nil
	}
	srcDir := pkg.Raw.Dir
	files, err := ioutil.ReadDir(srcDir) // only parsing files in the dir(don't contain sub dir files)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(srcDir, f.Name())
		if err := parser.parseFile(pkg.Name, path, nil); err != nil {
			return err
		}
	}

	for i := 0; i < len(pkg.Deps); i++ {
		if err := parser.getAllGoFileInfoFromDeps(&pkg.Deps[i]); err != nil {
			return err
		}
	}

	return nil
}

func (parser *Parser) parseFile(packageDir, path string, src interface{}) error {
	if strings.HasSuffix(strings.ToLower(path), "_test.go") || filepath.Ext(path) != ".go" {
		return nil
	}

	// positions are relative to FileSet
	astFile, err := goparser.ParseFile(token.NewFileSet(), path, src, goparser.ParseComments)
	if err != nil {
		return fmt.Errorf("ParseFile error:%+v", err)
	}

	err = parser.packages.CollectAstFile(packageDir, path, astFile)
	if err != nil {
		return err
	}

	return nil
}

func (parser *Parser) checkOperationIDUniqueness() error {
	// operationsIds contains all operationId annotations to check it's unique
	operationsIds := make(map[string]string)

	for path, item := range parser.swagger.Paths.Paths {
		var method, id string
		for method = range allMethod {
			op := refRouteMethodOp(&item, method)
			if *op != nil {
				id = (**op).ID
				break
			}
		}
		if id == "" {
			continue
		}

		current := fmt.Sprintf("%s %s", method, path)
		previous, ok := operationsIds[id]
		if ok {
			return fmt.Errorf(
				"duplicated @id annotation '%s' found in '%s', previously declared in: '%s'",
				id, current, previous)
		}
		operationsIds[id] = current
	}

	return nil
}

// Skip returns filepath.SkipDir error if match vendor and hidden folder.
func (parser *Parser) Skip(path string, f os.FileInfo) error {
	return walkWith(parser.excludes, parser.ParseVendor)(path, f)
}

func walkWith(excludes map[string]struct{}, parseVendor bool) func(path string, fileInfo os.FileInfo) error {
	return func(path string, f os.FileInfo) error {
		if f.IsDir() {
			if !parseVendor && f.Name() == "vendor" || // ignore "vendor"
				f.Name() == "docs" || // exclude docs
				len(f.Name()) > 1 && f.Name()[0] == '.' { // exclude all hidden folder
				return filepath.SkipDir
			}

			if excludes != nil {
				if _, ok := excludes[path]; ok {
					return filepath.SkipDir
				}
			}
		}

		return nil
	}
}

// GetSwagger returns *spec.Swagger which is the root document object for the API specification.
func (parser *Parser) GetSwagger() *spec.Swagger {
	return parser.swagger
}

// addTestType just for tests.
func (parser *Parser) addTestType(typename string) {
	typeDef := &TypeSpecDef{}
	parser.packages.uniqueDefinitions[typename] = typeDef
	parser.parsedSchemas[typeDef] = &Schema{
		PkgPath: "",
		Name:    typename,
		Schema:  PrimitiveSchema(OBJECT),
	}
}
