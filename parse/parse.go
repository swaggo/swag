package parse

import (
	goparser "go/parser"
	"go/token"
	"log"
	"strings"
	openapi "github.com/go-openapi/spec"
	"go/ast"
	"os"
	"path"
	"path/filepath"
)

type Parser struct {
	swagger         *openapi.Swagger
	files           map[string]*ast.File
	TypeDefinitions map[string]map[string]*ast.TypeSpec
}

func New() *Parser {
	parser := &Parser{

		swagger: &openapi.Swagger{
			SwaggerProps: openapi.SwaggerProps{
				Info: &openapi.Info{
					InfoProps: openapi.InfoProps{
						Contact: &openapi.ContactInfo{},
						License: &openapi.License{},
					},
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

	parser.ParseGeneralApiInfo(path.Join(searchDir, mainApiFile))

	parser.GetAllGoFileInfo(searchDir)
	for _, astFile := range parser.files {
		parser.ParseType(astFile)
	}

}

// ParseGeneralApiInfo parses general api info for gived mainApiFile path
func (parser *Parser) ParseGeneralApiInfo(mainApiFile string) {

	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, mainApiFile, nil, goparser.ParseComments)

	if err != nil {
		log.Panicf("ParseGeneralApiInfo occur error:%+v", err)
	}

	log.Printf("package name:%+v", fileTree.Name)
	log.Printf("imports in this file:%+v", fileTree.Imports)
	for _, importSpec := range fileTree.Imports {
		log.Printf("importSpec:%+v", importSpec.Name)
		log.Printf("importSpec:%+v", importSpec.Path)
	}
	log.Printf(" position of 'package' keyword:%+v", fileTree.Package)

	if err != nil {
		log.Fatalf("Can not parse general API information: %v\n", err)
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

func (parser *Parser) ParseRouterApiInfo() {
	//TODO: add in parser.spec

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
