package main

import (
	"net/http"

	"github.com/nguyennm96/swag/v2/example/basic/api"
	"github.com/nguyennm96/swag/v2/testdata/invalid_external_pkg/invalid"
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

// @host petstore.swagger.io
// @BasePath /v2
func main() {
	invalid.Foo()
	http.HandleFunc("/testapi/upload", api.Upload)
	http.ListenAndServe(":8080", nil)
}
