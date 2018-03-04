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

	if fileTree.Comments != nil {
		for _, comment := range fileTree.Comments {
			for _, commentLine := range strings.Split(comment.Text(), "\n") {
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
		}
	}
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
	parser.parseTypeSpec(pkgName, typeSpec, properties)

	for _, prop := range properties {
		// todo find the pkgName of the property type
		tname := prop.SchemaProps.Type[0]
		if _, ok := parser.TypeDefinitions[pkgName][tname]; ok {
			tspec := parser.TypeDefinitions[pkgName][tname]
			parser.ParseDefinition(pkgName, tspec, tname)
		}
	}

	log.Println("Generating " + refTypeName)
	parser.swagger.Definitions[refTypeName] = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       []string{"object"},
			Properties: properties,
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

func (parser *Parser) parseStruct(pkgName string, field *ast.Field) (properties map[string]spec.Schema) {
	properties = map[string]spec.Schema{}
	name, schemaType, arrayType, exampleValue := parser.parseField(field)
	// TODO: find package of schemaType and/or arrayType
	if _, ok := parser.TypeDefinitions[pkgName][schemaType]; ok { // user type field
		// write definition if not yet present
		parser.ParseDefinition(pkgName, parser.TypeDefinitions[pkgName][schemaType], schemaType)
		properties[name] = spec.Schema{
			SchemaProps: spec.SchemaProps{Type: []string{"object"}, // to avoid swagger validation error
				Ref: spec.Ref{
					Ref: jsonreference.MustCreateRef("#/definitions/" + pkgName + "." + schemaType),
				},
			},
		}
	} else if schemaType == "array" { // array field type
		// if defined -- ref it
		if _, ok := parser.TypeDefinitions[pkgName][arrayType]; ok { // user type in array
			parser.ParseDefinition(pkgName, parser.TypeDefinitions[pkgName][arrayType], arrayType)
			properties[name] = spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:  []string{schemaType},
					Items: &spec.SchemaOrArray{Schema: &spec.Schema{SchemaProps: spec.SchemaProps{Ref: spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/" + pkgName + "." + arrayType)}}}},
				},
			}
		} else { // standard type in array
			properties[name] = spec.Schema{
				SchemaProps: spec.SchemaProps{Type: []string{schemaType},
					Items: &spec.SchemaOrArray{Schema: &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{arrayType}}}}},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: exampleValue},
			}
		}
	} else {
		properties[name] = spec.Schema{
			SchemaProps:        spec.SchemaProps{Type: []string{schemaType}},
			SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: exampleValue},
		}
		nestStruct, ok := field.Type.(*ast.StructType)
		if ok {
			props := map[string]spec.Schema{}
			for _, v := range nestStruct.Fields.List {
				p := parser.parseStruct(pkgName, v)
				for k, v := range p {
					props[k] = v
				}
			}
			properties[name] = spec.Schema{
				SchemaProps:        spec.SchemaProps{Type: []string{schemaType}, Properties: props},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{Example: exampleValue},
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

func (parser *Parser) parseField(field *ast.Field) (propName, schemaType, arrayType string, exampleValue interface{}) {
	schemaType, arrayType = getPropertyName(field)
	if len(arrayType) == 0 {
		CheckSchemaType(schemaType)
	} else {
		CheckSchemaType("array")
	}
	propName = field.Names[0].Name
	if field.Tag != nil {
		// `json:"tag"` -> json:"tag"
		structTag := strings.Replace(field.Tag.Value, "`", "", -1)
		jsonTag := reflect.StructTag(structTag).Get("json")
		if jsonTag != "" {
			propName = jsonTag
		}
		exampleTag := reflect.StructTag(structTag).Get("example")
		if exampleTag != "" {
			exampleValue = defineTypeOfExample(schemaType, exampleTag)
		}
	}
	return
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

// GetAllGoFileInfo gets all Go source files information for gived searchDir.
func (parser *Parser) getAllGoFileInfo(searchDir string) {
	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		//exclude vendor folder
		if ext := filepath.Ext(path); ext == ".go" && !strings.Contains(string(os.PathSeparator)+path, string(os.PathSeparator)+"vendor"+string(os.PathSeparator)) {
			fset := token.NewFileSet() // positions are relative to fset
			astFile, err := goparser.ParseFile(fset, path, nil, goparser.ParseComments)

			if err != nil {
				log.Panicf("ParseFile panic:%+v", err)
			}

			parser.files[path] = astFile

		}
		return nil
	})
}

// GetSwagger returns *spec.Swagger which is the root document object for the API specification.
func (parser *Parser) GetSwagger() *spec.Swagger {
	return parser.swagger
}
