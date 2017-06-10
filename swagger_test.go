package swagger

import "testing"
import "github.com/gin-gonic/gin"

func TestNew(t *testing.T) {

	New(gin.New().Routes())
}

func TestEngine_Build(t *testing.T) {
	New(gin.New().Routes()).Build()
}

func TestEngine_Routes(t *testing.T) {
	New(gin.New().Routes()).Routes()
}