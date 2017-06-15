package parse

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	"fmt"
)

func TestNew(t *testing.T) {
	New()
}

func TestParser_ParseGeneralApiInfo(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)
	p := New()
	p.ParseGeneralApiInfo(path.Join(gopath, "src", "github.com/easonlin404/gin-swagger/example/main.go"))

	fmt.Println("%+v", p.spec)
	fmt.Printf("%+v\n", p.spec)

	assert.Equal(t, "2.0", p.spec.Swagger)
	assert.Equal(t, "Swagger Example API", p.spec.Info.Title)
	assert.Equal(t, "This is a sample server Petstore server.", p.spec.Info.Description)
	assert.Equal(t, "http://swagger.io/terms/", p.spec.Info.TermsOfService)
	assert.Equal(t, "API Support", p.spec.Info.Contact.Name)
	assert.Equal(t, "http://www.swagger.io/support", p.spec.Info.Contact.URL)
	assert.Equal(t, "support@swagger.io", p.spec.Info.Contact.Email)
	assert.Equal(t, "Apache 2.0", p.spec.Info.License.Name)
	assert.Equal(t, "http://www.apache.org/licenses/LICENSE-2.0.html", p.spec.Info.License.URL)
	assert.Equal(t, "http://easonlin404.github.com", p.spec.Host)
	assert.Equal(t, "petstore", p.spec.BasePath)
}


func TestGetAllGoFileInfo(t *testing.T) {
	searchDir := "../example"

	p:=New()
	p.GetAllGoFileInfo(searchDir)

	assert.NotEmpty(t, p.files["../example/main.go"])
	assert.NotEmpty(t, p.files["../example/web/handler.go"])
}

func TestParser_ParseType(t *testing.T) {
	searchDir := "../example"

	p:=New()
	p.GetAllGoFileInfo(searchDir)

	for _,file :=range p.files{
		p.ParseType(file)
	}

	fmt.Printf("%+v",p.TypeDefinitions)
}

func TestParser_ParseApi(t *testing.T) {
	searchDir := "../example"
	p:=New()
	p.ParseApi(searchDir)
}