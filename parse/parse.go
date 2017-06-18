package parse

import (
	"github.com/go-openapi/spec"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"net/http"
	"fmt"
)

type Parser struct {
	swagger         *spec.Swagger
	files           map[string]*ast.File                // map[real_go_file_path][astFile]
	TypeDefinitions map[string]map[string]*ast.TypeSpec // map [package name][type name][ast.TypeSpec]
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
					Paths:make( map[string]spec.PathItem),
				},
			},
		},
		files:           make(map[string]*ast.File),
		TypeDefinitions: make(map[string]map[string]*ast.TypeSpec),
	}
	return parser
}

func (parser *Parser) ParseApi(searchDir string) {
	mainApiFile := "./main.go"
	parser.GetAllGoFileInfo(searchDir)
	parser.ParseGeneralApiInfo(path.Join(searchDir, mainApiFile))

	for _, astFile := range parser.files {
		parser.ParseType(astFile)
	}

	for _, astFile := range parser.files {
		parser.parseRouterApiInfo(astFile)
	}
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
				operation := new(Operation) //for per 'function' comment, create a new 'Operation' object
				for _, comment := range astDeclaration.Doc.List {
						operation.ParseComment(comment.Text)
				}
				fmt.Println(operation.HttpMethod)
				pathItem :=spec.PathItem{
				}
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
