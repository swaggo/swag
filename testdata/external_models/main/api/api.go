package api

import (
	_ "github.com/swaggo/swag/testdata/external_models/external"
	"net/http"
)

// GetExternalModels example
// @Summary parse external models
// @Description get string by ID
// @ID get_external_models
// @Accept  json
// @Produce  json
// @Success 200 {string} string	"ok"
// @Failure 400 {object} http.Header "from internal pkg"
// @Failure 404 {object} external.MyError "from external pkg"
// @Router /testapi/external_models [get]
func GetExternalModels(w http.ResponseWriter, r *http.Request) {

}
