package swag

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
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

	//files is a map that stores map[real_go_file_path][astFile]
	files map[string]*ast.File

	// TypeDefinitions is a map that stores [package name][type name][*ast.TypeSpec]
	TypeDefinitions map[string]map[string]*ast.TypeSpec

	// CustomPrimitiveTypes is a map that stores custom primitive types to actual golang types [type name][string]
	CustomPrimitiveTypes map[string]string

	//registerTypes is a map that stores [refTypeName][*ast.TypeSpec]
	registerTypes map[string]*ast.TypeSpec

	PropNamingStrategy string

	ParseVendor bool

	// structStack stores full names of the structures that were already parsed or are being parsed now
	structStack []string
}

// New creates a new Parser with default properties.
func New() *Parser {
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
		CustomPrimitiveTypes: make(map[string]string),
		registerTypes:        make(map[string]*ast.TypeSpec),
	}
	return parser
}

// ParseAPI parses general api info for gived searchDir and mainAPIFile
func (parser *Parser) ParseAPI(searchDir string, mainAPIFile string) error {
	Println("Generate general API Info")
	if err := parser.getAllGoFileInfo(searchDir); err != nil {
		return err
	}
	parser.ParseGeneralAPIInfo(path.Join(searchDir, mainAPIFile))

	for _, astFile := range parser.files {
		parser.ParseType(astFile)
	}

	for fileName, astFile := range parser.files {
		if err := parser.ParseRouterAPIInfo(fileName, astFile); err != nil {
			return err
		}
	}

	parser.ParseDefinitions()

	return nil
}

// ParseGeneralAPIInfo parses general api info for gived mainAPIFile path
func (parser *Parser) ParseGeneralAPIInfo(mainAPIFile string) error {
	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, mainAPIFile, nil, goparser.ParseComments)
	if err != nil {
		return errors.Wrap(err, "cannot parse soure files")
	}

	parser.swagger.Swagger = "2.0"
	securityMap := map[string]*spec.SecurityScheme{}

	// templated defaults
	parser.swagger.Info.Version = "{{.Version}}"
	parser.swagger.Info.Title = "{{.Title}}"
	parser.swagger.Info.Description = "{{.Description}}"
	parser.swagger.Host = "{{.Host}}"
	parser.swagger.BasePath = "{{.BasePath}}"

	if fileTree.Comments != nil {
		for _, comment := range fileTree.Comments {
			comments := strings.Split(comment.Text(), "\n")
			previousAttribute := ""
			for _, commentLine := range comments {
				attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
				multilineBlock := false
				if previousAttribute == attribute {
					multilineBlock = true
				}
				switch attribute {
				case "@version":
					parser.swagger.Info.Version = strings.TrimSpace(commentLine[len(attribute):])
				case "@title":
					parser.swagger.Info.Title = strings.TrimSpace(commentLine[len(attribute):])
				case "@description":
					if parser.swagger.Info.Description == "{{.Description}}" {
						parser.swagger.Info.Description = strings.TrimSpace(commentLine[len(attribute):])
					} else if multilineBlock {
						parser.swagger.Info.Description += "\n" + strings.TrimSpace(commentLine[len(attribute):])
					}
				case "@termsofservice":
					parser.swagger.Info.TermsOfService = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact.name":
					parser.swagger.Info.Contact.Name = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact.email":
					parser.swagger.Info.Contact.Email = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact.url":
					parser.swagger.Info.Contact.URL = strings.TrimSpace(commentLine[len(attribute):])
				case "@license.name":
					parser.swagger.Info.License.Name = strings.TrimSpace(commentLine[len(attribute):])
				case "@license.url":
					parser.swagger.Info.License.URL = strings.TrimSpace(commentLine[len(attribute):])
				case "@host":
					parser.swagger.Host = strings.TrimSpace(commentLine[len(attribute):])
				case "@basepath":
					parser.swagger.BasePath = strings.TrimSpace(commentLine[len(attribute):])
				case "@schemes":
					parser.swagger.Schemes = getSchemes(commentLine)
				case "@tag.name":
					commentInfo := strings.TrimSpace(commentLine[len(attribute):])
					parser.swagger.Tags = append(parser.swagger.Tags, spec.Tag{
						TagProps: spec.TagProps{
							Name: strings.TrimSpace(commentInfo),
						},
					})
				case "@tag.description":
					commentInfo := strings.TrimSpace(commentLine[len(attribute):])
					tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
					tag.TagProps.Description = commentInfo
					replaceLastTag(parser.swagger.Tags, tag)
				case "@tag.docs.url":
					commentInfo := strings.TrimSpace(commentLine[len(attribute):])
					tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
					tag.TagProps.ExternalDocs = &spec.ExternalDocumentation{
						URL: commentInfo,
					}
					replaceLastTag(parser.swagger.Tags, tag)

				case "@tag.docs.description":
					commentInfo := strings.TrimSpace(commentLine[len(attribute):])
					tag := parser.swagger.Tags[len(parser.swagger.Tags)-1]
					if tag.TagProps.ExternalDocs == nil {
						return errors.New("@tag.docs.description needs to come after a @tags.docs.url")
					}
					tag.TagProps.ExternalDocs.Description = commentInfo
					replaceLastTag(parser.swagger.Tags, tag)
				}
				previousAttribute = attribute
			}

			for i := 0; i < len(comments); i++ {
				attribute := strings.ToLower(strings.Split(comments[i], " ")[0])
				switch attribute {
				case "@securitydefinitions.basic":
					securityMap[strings.TrimSpace(comments[i][len(attribute):])] = spec.BasicAuth()
				case "@securitydefinitions.apikey":
					attrMap := map[string]string{}
					for _, v := range comments[i+1:] {
						securityAttr := strings.ToLower(strings.Split(v, " ")[0])
						if securityAttr == "@in" || securityAttr == "@name" {
							attrMap[securityAttr] = strings.TrimSpace(v[len(securityAttr):])
						}
						// next securityDefinitions
						if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
							break
						}
					}
					if len(attrMap) != 2 {
						return errors.New("@securitydefinitions.apikey is @name and @in required")
					}
					securityMap[strings.TrimSpace(comments[i][len(attribute):])] = spec.APIKeyAuth(attrMap["@name"], attrMap["@in"])
				case "@securitydefinitions.oauth2.application":
					attrMap := map[string]string{}
					scopes := map[string]string{}
					for _, v := range comments[i+1:] {
						securityAttr := strings.ToLower(strings.Split(v, " ")[0])
						if securityAttr == "@tokenurl" {
							attrMap[securityAttr] = strings.TrimSpace(v[len(securityAttr):])
						} else {
							isExists, err := isExistsScope(securityAttr)
							if err != nil {
								return err
							}
							if isExists {
								scopScheme, err := getScopeScheme(securityAttr)
								if err != nil {
									return err
								}
								scopes[scopScheme] = v[len(securityAttr):]
							}
						}
						// next securityDefinitions
						if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
							break
						}
					}
					if len(attrMap) != 1 {
						return errors.New("@securitydefinitions.oauth2.application is @tokenUrl required")
					}
					securityScheme := spec.OAuth2Application(attrMap["@tokenurl"])
					for scope, description := range scopes {
						securityScheme.AddScope(scope, description)
					}
					securityMap[strings.TrimSpace(comments[i][len(attribute):])] = securityScheme
				case "@securitydefinitions.oauth2.implicit":
					attrMap := map[string]string{}
					scopes := map[string]string{}
					for _, v := range comments[i+1:] {
						securityAttr := strings.ToLower(strings.Split(v, " ")[0])
						if securityAttr == "@authorizationurl" {
							attrMap[securityAttr] = strings.TrimSpace(v[len(securityAttr):])
						} else {
							isExists, err := isExistsScope(securityAttr)
							if err != nil {
								return err
							}
							if isExists {
								scopScheme, err := getScopeScheme(securityAttr)
								if err != nil {
									return err
								}
								scopes[scopScheme] = v[len(securityAttr):]
							}
						}
						// next securityDefinitions
						if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
							break
						}
					}
					if len(attrMap) != 1 {
						return errors.New("@securitydefinitions.oauth2.implicit is @authorizationUrl required")
					}
					securityScheme := spec.OAuth2Implicit(attrMap["@authorizationurl"])
					for scope, description := range scopes {
						securityScheme.AddScope(scope, description)
					}
					securityMap[strings.TrimSpace(comments[i][len(attribute):])] = securityScheme
				case "@securitydefinitions.oauth2.password":
					attrMap := map[string]string{}
					scopes := map[string]string{}
					for _, v := range comments[i+1:] {
						securityAttr := strings.ToLower(strings.Split(v, " ")[0])
						if securityAttr == "@tokenurl" {
							attrMap[securityAttr] = strings.TrimSpace(v[len(securityAttr):])
						} else {
							isExists, err := isExistsScope(securityAttr)
							if err != nil {
								return err
							}
							if isExists {
								scopScheme, err := getScopeScheme(securityAttr)
								if err != nil {
									return err
								}
								scopes[scopScheme] = v[len(securityAttr):]
							}
						}
						// next securityDefinitions
						if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
							break
						}
					}
					if len(attrMap) != 1 {
						return errors.New("@securitydefinitions.oauth2.password is @tokenUrl required")
					}
					securityScheme := spec.OAuth2Password(attrMap["@tokenurl"])
					for scope, description := range scopes {
						securityScheme.AddScope(scope, description)
					}
					securityMap[strings.TrimSpace(comments[i][len(attribute):])] = securityScheme
				case "@securitydefinitions.oauth2.accesscode":
					attrMap := map[string]string{}
					scopes := map[string]string{}
					for _, v := range comments[i+1:] {
						securityAttr := strings.ToLower(strings.Split(v, " ")[0])
						if securityAttr == "@tokenurl" || securityAttr == "@authorizationurl" {
							attrMap[securityAttr] = strings.TrimSpace(v[len(securityAttr):])
						} else {
							isExists, err := isExistsScope(securityAttr)
							if err != nil {
								return err
							}
							if isExists {
								scopScheme, err := getScopeScheme(securityAttr)
								if err != nil {
									return err
								}
								scopes[scopScheme] = v[len(securityAttr):]
							}
						}
						// next securityDefinitions
						if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
							break
						}
					}
					if len(attrMap) != 2 {
						return errors.New("@securitydefinitions.oauth2.accessCode is @tokenUrl and @authorizationUrl required")
					}
					securityScheme := spec.OAuth2AccessToken(attrMap["@authorizationurl"], attrMap["@tokenurl"])
					for scope, description := range scopes {
						securityScheme.AddScope(scope, description)
					}
					securityMap[strings.TrimSpace(comments[i][len(attribute):])] = securityScheme
				}
			}
		}
	}
	if len(securityMap) > 0 {
		parser.swagger.SecurityDefinitions = securityMap
	}

	return nil
}

func getScopeScheme(scope string) (string, error) {
	scopeValue := scope[strings.Index(scope, "@scope."):]
	if scopeValue == "" {
		return "", errors.New("@scope is empty")
	}
	return scope[len("@scope."):], nil
}

func isExistsScope(scope string) (bool, error) {
	s := strings.Fields(scope)
	for _, v := range s {
		if strings.Index(v, "@scope.") != -1 {
			if strings.Index(v, ",") != -1 {
				return false, fmt.Errorf("@scope can't use comma(,) get=" + v)
			}
		}
	}
	return strings.Index(scope, "@scope.") != -1, nil
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
}

func (parser *Parser) isInStructStack(refTypeName string) bool {
	for _, structName := range parser.structStack {
		if refTypeName == structName {
			return true
		}
	}
	return false
}

// ParseDefinitions parses Swagger Api definitions.
func (parser *Parser) ParseDefinitions() {
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
		parser.ParseDefinition(pkgName, typeSpec.Name.Name, typeSpec)
	}
}

// ParseDefinition parses given type spec that corresponds to the type under
// given name and package, and populates swagger schema definitions registry
// with a schema for the given type
func (parser *Parser) ParseDefinition(pkgName, typeName string, typeSpec *ast.TypeSpec) error {
	refTypeName := fullTypeName(pkgName, typeName)
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
	parser.swagger.Definitions[refTypeName] = schema
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
		if tname != "object" {
			requiredFields = append(requiredFields, prop.SchemaProps.Required...)
		}
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
func (parser *Parser) parseTypeExpr(pkgName, typeName string, typeExpr ast.Expr) (spec.Schema, error) {
	//TODO: return pointer to spec.Schema

	switch expr := typeExpr.(type) {
	// type Foo struct {...}
	case *ast.StructType:
		refTypeName := fullTypeName(pkgName, typeName)
		if schema, isParsed := parser.swagger.Definitions[refTypeName]; isParsed {
			return schema, nil
		}

		extraRequired := make([]string, 0)
		properties := make(map[string]spec.Schema)
		for _, field := range expr.Fields.List {
			var fieldProps map[string]spec.Schema
			var requiredFromAnon []string
			if field.Names == nil {
				var err error
				fieldProps, requiredFromAnon, err = parser.parseAnonymousField(pkgName, field)
				if err != nil {
					return spec.Schema{}, err
				}
				extraRequired = append(extraRequired, requiredFromAnon...)
			} else {
				var err error
				fieldProps, err = parser.parseStruct(pkgName, field)
				if err != nil {
					return spec.Schema{}, err
				}
			}

			for k, v := range fieldProps {
				properties[k] = v
			}
		}

		// collect requireds from our properties and anonymous fields
		required := parser.collectRequiredFields(pkgName, properties, extraRequired)

		// unset required from properties because we've collected them
		for k, prop := range properties {
			tname := prop.SchemaProps.Type[0]
			if tname != "object" {
				prop.SchemaProps.Required = make([]string, 0)
			}
			properties[k] = prop
		}

		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:       []string{"object"},
				Properties: properties,
				Required:   required,
			}}, nil

	// type Foo Baz
	case *ast.Ident:
		refTypeName := fullTypeName(pkgName, expr.Name)
		if _, isParsed := parser.swagger.Definitions[refTypeName]; !isParsed {
			if typedef, ok := parser.TypeDefinitions[pkgName][expr.Name]; ok {
				parser.ParseDefinition(pkgName, expr.Name, typedef)
			}
		}
		return parser.swagger.Definitions[refTypeName], nil

	// type Foo *Baz
	case *ast.StarExpr:
		return parser.parseTypeExpr(pkgName, typeName, expr.X)

	// type Foo []Baz
	case *ast.ArrayType:
		itemSchema, err := parser.parseTypeExpr(pkgName, "", expr.Elt)
		if err != nil {
			return spec.Schema{}, err
		}
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"array"},
				Items: &spec.SchemaOrArray{
					Schema: &itemSchema,
				},
			},
		}, nil

	// type Foo pkg.Bar
	case *ast.SelectorExpr:
		if xIdent, ok := expr.X.(*ast.Ident); ok {
			pkgName = xIdent.Name
			typeName = expr.Sel.Name
			refTypeName := fullTypeName(pkgName, typeName)
			if _, isParsed := parser.swagger.Definitions[refTypeName]; !isParsed {
				typedef := parser.TypeDefinitions[pkgName][typeName]
				parser.ParseDefinition(pkgName, typeName, typedef)
			}
			return parser.swagger.Definitions[refTypeName], nil
		}

	// type Foo map[string]Bar
	case *ast.MapType:
		itemSchema, err := parser.parseTypeExpr(pkgName, "", expr.Value)
		if err != nil {
			return spec.Schema{}, err
		}
		return spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
				AdditionalProperties: &spec.SchemaOrBool{
					Schema: &itemSchema,
				},
			},
		}, nil
	// ...
	default:
		Printf("Type definition of type '%T' is not supported yet. Using 'object' instead.\n", typeExpr)
	}

	return spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
		},
	}, nil
}

type structField struct {
	name         string
	schemaType   string
	arrayType    string
	formatType   string
	isRequired   bool
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

func (parser *Parser) parseStruct(pkgName string, field *ast.Field) (map[string]spec.Schema, error) {
	properties := map[string]spec.Schema{}
	structField, err := parser.parseField(field)
	if err != nil {
		return properties, nil
	}
	if structField.name == "" {
		return properties, nil
	}
	var desc string
	if field.Doc != nil {
		desc = strings.TrimSpace(field.Doc.Text())
	}
	if desc == "" && field.Comment != nil {
		desc = strings.TrimSpace(field.Comment.Text())
	}
	// TODO: find package of schemaType and/or arrayType

	if structField.crossPkg != "" {
		pkgName = structField.crossPkg
	}
	if _, ok := parser.TypeDefinitions[pkgName][structField.schemaType]; ok { // user type field
		// write definition if not yet present
		parser.ParseDefinition(pkgName, structField.schemaType,
			parser.TypeDefinitions[pkgName][structField.schemaType])
		properties[structField.name] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:        []string{"object"}, // to avoid swagger validation error
				Description: desc,
				Ref: spec.Ref{
					Ref: jsonreference.MustCreateRef("#/definitions/" + pkgName + "." + structField.schemaType),
				},
			},
		}
	} else if structField.schemaType == "array" { // array field type
		// if defined -- ref it
		if _, ok := parser.TypeDefinitions[pkgName][structField.arrayType]; ok { // user type in array
			parser.ParseDefinition(pkgName, structField.arrayType,
				parser.TypeDefinitions[pkgName][structField.arrayType])
			properties[structField.name] = spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        []string{structField.schemaType},
					Description: desc,
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
			}
		} else { // standard type in array
			required := make([]string, 0)
			if structField.isRequired {
				required = append(required, structField.name)
			}

			properties[structField.name] = spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        []string{structField.schemaType},
					Description: desc,
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
					Example: structField.exampleValue,
				},
			}
		}
	} else {
		required := make([]string, 0)
		if structField.isRequired {
			required = append(required, structField.name)
		}
		properties[structField.name] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:        []string{structField.schemaType},
				Description: desc,
				Format:      structField.formatType,
				Required:    required,
				Maximum:     structField.maximum,
				Minimum:     structField.minimum,
				MaxLength:   structField.maxLength,
				MinLength:   structField.minLength,
				Enum:        structField.enums,
				Default:     structField.defaultValue,
			},
			SwaggerSchemaProps: spec.SwaggerSchemaProps{
				Example: structField.exampleValue,
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: structField.extensions,
			},
		}

		nestStruct, ok := field.Type.(*ast.StructType)
		if ok {
			props := map[string]spec.Schema{}
			nestRequired := make([]string, 0)
			for _, v := range nestStruct.Fields.List {
				p, err := parser.parseStruct(pkgName, v)
				if err != nil {
					return properties, err
				}
				for k, v := range p {
					if v.SchemaProps.Type[0] != "object" {
						nestRequired = append(nestRequired, v.SchemaProps.Required...)
						v.SchemaProps.Required = make([]string, 0)
					}
					props[k] = v
				}
			}

			properties[structField.name] = spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:        []string{structField.schemaType},
					Description: desc,
					Format:      structField.formatType,
					Properties:  props,
					Required:    nestRequired,
					Maximum:     structField.maximum,
					Minimum:     structField.minimum,
					MaxLength:   structField.maxLength,
					MinLength:   structField.minLength,
					Enum:        structField.enums,
					Default:     structField.defaultValue,
				},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					Example: structField.exampleValue,
				},
			}
		}
	}
	return properties, nil
}

func (parser *Parser) parseAnonymousField(pkgName string, field *ast.Field) (map[string]spec.Schema, []string, error) {
	properties := make(map[string]spec.Schema)

	fullTypeName := ""
	switch ftype := field.Type.(type) {
	case *ast.Ident:
		fullTypeName = ftype.Name
	case *ast.StarExpr:
		if ftypeX, ok := ftype.X.(*ast.Ident); ok {
			fullTypeName = ftypeX.Name
		} else if ftypeX, ok := ftype.X.(*ast.SelectorExpr); ok {
			if packageX, ok := ftypeX.X.(*ast.Ident); ok {
				fullTypeName = fmt.Sprintf("%s.%s", packageX.Name, ftypeX.Sel.Name)
			}
		} else {
			Printf("Composite field type of '%T' is unhandle by parser. Skipping", ftype)
			return properties, []string{}, nil
		}
	default:
		Printf("Field type of '%T' is unsupported. Skipping", ftype)
		return properties, []string{}, nil
	}

	typeName := fullTypeName
	if splits := strings.Split(fullTypeName, "."); len(splits) > 1 {
		pkgName = splits[0]
		typeName = splits[1]
	}

	typeSpec := parser.TypeDefinitions[pkgName][typeName]
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
		properties[typeName] = schema
	default:
		Printf("Can't extract properties from a schema of type '%s'", schemaType)
	}

	return properties, schema.SchemaProps.Required, nil
}

func (parser *Parser) parseField(field *ast.Field) (*structField, error) {
	prop := getPropertyName(field.Type, parser)
	if len(prop.ArrayType) == 0 {
		CheckSchemaType(prop.SchemaType)
	} else {
		CheckSchemaType("array")
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
				} else if newSchemaType == "primitive" {
					newSchemaType = parts[1]
					newArrayType = parts[1]
				}
			}

			CheckSchemaType(newSchemaType)
			CheckSchemaType(newArrayType)
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
			structField.enums = append(structField.enums, defineType(enumType, e))
		}
	}
	if defaultTag := structTag.Get("default"); defaultTag != "" {
		structField.defaultValue = defineType(structField.schemaType, defaultTag)
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

func (parser *Parser) visit(path string, f os.FileInfo, err error) error {
	if err := parser.Skip(path, f); err != nil {
		return err
	}

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
