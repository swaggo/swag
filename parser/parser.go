package parser

import (
	"github.com/go-openapi/spec"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Parser struct {
	swagger         *spec.Swagger
	files           map[string]*ast.File                // map[real_go_file_path][astFile]
	TypeDefinitions map[string]map[string]*ast.TypeSpec // map [package name][type name][*ast.TypeSpec]
	registerTypes   map[string]*ast.TypeSpec            //map [refTypeName][*ast.TypeSpec]
}

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

func (parser *Parser) ParseApi(searchDir string, mainApiFile string) {
	parser.GetAllGoFileInfo(searchDir)
	parser.ParseGeneralApiInfo(path.Join(searchDir, mainApiFile))

	for _, astFile := range parser.files {
		parser.ParseType(astFile)
	}

	for _, astFile := range parser.files {
		parser.parseRouterApiInfo(astFile)
	}

	parser.ParseDefinitions()
}

// ParseGeneralApiInfo parses general api info for gived mainApiFile path
func (parser *Parser) ParseGeneralApiInfo(mainApiFile string) {
	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, mainApiFile, nil, goparser.ParseComments)

	if err != nil {
		log.Panicf("ParseGeneralApiInfo occur error:%+v", err)
	}

	parser.swagger.BasePath = "{{.}}"
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
				}
			}
		}
	}
}

func (parser *Parser) parseRouterApiInfo(astFile *ast.File) {
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

				pathItem := spec.PathItem{}
				switch strings.ToUpper(operation.HttpMethod) {
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

func (parser *Parser) ParseDefinitions() {
	for refTypeName, typeSpec := range parser.registerTypes {
		var properties map[string]spec.Schema
		properties = make(map[string]spec.Schema)

		switch typeSpec.Type.(type) {
		case *ast.StructType:
			structDecl := typeSpec.Type.(*ast.StructType)
			fields := structDecl.Fields.List

			for _, field := range fields {
				name := field.Names[0].Name
				propName := getPropertyName(field)
				properties[name] = spec.Schema{
					SchemaProps: spec.SchemaProps{Type: []string{propName}},
				}
			}

		case *ast.ArrayType:
			log.Panic("ParseDefinitions not supported 'array' yet.")
		case *ast.InterfaceType:
			log.Panic("ParseDefinitions not supported 'array' yet.")
		case *ast.MapType:
			log.Panic("ParseDefinitions not supported 'array' yet.")
		}

		parser.swagger.Definitions[refTypeName] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:       []string{"object"},
				Properties: properties,
			},
		}

	}
}

func (parser *Parser) GetAllGoFileInfo(searchDir string) {
	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		//exclude vendor folder
		if ext := filepath.Ext(path); ext == ".go" && !strings.Contains(path, "/vendor") {
			astFile, err := goparser.ParseFile(token.NewFileSet(), path, nil, goparser.ParseComments)
			if err != nil {
				log.Panicf("ParseFile panic:%+v", err)
			}
			parser.files[path] = astFile

		}
		return nil
	})
}


func (parser *Parser) GetSwagger() *spec.Swagger {
	return parser.swagger
}