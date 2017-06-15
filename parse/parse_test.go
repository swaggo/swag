package parse

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestNew(t *testing.T) {
	New()
}

var expected = `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        }
    },
    "host": "http://easonlin404.github.com",
    "basePath": "petstore",
    "paths": null
}`

func TestParser_ParseGeneralApiInfo(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)
	p := New()
	p.ParseGeneralApiInfo(path.Join(gopath, "src", "github.com/easonlin404/gin-swagger/example/main.go"))

	fmt.Println("%+v", p.swagger)
	fmt.Printf("%+v\n", p.swagger)

	assert.Equal(t, "2.0", p.swagger.Swagger)
	assert.Equal(t, "Swagger Example API", p.swagger.Info.Title)
	assert.Equal(t, "This is a sample server Petstore server.", p.swagger.Info.Description)
	assert.Equal(t, "http://swagger.io/terms/", p.swagger.Info.TermsOfService)
	assert.Equal(t, "API Support", p.swagger.Info.Contact.Name)
	assert.Equal(t, "http://www.swagger.io/support", p.swagger.Info.Contact.URL)
	assert.Equal(t, "support@swagger.io", p.swagger.Info.Contact.Email)
	assert.Equal(t, "Apache 2.0", p.swagger.Info.License.Name)
	assert.Equal(t, "http://www.apache.org/licenses/LICENSE-2.0.html", p.swagger.Info.License.URL)
	assert.Equal(t, "http://easonlin404.github.com", p.swagger.Host)
	assert.Equal(t, "petstore", p.swagger.BasePath)

	b, err := json.MarshalIndent(p.swagger, "", "    ")
	if err != nil {
		panic("err")
	}
	assert.Equal(t, expected, string(b))
}

func TestGetAllGoFileInfo(t *testing.T) {
	searchDir := "../example"

	p := New()
	p.GetAllGoFileInfo(searchDir)

	assert.NotEmpty(t, p.files["../example/main.go"])
	assert.NotEmpty(t, p.files["../example/web/handler.go"])
}

func TestParser_ParseType(t *testing.T) {
	searchDir := "../example"

	p := New()
	p.GetAllGoFileInfo(searchDir)

	for _, file := range p.files {
		p.ParseType(file)
	}

	fmt.Printf("%+v", p.TypeDefinitions)
}

func TestParser_ParseApi(t *testing.T) {
	searchDir := "../example"
	p := New()
	p.ParseApi(searchDir)
}
