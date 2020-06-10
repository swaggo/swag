package api

import (
	"github.com/gin-gonic/gin"
)

// @Description get Foo
// @ID get-foo
// @Success 200 {string} string
// @Router /testapi/get-foo [get]
func GetFoo(c *gin.Context) {}

// @Description post Bar
// @ID get-foo
// @Success 200 {string} string
// @Router /testapi/post-bar [post]
func PostBar(c *gin.Context) {}
