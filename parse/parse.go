package parse

import (
	"github.com/easonlin404/gin-swagger/spec"
	goparser "go/parser"
	"go/token"
	"log"
	"strings"


	"go/ast"
	"os"
	"path/filepath"
)

type Parser struct {
	spec            *spec.SwaggerSpec
	files           map[string]*ast.File
	TypeDefinitions map[string]map[string]*ast.TypeSpec
}

func New() *Parser {
	parser := &Parser{
		spec:  spec.New(),
		files: make(map[string]*ast.File),
		TypeDefinitions:make(map[string]map[string]*ast.TypeSpec),
	}
	return parser
}

func (parser *Parser) GetSpec() *spec.SwaggerSpec {
	return parser.spec
}

// ParseGeneralApiInfo parses general api info for gived mainApiFile path
func (parser *Parser) ParseGeneralApiInfo(mainApiFile string) {

	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, mainApiFile, nil, goparser.ParseComments)

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

	parser.spec.BasePath = "{{.}}"
	parser.spec.Swagger = spec.SwaggerVesion

	if fileTree.Comments != nil {
		for _, comment := range fileTree.Comments {
			for _, commentLine := range strings.Split(comment.Text(), "\n") {
				attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
				switch attribute {
				case "@version":
					parser.spec.Info.Version = strings.TrimSpace(commentLine[len(attribute):])
				case "@title":
					parser.spec.Info.Title = strings.TrimSpace(commentLine[len(attribute):])
				case "@description":
					parser.spec.Info.Description = strings.TrimSpace(commentLine[len(attribute):])
				case "@termsofservice":
					parser.spec.Info.TermsOfService = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact.name":
					parser.spec.Info.Contact.Name = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact.email":
					parser.spec.Info.Contact.Email = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact.url":
					parser.spec.Info.Contact.URL = strings.TrimSpace(commentLine[len(attribute):])
				case "@license.name":
					parser.spec.Info.License.Name = strings.TrimSpace(commentLine[len(attribute):])
				case "@license.url":
					parser.spec.Info.License.URL = strings.TrimSpace(commentLine[len(attribute):])
				case "@host":
					parser.spec.Host = strings.TrimSpace(commentLine[len(attribute):])
				case "@basepath":
					parser.spec.BasePath = strings.TrimSpace(commentLine[len(attribute):])
				}
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

func (parser *Parser) GetAllGoFileInfo(searchDir string) map[string]*ast.File {
	files := make(map[string]*ast.File)

	fileList := []string{}
	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		//exclude vendor folder
		if ext := filepath.Ext(path); ext == ".go" && !strings.Contains(path, "/vendor") {
			astFile, err := goparser.ParseFile(token.NewFileSet(), path, nil, goparser.ParseComments)

			if err != nil {
				log.Panicf("ParseFile panic:%+v", err)
			}

			files[path] = astFile
			fileList = append(fileList, path)
		}
		return nil
	})

	//for _, file := range fileList {
	//	fmt.Println(file)
	//}
	parser.files = files
	return files

}
