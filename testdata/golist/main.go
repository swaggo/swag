package main

import (
	"net/http"

	"github.com/swaggo/swag/example/basic/api"
	goapi "github.com/swaggo/swag/testdata/golist/api"
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

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @query.collection.format multi
// @host petstore.swagger.io
// @BasePath /v2
func main() {
	goapi.PrintInt(10, 5)
	http.HandleFunc("/testapi/get-string-by-int/", api.GetStringByInt)
	http.HandleFunc("/testapi/get-struct-array-by-string/", api.GetStructArrayByString)
	http.HandleFunc("/testapi/upload", api.Upload)
	http.HandleFunc("/testapi/foo", goapi.GetFoo)
	http.ListenAndServe(":8080", nil)
}
