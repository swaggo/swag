package main

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/basic/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host petstore.swagger.io
// @BasePath /v2
func main() {
	r := gin.New()
	r.GET("/testapi/get-string-by-int/:some_id", api.GetStringByInt)
	r.GET("//testapi/get-struct-array-by-string/:some_id", api.GetStructArrayByString)
	r.POST("/testapi/upload", api.Upload)
	r.Run()

}
