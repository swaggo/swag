package api

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/testdata/nested2"
)

type Foo struct {
	Field1      string
	OutsideData *nested2.Body
}

// @Description get Foo
// @ID get-foo
// @Accept json
// @Produce json
// @Success 200 {object} api.Foo
// @Router /testapi/get-foo [get]
func GetFoo(c *gin.Context) {
	//write your code
	var _ = Foo{}
}
