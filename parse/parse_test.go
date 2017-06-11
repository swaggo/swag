package parse

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"

	_ "fmt"
)

func TestNew(t *testing.T) {
	New()
}

func TestParser_ParseGeneralApiInfo(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	assert.NotNil(t, gopath)

	New().ParseGeneralApiInfo(path.Join(gopath, "src", "github.com/yvasiyarov/swagger/example/web/main.go"))

}
