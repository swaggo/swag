package composition

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/testdata/composition/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server
// @termsOfService http://swagger.io/terms/

// @host petstore.swagger.io
// @BasePath /v2

func main() {
	r := gin.New()
	r.GET("/testapi/get-foo", api.GetFoo)
	r.GET("/testapi/get-bar", api.GetBar)
	r.GET("/testapi/get-foobar", api.GetFooBar)
	r.GET("/testapi/get-foobar-pointer", api.GetFooBarPointer)
	r.GET("/testapi/get-barmap", api.GetBarMap)
	r.Run()
}
