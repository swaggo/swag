package main

import (
	"net/http"

	"github.com/swaggo/swag/v2/testdata/v3/simple/api"
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

// @servers.url {scheme}://{host}:{port}
// @servers.description Test Petstore server.
// @servers.variables.enum scheme http
// @servers.variables.enum scheme https
// @servers.variables.default scheme https
// @servers.variables.default host test.petstore.com
// @servers.variables.default port 443

// @servers.url https://petstore.com/v3
// @servers.description Production Petstore server.
func main() {
	http.HandleFunc("/testapi/get-string-by-int/", api.GetStringByInt)
	http.HandleFunc("/testapi/get-struct-array-by-string/", api.GetStructArrayByString)
	http.HandleFunc("/testapi/upload", api.Upload)

	http.ListenAndServe(":8080", nil)
}
