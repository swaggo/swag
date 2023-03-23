package composition

import (
	"net/http"

	"github.com/swaggo/swag/testdata/composition/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server
// @termsOfService http://swagger.io/terms/

// @host petstore.swagger.io
// @BasePath /v2

func main() {
	http.handleFunc("/testapi/get-foo", api.GetFoo)
	http.handleFunc("/testapi/get-bar", api.GetBar)
	http.handleFunc("/testapi/get-foobar", api.GetFooBar)
	http.handleFunc("/testapi/get-foobar-pointer", api.GetFooBarPointer)
	http.handleFunc("/testapi/get-barmap", api.GetBarMap)
	http.ListenAndServe(":8080", nil)
}
