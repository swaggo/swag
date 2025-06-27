package alias_import

import (
	"net/http"

	"github.com/swaggo/swag/testdata/alias_import/api"
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
	http.HandleFunc("/testapi/application", api.GetApplication)
	http.ListenAndServe(":8080", nil)
}
