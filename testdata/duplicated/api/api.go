package api

import (
	"github.com/gin-gonic/gin"
)

type Foo struct {
	Field1 string `validate:"required"`
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

// @Description get Foo
// @ID get-foo
// @Accept json
// @Produce json
// @Success 200 {object} api.Foo
// @Router /testapi/get-bar [post]
func GetBar(c *gin.Context) {
	//write your code
	var _ = Foo{}
}
