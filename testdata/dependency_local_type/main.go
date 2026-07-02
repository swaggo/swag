package dependency_local_type

import (
	"net/http"

	"github.com/swaggo/swag/testdata/dependency_local_type/api"
)

// @title Dependency Local Type API
// @version 1.0
// @description Test dependency package local type resolution.
// @host example.com
// @BasePath /
func main() {
	http.HandleFunc("/test", api.Get)
}
