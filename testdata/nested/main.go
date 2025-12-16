package composition

import (
	"net/http"

	"github.com/swaggo/swag/testdata/nested/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server
// @termsOfService http://swagger.io/terms/

// @host petstore.swagger.io
// @BasePath /v2

func main() {
	http.HandleFunc("/testapi/get-foo", api.GetFoo)
	http.ListenAndServe(":8080", nil)
}
