package parse

import (
	"github.com/easonlin404/gin-swagger/spec"
	goparser "go/parser"
	"go/token"
	"log"
	"strings"
)

type Parser struct {
	spec *spec.SwaggerSpec
}

func New() *Parser {

	parser := &Parser{
		spec: spec.New(),
	}
	return parser
}

//Read web/main.go to get General info
func (parser *Parser) ParseGeneralApiInfo(mainApiFile string) {

	fileSet := token.NewFileSet()
	fileTree, err := goparser.ParseFile(fileSet, mainApiFile, nil, goparser.ParseComments)
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
				case "@basepath":
					parser.spec.BasePath = strings.TrimSpace(commentLine[len(attribute):])
				}
			}
		}
	}
}
