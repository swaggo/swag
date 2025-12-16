package composition

import (
	"net/http"

	"github.com/swaggo/swag/testdata/duplicated2/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server
// @termsOfService http://swagger.io/terms/

// @host petstore.swagger.io
// @BasePath /v2

func main() {
	http.HandleFunc("/testapi/put-foo", api.PutFoo)
	http.HandleFunc("/testapi/head-foo", api.HeadFoo)
	http.HandleFunc("/testapi/options-foo", api.OptionsFoo)
	http.HandleFunc("/testapi/patch-foo", api.PatchFoo)
	http.HandleFunc("/testapi/delete-foo", api.DeleteFoo)
	http.ListenAndServe(":8080", nil)
}
