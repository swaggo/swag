package swag

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
)

// Parser implements a parser for Go source files.
type Parser struct {
	// swagger represents the root document object for the API specification
	swagger *spec.Swagger

	//files is a map that stores map[real_go_file_path][astFile]
	files map[string]*ast.File

	// TypeDefinitions is a map that stores [package name][type name][*ast.TypeSpec]
	TypeDefinitions map[string]map[string]*ast.TypeSpec

	//registerTypes is a map that stores [refTypeName][*ast.TypeSpec]
	registerTypes map[string]*ast.TypeSpec

	PropNamingStrategy string
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
		files:           make(map[string]*ast.File),
		TypeDefinitions: make(map[string]map[string]*ast.TypeSpec),
		registerTypes:   make(map[string]*ast.TypeSpec),
	}
	return parser
}

// ParseAPI parses general api info for gived searchDir and mainAPIFile
func (parser *Parser) ParseAPI(searchDir string, mainAPIFile string) {
	log.Println("Generate general API Info")
	parser.getAllGoFileInfo(searchDir)
	parser.ParseGeneralAPIInfo(path.Join(searchDir, mainAPIFile))

	for _, astFile := range parser.files {
		parser.ParseType(astFile)
	}

	for _, astFile := range parser.files {
		parser.ParseRouterAPIInfo(astFile)
	}

	parser.ParseDefinitions()
}

// ParseGeneralAPIInfo parses general api info for gived mainAPIFile path
func (parser *Parser) ParseGeneralAPIInfo(mainAPIFile string) {
	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, mainAPIFile, nil, goparser.ParseComments)

	if err != nil {
		log.Panicf("ParseGeneralApiInfo occur error:%+v", err)
	}

	parser.swagger.Swagger = "2.0"
	securityMap := map[string]*spec.SecurityScheme{}

	if fileTree.Comments != nil {
		for _, comment := range fileTree.Comments {
			comments := strings.Split(comment.Text(), "\n")
			for _, commentLine := range comments {
				attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
				switch attribute {
				case "@version":
					parser.swagger.Info.Version = strings.TrimSpace(commentLine[len(attribute):])
				case "@title":
					parser.swagger.Info.Title = strings.TrimSpace(commentLine[len(attribute):])
				case "@description":
					parser.swagger.Info.Description = strings.TrimSpace(commentLine[len(attribute):])
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
					parser.swagger.Schemes = GetSchemes(commentLine)
				}
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
						log.Panic("@securitydefinitions.apikey is @name and @in required")
					}
					securityMap[strings.TrimSpace(comments[i][len(attribute):])] = spec.APIKeyAuth(attrMap["@name"], attrMap["@in"])
				case "@securitydefinitions.oauth2.application":
					attrMap := map[string]string{}
					scopes := map[string]string{}
					for _, v := range comments[i+1:] {
						securityAttr := strings.ToLower(strings.Split(v, " ")[0])
						if securityAttr == "@tokenurl" {
							attrMap[securityAttr] = strings.TrimSpace(v[len(securityAttr):])
						} else if isExistsScope(securityAttr) {
							scopes[getScopeScheme(securityAttr)] = v[len(securityAttr):]
						}
						// next securityDefinitions
						if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
							break
						}
					}
					if len(attrMap) != 1 {
						log.Panic("@securitydefinitions.oauth2.application is @tokenUrl required")
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
						} else if isExistsScope(securityAttr) {
							scopes[getScopeScheme(securityAttr)] = v[len(securityAttr):]
						}
						// next securityDefinitions
						if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
							break
						}
					}
					if len(attrMap) != 1 {
						log.Panic("@securitydefinitions.oauth2.implicit is @authorizationUrl required")
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
						} else if isExistsScope(securityAttr) {
							scopes[getScopeScheme(securityAttr)] = v[len(securityAttr):]
						}
						// next securityDefinitions
						if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
							break
						}
					}
					if len(attrMap) != 1 {
						log.Panic("@securitydefinitions.oauth2.password is @tokenUrl required")
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
						} else if isExistsScope(securityAttr) {
							scopes[getScopeScheme(securityAttr)] = v[len(securityAttr):]
						}
						// next securityDefinitions
						if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
							break
						}
					}
					if len(attrMap) != 2 {
						log.Panic("@securitydefinitions.oauth2.accessCode is @tokenUrl and @authorizationUrl required")
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
}

func getScopeScheme(scope string) string {
	scopeValue := scope[strings.Index(scope, "@scope."):]
	if scopeValue == "" {
		panic("@scope is empty")
	}
	return scope[len("@scope."):]
}

func isExistsScope(scope string) bool {
	s := strings.Fields(scope)
	for _, v := range s {
		if strings.Index(v, "@scope.") != -1 {
			if strings.Index(v, ",") != -1 {
				panic("@scope can't use comma(,) get=" + v)
			}
		}
	}
	return strings.Index(scope, "@scope.") != -1
}

// GetSchemes parses swagger schemes for gived commentLine
func GetSchemes(commentLine string) []string {
	attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
	return strings.Split(strings.TrimSpace(commentLine[len(attribute):]), " ")
}

// ParseRouterAPIInfo parses router api info for gived astFile
func (parser *Parser) ParseRouterAPIInfo(astFile *ast.File) {
	for _, astDescription := range astFile.Decls {
		switch astDeclaration := astDescription.(type) {
		case *ast.FuncDecl:
			if astDeclaration.Doc != nil && astDeclaration.Doc.List != nil {
				operation := NewOperation() //for per 'function' comment, create a new 'Operation' object
				operation.parser = parser
				for _, comment := range astDeclaration.Doc.List {
					if err := operation.ParseComment(comment.Text); err != nil {
						log.Panicf("ParseComment panic:%+v", err)
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
}

// ParseType parses type info for gived astFile
func (parser *Parser) ParseType(astFile *ast.File) {
	if _, ok := parser.TypeDefinitions[astFile.Name.String()]; !ok {
		parser.TypeDefinitions[astFile.Name.String()] = make(map[string]*ast.TypeSpec)
	}

	for _, astDeclaration := range astFile.Decls {
		if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
			for _, astSpec := range generalDeclaration.Specs {
				if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
					parser.TypeDefinitions[astFile.Name.String()][typeSpec.Name.String()] = typeSpec
				}
			}
		}
	}
}

// ParseDefinitions parses Swagger Api definitions
func (parser *Parser) ParseDefinitions() {
	for refTypeName, typeSpec := range parser.registerTypes {
		ss := strings.Split(refTypeName, ".")
		pkgName := ss[0]
		parser.ParseDefinition(pkgName, typeSpec, typeSpec.Name.Name)
	}
}

var structStacks []string

// isNotRecurringNestStruct check if a structure that is not a not repeating
func isNotRecurringNestStruct(refTypeName string, structStacks []string) bool {
	if len(structStacks) <= 0 {
		return true
	}
	startStruct := structStacks[0]
	for _, v := range structStacks[1:] {
		if startStruct == v {
			return false
		}
	}
	return true
}

// ParseDefinition TODO: NEEDS COMMENT INFO
func (parser *Parser) ParseDefinition(pkgName string, typeSpec *ast.TypeSpec, typeName string) {
	var refTypeName string
	if len(pkgName) > 0 {
		refTypeName = pkgName + "." + typeName
	} else {
		refTypeName = typeName
	}
	if _, already := parser.swagger.Definitions[refTypeName]; already {
		log.Println("Skipping '" + refTypeName + "', already present.")
		return
	}
	properties := make(map[string]spec.Schema)
	// stop repetitive structural parsing
	if isNotRecurringNestStruct(refTypeName, structStacks) {
		structStacks = append(structStacks, refTypeName)
		parser.parseTypeSpec(pkgName, typeSpec, properties)
	}
	structStacks = []string{}

	requiredFields := make([]string, 0)
	for k, prop := range properties {
		// todo find the pkgName of the property type
		tname := prop.SchemaProps.Type[0]
		if _, ok := parser.TypeDefinitions[pkgName][tname]; ok {
			tspec := parser.TypeDefinitions[pkgName][tname]
			parser.ParseDefinition(pkgName, tspec, tname)
		}
		if tname != "object" {
			requiredFields = append(requiredFields, prop.SchemaProps.Required...)
			prop.SchemaProps.Required = make([]string, 0)
		}
		properties[k] = prop
	}
	log.Println("Generating " + refTypeName)
	parser.swagger.Definitions[refTypeName] = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       []string{"object"},
			Properties: properties,
			Required:   requiredFields,
		},
	}
}

func (parser *Parser) parseTypeSpec(pkgName string, typeSpec *ast.TypeSpec, properties map[string]spec.Schema) {
	switch typeSpec.Type.(type) {
	case *ast.StructType:
		structDecl := typeSpec.Type.(*ast.StructType)
		fields := structDecl.Fields.List

		for _, field := range fields {
			if field.Names == nil { //anonymous field
				parser.parseAnonymousField(pkgName, field, properties)
			} else {
				props := parser.parseStruct(pkgName, field)
				for k, v := range props {
					properties[k] = v
				}
			}
		}

	case *ast.ArrayType:
		log.Panic("ParseDefinitions not supported 'Array' yet.")
	case *ast.InterfaceType:
		log.Panic("ParseDefinitions not supported 'Interface' yet.")
	case *ast.MapType:
		log.Panic("ParseDefinitions not supported 'Map' yet.")
	}
}

type structField struct {
	name         string
	schemaType   string
	arrayType    string
	formatType   string
	isRequired   bool
	exampleValue interface{}
}

func (parser *Parser) parseStruct(pkgName string, field *ast.Field) (properties map[string]spec.Schema) {
	properties = map[string]spec.Schema{}
	// name, schemaType, arrayType, formatType, exampleValue :=
	structField := parser.parseField(field)
	if structField.name == "" {
		return
	}
	var desc string
	if field.Doc != nil {
		desc = field.Doc.Text()
	}

	// TODO: find package of schemaType and/or arrayType
	if _, ok := parser.TypeDefinitions[pkgName][structField.schemaType]; ok { // user type field
		// write definition if not yet present
		parser.ParseDefinition(pkgName, parser.TypeDefinitions[pkgName][structField.schemaType], structField.schemaType)
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
			parser.ParseDefinition(pkgName, parser.TypeDefinitions[pkgName][structField.arrayType], structField.arrayType)
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
								Type: []string{structField.arrayType},
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
			},
			SwaggerSchemaProps: spec.SwaggerSchemaProps{
				Example: structField.exampleValue,
			},
		}
		nestStruct, ok := field.Type.(*ast.StructType)
		if ok {
			props := map[string]spec.Schema{}
			nestRequired := make([]string, 0)
			for _, v := range nestStruct.Fields.List {
				p := parser.parseStruct(pkgName, v)
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
				},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					Example: structField.exampleValue,
				},
			}
		}
	}
	return
}

func (parser *Parser) parseAnonymousField(pkgName string, field *ast.Field, properties map[string]spec.Schema) {
	if astTypeIdent, ok := field.Type.(*ast.Ident); ok {
		findPgkName := pkgName
		findBaseTypeName := astTypeIdent.Name
		ss := strings.Split(astTypeIdent.Name, ".")
		if len(ss) > 1 {
			findPgkName = ss[0]
			findBaseTypeName = ss[1]
		}

		baseTypeSpec := parser.TypeDefinitions[findPgkName][findBaseTypeName]
		parser.parseTypeSpec(findPgkName, baseTypeSpec, properties)
	}
}

func (parser *Parser) parseField(field *ast.Field) *structField {
	prop := getPropertyName(field)
	if len(prop.ArrayType) == 0 {
		CheckSchemaType(prop.SchemaType)
	} else {
		CheckSchemaType("array")
	}
	structField := &structField{
		name:       field.Names[0].Name,
		schemaType: prop.SchemaType,
		arrayType:  prop.ArrayType,
	}

	if parser.PropNamingStrategy == "snakecase" {
		// snakecase
		structField.name = toSnakeCase(structField.name)
	} else if parser.PropNamingStrategy != "uppercamelcase" {
		// default
		structField.name = toLowerCamelCase(structField.name)
	}

	if field.Tag == nil {
		return structField
	}
	// `json:"tag"` -> json:"tag"
	structTag := strings.Replace(field.Tag.Value, "`", "", -1)
	jsonTag := reflect.StructTag(structTag).Get("json")
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

	exampleTag := reflect.StructTag(structTag).Get("example")
	if exampleTag != "" {
		structField.exampleValue = defineTypeOfExample(structField.schemaType, exampleTag)
	}
	formatTag := reflect.StructTag(structTag).Get("format")
	if formatTag != "" {
		structField.formatType = formatTag
	}
	bindingTag := reflect.StructTag(structTag).Get("binding")
	if bindingTag != "" {
		for _, val := range strings.Split(bindingTag, ",") {
			if val == "required" {
				structField.isRequired = true
				break
			}
		}
	}
	validateTag := reflect.StructTag(structTag).Get("validate")
	if validateTag != "" {
		for _, val := range strings.Split(validateTag, ",") {
			if val == "required" {
				structField.isRequired = true
				break
			}
		}
	}
	return structField
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
func defineTypeOfExample(schemaType string, exampleValue string) interface{} {
	switch schemaType {
	case "string":
		return exampleValue
	case "number":
		v, err := strconv.ParseFloat(exampleValue, 64)
		if err != nil {
			panic(fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err))
		}
		return v
	case "integer":
		v, err := strconv.Atoi(exampleValue)
		if err != nil {
			panic(fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err))
		}
		return v
	case "boolean":
		v, err := strconv.ParseBool(exampleValue)
		if err != nil {
			panic(fmt.Errorf("example value %s can't convert to %s err: %s", exampleValue, schemaType, err))
		}
		return v
	case "array":
		return strings.Split(exampleValue, ",")
	default:
		panic(fmt.Errorf("%s is unsupported type in example value", schemaType))
	}
}

// GetAllGoFileInfo gets all Go source files information for given searchDir.
func (parser *Parser) getAllGoFileInfo(searchDir string) {
	filepath.Walk(searchDir, parser.visit)
}

func (parser *Parser) visit(path string, f os.FileInfo, err error) error {
	if err := Skip(f); err != nil {
		return err
	}

	if ext := filepath.Ext(path); ext == ".go" {
		fset := token.NewFileSet() // positions are relative to fset
		astFile, err := goparser.ParseFile(fset, path, nil, goparser.ParseComments)
		if err != nil {
			log.Panicf("ParseFile panic:%+v", err)
		}

		parser.files[path] = astFile
	}
	return nil
}

// Skip returns filepath.SkipDir error if match vendor and hidden folder
func Skip(f os.FileInfo) error {
	// exclude vendor folder
	if f.IsDir() && f.Name() == "vendor" {
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
