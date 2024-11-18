package main

import (
	"net/http"

	"github.com/rampnow-io/swag/testdata/state/api"
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

// @hostState admin petstore-admin.swagger.io
// @hostState user petstore-user.swagger.io
// @BasePath /v3
func main() {
	state := "admin" // "admin" or "user"
	switch state {
	case "admin":
		http.HandleFunc("/admin/testapi/get-string-by-int/", api.GetStringByInt)
		http.HandleFunc("/admin/testapi/get-struct-array-by-string/", api.GetStructArrayByString)
		http.HandleFunc("/admin/testapi/upload", api.Upload)
		http.ListenAndServe(":8080", nil)
	case "user":
		http.HandleFunc("/testapi/get-string-by-int/", api.GetStringByIntUser)
		http.HandleFunc("/testapi/get-struct-array-by-string/", api.GetStructArrayByStringUser)
		http.HandleFunc("/testapi/upload", api.UploadUser)
		http.ListenAndServe(":8080", nil)
	}
}
