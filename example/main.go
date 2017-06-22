package main

import (
	"github.com/swag-gonic/swag/example/api"
	"github.com/gin-gonic/gin"
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
	//hello.Hello()
	r.GET("/")
	r.Run()

}
