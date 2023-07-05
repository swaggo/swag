package global_security

import (
	"net/http"

	"github.com/swaggo/swag/testdata/global_security/api"
)

// @title Swagger Example API
// @version 1.0

// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name Authorization

// @security APIKeyAuth
func main() {
	http.HandleFunc("/testapi/application", api.GetApplication)
	http.ListenAndServe(":8080", nil)
}
