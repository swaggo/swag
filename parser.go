package swag

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/KyleBanks/depth"
	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
)

const (
	// CamelCase indicates using CamelCase strategy for struct field.
	CamelCase = "camelcase"

	// PascalCase indicates using PascalCase strategy for struct field.
	PascalCase = "pascalcase"

	// SnakeCase indicates using SnakeCase strategy for struct field.
	SnakeCase = "snakecase"
)

// Parser implements a parser for Go source files.
type Parser struct {
	// swagger represents the root document object for the API specification
	swagger *spec.Swagger

	// files is a map that stores map[real_go_file_path][astFile]
	files map[string]*ast.File

	// TypeDefinitions is a map that stores [package name][type name][*ast.TypeSpec]
	TypeDefinitions map[string]map[string]*ast.TypeSpec

	// ImportAliases is map that stores [import name][import package name][*ast.ImportSpec]
	ImportAliases map[string]map[string]*ast.ImportSpec

	// CustomPrimitiveTypes is a map that stores custom primitive types to actual golang types [type name][string]
	CustomPrimitiveTypes map[string]string

	// registerTypes is a map that stores [refTypeName][*ast.TypeSpec]
	registerTypes map[string]*ast.TypeSpec

	PropNamingStrategy string

	ParseVendor bool

	// ParseDependencies whether swag should be parse outside dependency folder
	ParseDependency bool

	// structStack stores full names of the structures that were already parsed or are being parsed now
	structStack []string

	// markdownFileDir holds the path to the folder, where markdown files are stored
	markdownFileDir string
}

// New creates a new Parser with default properties.
func New(options ...func(*Parser)) *Parser {
	parser := &Parser{
		swagger: &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Info: &spec.Info{
					InfoProps: spec.InfoProps{
						Contact: &spec.ContactInfo{},
						License: &spec.License{},
					},
				},
				Paths: &spec.Paths{
					Paths: make(map[string]spec.PathItem),
				},
				Definitions: make(map[string]spec.Schema),
			},
		},
		files:                make(map[string]*ast.File),
		TypeDefinitions:      make(map[string]map[string]*ast.TypeSpec),
		ImportAliases:        make(map[string]map[string]*ast.ImportSpec),
		CustomPrimitiveTypes: make(map[string]string),
		registerTypes:        make(map[string]*ast.TypeSpec),
	}

	for _, option := range options {
		option(parser)
	}

	return parser
}

// SetMarkdownFileDirectory sets the directory to search for markdownfiles
func SetMarkdownFileDirectory(directoryPath string) func(*Parser) {
	return func(p *Parser) {
		p.markdownFileDir = directoryPath
	}
}

// ParseAPI parses general api info for given searchDir and mainAPIFile
func (parser *Parser) ParseAPI(searchDir string, mainAPIFile string) error {
	Printf("Generate general API Info, search dir:%s", searchDir)

	if err := parser.getAllGoFileInfo(searchDir); err != nil {
		return err
	}

	var t depth.Tree

	absMainAPIFilePath, err := filepath.Abs(filepath.Join(searchDir, mainAPIFile))
	if err != nil {
		return err
	}

	if parser.ParseDependency {
		pkgName, err := getPkgName(path.Dir(absMainAPIFilePath))
		if err != nil {
			return err
		}
		if err := t.Resolve(pkgName); err != nil {
			return fmt.Errorf("pkg %s cannot find all dependencies, %s", pkgName, err)
		}
		for i := 0; i < len(t.Root.Deps); i++ {
			if err := parser.getAllGoFileInfoFromDeps(&t.Root.Deps[i]); err != nil {
				return err
			}
		}
	}

	if err := parser.ParseGeneralAPIInfo(absMainAPIFilePath); err != nil {
		return err
	}

	for _, astFile := range parser.files {
		parser.ParseType(astFile)
	}

	for fileName, astFile := range parser.files {
		if err := parser.ParseRouterAPIInfo(fileName, astFile); err != nil {
			return err
		}
	}

	return parser.parseDefinitions()
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

// ParseGeneralAPIInfo parses general api info for given mainAPIFile path
func (parser *Parser) ParseGeneralAPIInfo(mainAPIFile string) error {
	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, mainAPIFile, nil, goparser.ParseComments)
	if err != nil {
		return fmt.Errorf("cannot parse source files %s: %s", mainAPIFile, err)
	}

	parser.swagger.Swagger = "2.0"
	securityMap := map[string]*spec.SecurityScheme{}

	for _, comment := range fileTree.Comments {
		if !isGeneralAPIComment(comment) {
			continue
		}
		comments := strings.Split(comment.Text(), "\n")
		previousAttribute := ""
		// parsing classic meta data model
		for i, commentLine := range comments {
			attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
			value := strings.TrimSpace(commentLine[len(attribute):])
			multilineBlock := false
			if previousAttribute == attribute {
				multilineBlock = true
			}
			switch attribute {
			case "@version":
				parser.swagger.Info.Version = value
			case "@title":
				parser.swagger.Info.Title = value
			case "@description":
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
				parser.swagger.Info.License.Name = value
			case "@license.url":
				parser.swagger.Info.License.URL = value
			case "@host":
				parser.swagger.Host = value
			case "@basepath":
				parser.swagger.BasePath = value
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
				securityMap[value] = spec.BasicAuth()
			case "@securitydefinitions.apikey":
				attrMap, _, err := extractSecurityAttribute(attribute, []string{"@in", "@name"}, comments[i+1:])
				if err != nil {
					return err
				}
				securityMap[value] = spec.APIKeyAuth(attrMap["@name"], attrMap["@in"])
			case "@securitydefinitions.oauth2.application":
				attrMap, scopes, err := extractSecurityAttribute(attribute, []string{"@tokenurl"}, comments[i+1:])
				if err != nil {
					return err
				}
				securityMap[value] = securitySchemeOAuth2Application(attrMap["@tokenurl"], scopes)
			case "@securitydefinitions.oauth2.implicit":
				attrMap, scopes, err := extractSecurityAttribute(attribute, []string{"@authorizationurl"}, comments[i+1:])
				if err != nil {
					return err
				}
				securityMap[value] = securitySchemeOAuth2Implicit(attrMap["@authorizationurl"], scopes)
			case "@securitydefinitions.oauth2.password":
				attrMap, scopes, err := extractSecurityAttribute(attribute, []string{"@tokenurl"}, comments[i+1:])
				if err != nil {
					return err
				}
				securityMap[value] = securitySchemeOAuth2Password(attrMap["@tokenurl"], scopes)
			case "@securitydefinitions.oauth2.accesscode":
				attrMap, scopes, err := extractSecurityAttribute(attribute, []string{"@tokenurl", "@authorizationurl"}, comments[i+1:])
				if err != nil {
					return err
				}
				securityMap[value] = securitySchemeOAuth2AccessToken(attrMap["@authorizationurl"], attrMap["@tokenurl"], scopes)

			default:
				prefixExtension := "@x-"
				if len(attribute) > 5 { // Prefix extension + 1 char + 1 space  + 1 char
					if attribute[:len(prefixExtension)] == prefixExtension {
						var valueJSON interface{}
						split := strings.SplitAfter(commentLine, attribute+" ")
						if len(split) < 2 {
							return fmt.Errorf("annotation %s need a value", attribute)
						}
						extensionName := "x-" + strings.SplitAfter(attribute, prefixExtension)[1]
						if err := json.Unmarshal([]byte(split[1]), &valueJSON); err != nil {
							return fmt.Errorf("annotation %s need a valid json value", attribute)
						}
						parser.swagger.AddExtension(extensionName, valueJSON)
					}
				}
			}
			previousAttribute = attribute
		}
	}

	if len(securityMap) > 0 {
		parser.swagger.SecurityDefinitions = securityMap
	}

	return nil
}

func isGeneralAPIComment(comment *ast.CommentGroup) bool {
	for _, commentLine := range strings.Split(comment.Text(), "\n") {
		attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
		switch attribute {
		// The @summary, @router, @success,@failure  annotation belongs to Operation
		case "@summary", "@router", "@success", "@failure":
			return false
		}
	}
	return true
}

func extractSecurityAttribute(context string, search []string, lines []string) (map[string]string, map[string]string, error) {
	attrMap := map[string]string{}
	scopes := map[string]string{}
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
			return nil, nil, err
		}
		if isExists {
			scopScheme, err := getScopeScheme(securityAttr)
			if err != nil {
				return nil, nil, err
			}
			scopes[scopScheme] = v[len(securityAttr):]
		}
		// next securityDefinitions
		if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
			break
		}
	}
	if len(attrMap) != len(search) {
		return nil, nil, fmt.Errorf("%s is %v required", context, search)
	}
	return attrMap, scopes, nil
}

func securitySchemeOAuth2Application(tokenurl string, scopes map[string]string) *spec.SecurityScheme {
	securityScheme := spec.OAuth2Application(tokenurl)
	for scope, description := range scopes {
		securityScheme.AddScope(scope, description)
	}
	return securityScheme
}

func securitySchemeOAuth2Implicit(authorizationurl string, scopes map[string]string) *spec.SecurityScheme {
	securityScheme := spec.OAuth2Implicit(authorizationurl)
	for scope, description := range scopes {
		securityScheme.AddScope(scope, description)
	}
	return securityScheme
}

func securitySchemeOAuth2Password(tokenurl string, scopes map[string]string) *spec.SecurityScheme {
	securityScheme := spec.OAuth2Password(tokenurl)
	for scope, description := range scopes {
		securityScheme.AddScope(scope, description)
	}
	return securityScheme
}

func securitySchemeOAuth2AccessToken(authorizationurl, tokenurl string, scopes map[string]string) *spec.SecurityScheme {
	securityScheme := spec.OAuth2AccessToken(authorizationurl, tokenurl)
	for scope, description := range scopes {
		securityScheme.AddScope(scope, description)
	}
	return securityScheme
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

func getScopeScheme(scope string) (string, error) {
	scopeValue := scope[strings.Index(scope, "@scope."):]
	if scopeValue == "" {
		return "", fmt.Errorf("@scope is empty")
	}
	return scope[len("@scope."):], nil
}

func isExistsScope(scope string) (bool, error) {
	s := strings.Fields(scope)
	for _, v := range s {
		if strings.Contains(v, "@scope.") {
			if strings.Contains(v, ",") {
				return false, fmt.Errorf("@scope can't use comma(,) get=" + v)
			}
		}
	}
	return strings.Contains(scope, "@scope."), nil
}

// getSchemes parses swagger schemes for given commentLine
func getSchemes(commentLine string) []string {
	attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
	return strings.Split(strings.TrimSpace(commentLine[len(attribute):]), " ")
}

// ParseRouterAPIInfo parses router api info for given astFile
func (parser *Parser) ParseRouterAPIInfo(fileName string, astFile *ast.File) error {
	for _, astDescription := range astFile.Decls {
		switch astDeclaration := astDescription.(type) {
		case *ast.FuncDecl:
			if astDeclaration.Doc != nil && astDeclaration.Doc.List != nil {
				operation := NewOperation() //for per 'function' comment, create a new 'Operation' object
				operation.parser = parser
				for _, comment := range astDeclaration.Doc.List {
					if err := operation.ParseComment(comment.Text, astFile); err != nil {
						return fmt.Errorf("ParseComment error in file %s :%+v", fileName, err)
					}
				}
				var pathItem spec.PathItem
				var ok bool

				if pathItem, ok = parser.swagger.Paths.Paths[operation.Path]; !ok {
					pathItem = spec.PathItem{}
				}
				switch strings.ToUpper(operation.HTTPMethod) {
				case http.MethodGet:
					pathItem.Get = &operation.Operation
				case http.MethodPost:
					pathItem.Post = &operation.Operation
				case http.MethodDelete:
					pathItem.Delete = &operation.Operation
				case http.MethodPut:
					pathItem.Put = &operation.Operation
				case http.MethodPatch:
					pathItem.Patch = &operation.Operation
				case http.MethodHead:
					pathItem.Head = &operation.Operation
				case http.MethodOptions:
					pathItem.Options = &operation.Operation
				}

				parser.swagger.Paths.Paths[operation.Path] = pathItem
			}
		}
	}

	return nil
}

// ParseType parses type info for given astFile.
func (parser *Parser) ParseType(astFile *ast.File) {
	if _, ok := parser.TypeDefinitions[astFile.Name.String()]; !ok {
		parser.TypeDefinitions[astFile.Name.String()] = make(map[string]*ast.TypeSpec)
	}

	for _, astDeclaration := range astFile.Decls {
		if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
			for _, astSpec := range generalDeclaration.Specs {
				if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
					typeName := fmt.Sprintf("%v", typeSpec.Type)
					// check if its a custom primitive type
					if IsGolangPrimitiveType(typeName) {
						parser.CustomPrimitiveTypes[typeSpec.Name.String()] = TransToValidSchemeType(typeName)
					} else {
						parser.TypeDefinitions[astFile.Name.String()][typeSpec.Name.String()] = typeSpec
					}

				}
			}
		}
	}

	for _, importSpec := range astFile.Imports {
		if importSpec.Name == nil {
			continue
		}

		alias := importSpec.Name.Name

		if _, ok := parser.ImportAliases[alias]; !ok {
			parser.ImportAliases[alias] = make(map[string]*ast.ImportSpec)
		}

		importParts := strings.Split(strings.Trim(importSpec.Path.Value, "\""), "/")
		importPkgName := importParts[len(importParts)-1]

		parser.ImportAliases[alias][importPkgName] = importSpec
	}
}

func (parser *Parser) isInStructStack(refTypeName string) bool {
	for _, structName := range parser.structStack {
		if refTypeName == structName {
			return true
		}
	}
	return false
}

// parseDefinitions parses Swagger Api definitions.
func (parser *Parser) parseDefinitions() error {
	// sort the typeNames so that parsing definitions is deterministic
	typeNames := make([]string, 0, len(parser.registerTypes))
	for refTypeName := range parser.registerTypes {
		typeNames = append(typeNames, refTypeName)
	}
	sort.Strings(typeNames)

	for _, refTypeName := range typeNames {
		typeSpec := parser.registerTypes[refTypeName]
		ss := strings.Split(refTypeName, ".")
		pkgName := ss[0]
		parser.structStack = nil
		if err := parser.ParseDefinition(pkgName, typeSpec.Name.Name, typeSpec); err != nil {
			return err
		}
	}
	return nil
}

// ParseDefinition parses given type spec that corresponds to the type under
// given name and package, and populates swagger schema definitions registry
// with a schema for the given type
func (parser *Parser) ParseDefinition(pkgName, typeName string, typeSpec *ast.TypeSpec) error {
	refTypeName := fullTypeName(pkgName, typeName)

	if typeSpec == nil {
		Println("Skipping '" + refTypeName + "', pkg '" + pkgName + "' not found, try add flag --parseDependency or --parseVendor.")
		return nil
	}

	if _, isParsed := parser.swagger.Definitions[refTypeName]; isParsed {
		Println("Skipping '" + refTypeName + "', already parsed.")
		return nil
	}

	if parser.isInStructStack(refTypeName) {
		Println("Skipping '" + refTypeName + "', recursion detected.")
		return nil
	}
	parser.structStack = append(parser.structStack, refTypeName)

	Println("Generating " + refTypeName)

	schema, err := parser.parseTypeExpr(pkgName, typeName, typeSpec.Type)
	if err != nil {
		return err
	}
	parser.swagger.Definitions[refTypeName] = *schema
	return nil
}

func (parser *Parser) collectRequiredFields(pkgName string, properties map[string]spec.Schema, extraRequired []string) (requiredFields []string) {
	// created sorted list of properties keys so when we iterate over them it's deterministic
	ks := make([]string, 0, len(properties))
	for k := range properties {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	requiredFields = make([]string, 0)

	// iterate over keys list instead of map to avoid the random shuffle of the order that go does for maps
	for _, k := range ks {
		prop := properties[k]

		// todo find the pkgName of the property type
		tname := prop.SchemaProps.Type[0]
		if _, ok := parser.TypeDefinitions[pkgName][tname]; ok {
			tspec := parser.TypeDefinitions[pkgName][tname]
			parser.ParseDefinition(pkgName, tname, tspec)
		}
		requiredFields = append(requiredFields, prop.SchemaProps.Required...)
		properties[k] = prop
	}

	if extraRequired != nil {
		requiredFields = append(requiredFields, extraRequired...)
	}

	sort.Strings(requiredFields)

	return
}

func fullTypeName(pkgName, typeName string) string {
	if pkgName != "" {
		return pkgName + "." + typeName
	}
	return typeName
}

// parseTypeExpr parses given type expression that corresponds to the type under
// given name and package, and returns swagger schema for it.
func (parser *Parser) parseTypeExpr(pkgName, typeName string, typeExpr ast.Expr) (*spec.Schema, error) {

	switch expr := typeExpr.(type) {
	// type Foo struct {...}
	case *ast.StructType:
		refTypeName := fullTypeName(pkgName, typeName)
		if schema, isParsed := parser.swagger.Definitions[refTypeName]; isParsed {
			return &schema, nil
		}

		return parser.parseStruct(pkgName, expr.Fields)

	// type Foo Baz
	case *ast.Ident:
		if IsGolangPrimitiveType(expr.Name) {
			return &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{TransToValidSchemeType(expr.Name)},
				},
			}, nil
		}
		refTypeName := fullTypeName(pkgName, expr.Name)
		if _, isParsed := parser.swagger.Definitions[refTypeName]; !isParsed {
			if typedef, ok := parser.TypeDefinitions[pkgName][expr.Name]; ok {
				parser.ParseDefinition(pkgName, expr.Name, typedef)
			}
		}
		return &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Ref: spec.Ref{
					Ref: jsonreference.MustCreateRef("#/definitions/" + refTypeName),
				},
			},
		}, nil

	// type Foo *Baz
	case *ast.StarExpr:
		return parser.parseTypeExpr(pkgName, typeName, expr.X)

	// type Foo []Baz
	case *ast.ArrayType:
		itemSchema, err := parser.parseTypeExpr(pkgName, "", expr.Elt)
		if err != nil {
			return &spec.Schema{}, err
		}
		return &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"array"},
				Items: &spec.SchemaOrArray{
					Schema: itemSchema,
				},
			},
		}, nil

	// type Foo pkg.Bar
	case *ast.SelectorExpr:
		if xIdent, ok := expr.X.(*ast.Ident); ok {
			return parser.parseTypeExpr(xIdent.Name, expr.Sel.Name, expr.Sel)
		}

	// type Foo map[string]Bar
	case *ast.MapType:
		var valueSchema spec.SchemaOrBool
		if _, ok := expr.Value.(*ast.InterfaceType); ok {
			valueSchema.Allows = true
		} else {
			schema, err := parser.parseTypeExpr(pkgName, "", expr.Value)
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
	// ...
	default:
		Printf("Type definition of type '%T' is not supported yet. Using 'object' instead.\n", typeExpr)
	}

	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
		},
	}, nil
}

func (parser *Parser) parseStruct(pkgName string, fields *ast.FieldList) (*spec.Schema, error) {

	extraRequired := make([]string, 0)
	properties := make(map[string]spec.Schema)
	for _, field := range fields.List {
		fieldProps, requiredFromAnon, err := parser.parseStructField(pkgName, field)
		if err != nil {
			return &spec.Schema{}, err
		}
		extraRequired = append(extraRequired, requiredFromAnon...)
		for k, v := range fieldProps {
			properties[k] = v
		}
	}

	// collect requireds from our properties and anonymous fields
	required := parser.collectRequiredFields(pkgName, properties, extraRequired)

	// unset required from properties because we've collected them
	for k, prop := range properties {
		prop.SchemaProps.Required = make([]string, 0)
		properties[k] = prop
	}

	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       []string{"object"},
			Properties: properties,
			Required:   required,
		}}, nil
}

type structField struct {
	name         string
	desc         string
	schemaType   string
	arrayType    string
	formatType   string
	isRequired   bool
	readOnly     bool
	crossPkg     string
	exampleValue interface{}
	maximum      *float64
	minimum      *float64
	maxLength    *int64
	minLength    *int64
	enums        []interface{}
	defaultValue interface{}
	extensions   map[string]interface{}
}

func (sf *structField) toStandardSchema() *spec.Schema {
	required := make([]string, 0)
	if sf.isRequired {
		required = append(required, sf.name)
	}
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:        []string{sf.schemaType},
			Description: sf.desc,
			Format:      sf.formatType,
			Required:    required,
			Maximum:     sf.maximum,
			Minimum:     sf.minimum,
			MaxLength:   sf.maxLength,
			MinLength:   sf.minLength,
			Enum:        sf.enums,
			Default:     sf.defaultValue,
		},
		SwaggerSchemaProps: spec.SwaggerSchemaProps{
			Example:  sf.exampleValue,
			ReadOnly: sf.readOnly,
		},
		VendorExtensible: spec.VendorExtensible{
			Extensions: sf.extensions,
		},
	}
}

func (parser *Parser) parseStructField(pkgName string, field *ast.Field) (map[string]spec.Schema, []string, error) {
	properties := map[string]spec.Schema{}

	if field.Names == nil {
		fullTypeName, err := getFieldType(field.Type)
		if err != nil {
			return properties, []string{}, nil
		}

		typeName := fullTypeName

		if splits := strings.Split(fullTypeName, "."); len(splits) > 1 {
			pkgName = splits[0]
			typeName = splits[1]
		}

		typeSpec := parser.TypeDefinitions[pkgName][typeName]
		if typeSpec != nil {
			schema, err := parser.parseTypeExpr(pkgName, typeName, typeSpec.Type)
			if err != nil {
				return properties, []string{}, err
			}
			schemaType := "unknown"
			if len(schema.SchemaProps.Type) > 0 {
				schemaType = schema.SchemaProps.Type[0]
			}

			switch schemaType {
			case "object":
				for k, v := range schema.SchemaProps.Properties {
					properties[k] = v
				}
			case "array":
				properties[typeName] = *schema
			default:
				Printf("Can't extract properties from a schema of type '%s'", schemaType)
			}
			return properties, schema.SchemaProps.Required, nil
		}

		return properties, nil, nil
	}

	structField, err := parser.parseField(field)
	if err != nil {
		return properties, nil, err
	}
	if structField.name == "" {
		return properties, nil, nil
	}

	// TODO: find package of schemaType and/or arrayType
	if structField.crossPkg != "" {
		pkgName = structField.crossPkg
	}

	fillObject := func(src, dest interface{}) error {
		bin, err := json.Marshal(src)
		if err != nil {
			return err
		}
		return json.Unmarshal(bin, dest)
	}

	//for spec.Schema have implemented json.Marshaler, here in another way to convert
	fillSchema := func(src, dest *spec.Schema) error {
		err = fillObject(&src.SchemaProps, &dest.SchemaProps)
		if err != nil {
			return err
		}
		err = fillObject(&src.SwaggerSchemaProps, &dest.SwaggerSchemaProps)
		if err != nil {
			return err
		}
		return fillObject(&src.VendorExtensible, &dest.VendorExtensible)
	}

	if _, ok := parser.TypeDefinitions[pkgName][structField.schemaType]; ok { // user type field
		// write definition if not yet present
		parser.ParseDefinition(pkgName, structField.schemaType,
			parser.TypeDefinitions[pkgName][structField.schemaType])
		required := make([]string, 0)
		if structField.isRequired {
			required = append(required, structField.name)
		}
		properties[structField.name] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:        []string{"object"}, // to avoid swagger validation error
				Description: structField.desc,
				Required:    required,
				Ref: spec.Ref{
					Ref: jsonreference.MustCreateRef("#/definitions/" + pkgName + "." + structField.schemaType),
				},
			},
			SwaggerSchemaProps: spec.SwaggerSchemaProps{
				ReadOnly: structField.readOnly,
			},
		}
	} else if structField.schemaType == "array" { // array field type
		// if defined -- ref it
		if _, ok := parser.TypeDefinitions[pkgName][structField.arrayType]; ok { // user type in array
			parser.ParseDefinition(pkgName, structField.arrayType,
				parser.TypeDefinitions[pkgName][structField.arrayType])
			required := make([]string, 0)
			if structField.isRequired {
				required = append(required, structField.name)
			}
			properties[structField.name] = spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        []string{structField.schemaType},
					Description: structField.desc,
					Required:    required,
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Ref: spec.Ref{
									Ref: jsonreference.MustCreateRef("#/definitions/" + pkgName + "." + structField.arrayType),
								},
							},
						},
					},
				},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: structField.readOnly,
				},
			}
		} else if structField.arrayType == "object" {
			// Anonymous struct
			if astTypeArray, ok := field.Type.(*ast.ArrayType); ok { // if array
				props := make(map[string]spec.Schema)
				if expr, ok := astTypeArray.Elt.(*ast.StructType); ok {
					for _, field := range expr.Fields.List {
						var fieldProps map[string]spec.Schema
						fieldProps, _, err = parser.parseStructField(pkgName, field)
						if err != nil {
							return properties, nil, err
						}
						for k, v := range fieldProps {
							props[k] = v
						}
					}
					properties[structField.name] = spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type:        []string{structField.schemaType},
							Description: structField.desc,
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:       []string{"object"},
										Properties: props,
									},
								},
							},
						},
						SwaggerSchemaProps: spec.SwaggerSchemaProps{
							ReadOnly: structField.readOnly,
						},
					}
				} else {
					schema, _ := parser.parseTypeExpr(pkgName, "", astTypeArray.Elt)
					properties[structField.name] = spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type:        []string{structField.schemaType},
							Description: structField.desc,
							Items: &spec.SchemaOrArray{
								Schema: schema,
							},
						},
						SwaggerSchemaProps: spec.SwaggerSchemaProps{
							ReadOnly: structField.readOnly,
						},
					}
				}
			}
		} else if structField.arrayType == "array" {
			if astTypeArray, ok := field.Type.(*ast.ArrayType); ok {
				schema, _ := parser.parseTypeExpr(pkgName, "", astTypeArray.Elt)
				properties[structField.name] = spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type:        []string{structField.schemaType},
						Description: structField.desc,
						Items: &spec.SchemaOrArray{
							Schema: schema,
						},
					},
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: structField.readOnly,
					},
				}
			}
		} else {
			// standard type in array
			required := make([]string, 0)
			if structField.isRequired {
				required = append(required, structField.name)
			}

			properties[structField.name] = spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        []string{structField.schemaType},
					Description: structField.desc,
					Format:      structField.formatType,
					Required:    required,
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type:      []string{structField.arrayType},
								Maximum:   structField.maximum,
								Minimum:   structField.minimum,
								MaxLength: structField.maxLength,
								MinLength: structField.minLength,
								Enum:      structField.enums,
								Default:   structField.defaultValue,
							},
						},
					},
				},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					Example:  structField.exampleValue,
					ReadOnly: structField.readOnly,
				},
			}
		}
	} else if astTypeMap, ok := field.Type.(*ast.MapType); ok { // if map
		stdSchema := structField.toStandardSchema()
		mapValueSchema, err := parser.parseTypeExpr(pkgName, "", astTypeMap)
		if err != nil {
			return properties, nil, err
		}
		stdSchema.Type = mapValueSchema.Type
		stdSchema.AdditionalProperties = mapValueSchema.AdditionalProperties
		properties[structField.name] = *stdSchema
	} else {
		stdSchema := structField.toStandardSchema()
		properties[structField.name] = *stdSchema

		if nestStar, ok := field.Type.(*ast.StarExpr); ok {
			if !IsGolangPrimitiveType(structField.schemaType) {
				schema, err := parser.parseTypeExpr(pkgName, structField.schemaType, nestStar.X)
				if err != nil {
					return properties, nil, err
				}

				if len(schema.SchemaProps.Type) > 0 {
					err = fillSchema(schema, stdSchema)
					if err != nil {
						return properties, nil, err
					}
					properties[structField.name] = *stdSchema
					return properties, nil, nil
				}
			}
		} else if nestStruct, ok := field.Type.(*ast.StructType); ok {
			props := map[string]spec.Schema{}
			nestRequired := make([]string, 0)
			for _, v := range nestStruct.Fields.List {
				p, _, err := parser.parseStructField(pkgName, v)
				if err != nil {
					return properties, nil, err
				}
				for k, v := range p {
					if v.SchemaProps.Type[0] != "object" {
						nestRequired = append(nestRequired, v.SchemaProps.Required...)
						v.SchemaProps.Required = make([]string, 0)
					}
					props[k] = v
				}
			}
			stdSchema.Properties = props
			stdSchema.Required = nestRequired
			properties[structField.name] = *stdSchema
		}
	}
	return properties, nil, nil
}

func getFieldType(field interface{}) (string, error) {

	switch ftype := field.(type) {
	case *ast.Ident:
		return ftype.Name, nil

	case *ast.SelectorExpr:
		packageName, err := getFieldType(ftype.X)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.%s", packageName, ftype.Sel.Name), nil

	case *ast.StarExpr:
		fullName, err := getFieldType(ftype.X)
		if err != nil {
			return "", err
		}
		return fullName, nil

	}
	return "", fmt.Errorf("unknown field type %#v", field)
}

func (parser *Parser) parseField(field *ast.Field) (*structField, error) {
	prop, err := getPropertyName(field.Type, parser)
	if err != nil {
		return nil, err
	}

	if len(prop.ArrayType) == 0 {
		if err := CheckSchemaType(prop.SchemaType); err != nil {
			return nil, err
		}
	} else {
		if err := CheckSchemaType("array"); err != nil {
			return nil, err
		}
	}

	structField := &structField{
		name:       field.Names[0].Name,
		schemaType: prop.SchemaType,
		arrayType:  prop.ArrayType,
		crossPkg:   prop.CrossPkg,
	}

	switch parser.PropNamingStrategy {
	case SnakeCase:
		structField.name = toSnakeCase(structField.name)
	case PascalCase:
		//use struct field name
	case CamelCase:
		structField.name = toLowerCamelCase(structField.name)
	default:
		structField.name = toLowerCamelCase(structField.name)
	}

	if field.Doc != nil {
		structField.desc = strings.TrimSpace(field.Doc.Text())
	}
	if structField.desc == "" && field.Comment != nil {
		structField.desc = strings.TrimSpace(field.Comment.Text())
	}

	if field.Tag == nil {
		return structField, nil
	}
	// `json:"tag"` -> json:"tag"
	structTag := reflect.StructTag(strings.Replace(field.Tag.Value, "`", "", -1))
	jsonTag := structTag.Get("json")
	// json:"tag,hoge"
	if strings.Contains(jsonTag, ",") {
		// json:",hoge"
		if strings.HasPrefix(jsonTag, ",") {
			jsonTag = ""
		} else {
			jsonTag = strings.SplitN(jsonTag, ",", 2)[0]
		}
	}
	if jsonTag == "-" {
		structField.name = ""
	} else if jsonTag != "" {
		structField.name = jsonTag
	}

	if typeTag := structTag.Get("swaggertype"); typeTag != "" {
		parts := strings.Split(typeTag, ",")
		if 0 < len(parts) && len(parts) <= 2 {
			newSchemaType := parts[0]
			newArrayType := structField.arrayType
			if len(parts) >= 2 {
				if newSchemaType == "array" {
					newArrayType = parts[1]
					if err := CheckSchemaType(newArrayType); err != nil {
						return nil, err
					}
				} else if newSchemaType == "primitive" {
					newSchemaType = parts[1]
					newArrayType = parts[1]
				}
			}

			if err := CheckSchemaType(newSchemaType); err != nil {
				return nil, err
			}

			structField.schemaType = newSchemaType
			structField.arrayType = newArrayType
		}
	}
	if exampleTag := structTag.Get("example"); exampleTag != "" {
		example, err := defineTypeOfExample(structField.schemaType, structField.arrayType, exampleTag)
		if err != nil {
			return nil, err
		}
		structField.exampleValue = example
	}
	if formatTag := structTag.Get("format"); formatTag != "" {
		structField.formatType = formatTag
	}
	if bindingTag := structTag.Get("binding"); bindingTag != "" {
		for _, val := range strings.Split(bindingTag, ",") {
			if val == "required" {
				structField.isRequired = true
				break
			}
		}
	}
	if validateTag := structTag.Get("validate"); validateTag != "" {
		for _, val := range strings.Split(validateTag, ",") {
			if val == "required" {
				structField.isRequired = true
				break
			}
		}
	}
	if extensionsTag := structTag.Get("extensions"); extensionsTag != "" {
		structField.extensions = map[string]interface{}{}
		for _, val := range strings.Split(extensionsTag, ",") {
			parts := strings.SplitN(val, "=", 2)
			if len(parts) == 2 {
				structField.extensions[parts[0]] = parts[1]
			} else {
				structField.extensions[parts[0]] = true
			}
		}
	}
	if enumsTag := structTag.Get("enums"); enumsTag != "" {
		enumType := structField.schemaType
		if structField.schemaType == "array" {
			enumType = structField.arrayType
		}

		for _, e := range strings.Split(enumsTag, ",") {
			value, err := defineType(enumType, e)
			if err != nil {
				return nil, err
			}
			structField.enums = append(structField.enums, value)
		}
	}
	if defaultTag := structTag.Get("default"); defaultTag != "" {
		value, err := defineType(structField.schemaType, defaultTag)
		if err != nil {
			return nil, err
		}
		structField.defaultValue = value
	}

	if IsNumericType(structField.schemaType) || IsNumericType(structField.arrayType) {
		maximum, err := getFloatTag(structTag, "maximum")
		if err != nil {
			return nil, err
		}
		structField.maximum = maximum

		minimum, err := getFloatTag(structTag, "minimum")
		if err != nil {
			return nil, err
		}
		structField.minimum = minimum
	}
	if structField.schemaType == "string" || structField.arrayType == "string" {
		maxLength, err := getIntTag(structTag, "maxLength")
		if err != nil {
			return nil, err
		}
		structField.maxLength = maxLength

		minLength, err := getIntTag(structTag, "minLength")
		if err != nil {
			return nil, err
		}
		structField.minLength = minLength
	}
	if readOnly := structTag.Get("readonly"); readOnly != "" {
		structField.readOnly = readOnly == "true"
	}

	return structField, nil
}

func replaceLastTag(slice []spec.Tag, element spec.Tag) {
	slice = slice[:len(slice)-1]
	slice = append(slice, element)
}

func getFloatTag(structTag reflect.StructTag, tagName string) (*float64, error) {
	strValue := structTag.Get(tagName)
	if strValue == "" {
		return nil, nil
	}

	value, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		return nil, fmt.Errorf("can't parse numeric value of %q tag: %v", tagName, err)
	}

	return &value, nil
}

func getIntTag(structTag reflect.StructTag, tagName string) (*int64, error) {
	strValue := structTag.Get(tagName)
	if strValue == "" {
		return nil, nil
	}

	value, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("can't parse numeric value of %q tag: %v", tagName, err)
	}

	return &value, nil
}

func toSnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}
	return string(out)
}

func toLowerCamelCase(in string) string {
	runes := []rune(in)

	var out []rune
	flag := false
	for i, curr := range runes {
		if (i == 0 && unicode.IsUpper(curr)) || (flag && unicode.IsUpper(curr)) {
			out = append(out, unicode.ToLower(curr))
			flag = true
		} else {
			out = append(out, curr)
			flag = false
		}
	}

	return string(out)
}

// defineTypeOfExample example value define the type (object and array unsupported)
func defineTypeOfExample(schemaType, arrayType, exampleValue string) (interface{}, error) {
	switch schemaType {
	case "string":
		return exampleValue, nil
	case "number":
		v, err := strconv.ParseFloat(exampleValue, 64)
		if err != nil {
			return nil, fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err)
		}
		return v, nil
	case "integer":
		v, err := strconv.Atoi(exampleValue)
		if err != nil {
			return nil, fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err)
		}
		return v, nil
	case "boolean":
		v, err := strconv.ParseBool(exampleValue)
		if err != nil {
			return nil, fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err)
		}
		return v, nil
	case "array":
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
	default:
		return nil, fmt.Errorf("%s is unsupported type in example value", schemaType)
	}
}

// GetAllGoFileInfo gets all Go source files information for given searchDir.
func (parser *Parser) getAllGoFileInfo(searchDir string) error {
	return filepath.Walk(searchDir, parser.visit)
}

func (parser *Parser) getAllGoFileInfoFromDeps(pkg *depth.Pkg) error {
	if pkg.Internal || !pkg.Resolved { // ignored internal and not resolved dependencies
		return nil
	}

	srcDir := pkg.Raw.Dir
	files, err := ioutil.ReadDir(srcDir) // only parsing files in the dir(don't contains sub dir files)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(srcDir, f.Name())
		if err := parser.parseFile(path); err != nil {
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

func (parser *Parser) visit(path string, f os.FileInfo, err error) error {
	if err := parser.Skip(path, f); err != nil {
		return err
	}
	return parser.parseFile(path)
}

func (parser *Parser) parseFile(path string) error {
	if ext := filepath.Ext(path); ext == ".go" {
		fset := token.NewFileSet() // positions are relative to fset
		astFile, err := goparser.ParseFile(fset, path, nil, goparser.ParseComments)
		if err != nil {
			return fmt.Errorf("ParseFile error:%+v", err)
		}

		parser.files[path] = astFile
	}
	return nil
}

// Skip returns filepath.SkipDir error if match vendor and hidden folder
func (parser *Parser) Skip(path string, f os.FileInfo) error {

	if !parser.ParseVendor { // ignore vendor
		if f.IsDir() && f.Name() == "vendor" {
			return filepath.SkipDir
		}
	}

	// issue
	if f.IsDir() && f.Name() == "docs" {
		return filepath.SkipDir
	}

	// exclude all hidden folder
	if f.IsDir() && len(f.Name()) > 1 && f.Name()[0] == '.' {
		return filepath.SkipDir
	}
	return nil
}

// GetSwagger returns *spec.Swagger which is the root document object for the API specification.
func (parser *Parser) GetSwagger() *spec.Swagger {
	return parser.swagger
}
