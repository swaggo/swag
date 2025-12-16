package api

import (
	_ "github.com/swaggo/swag/testdata/conflict_name/model2"
	"net/http"
)

// @Tags Health
// @Description Check if Health  of service it's OK!
// @ID health2
// @Accept  json
// @Produce  json
// @Success 200 {object} model.ErrorsResponse
// @Router /health2 [get]
func Get2(w http.ResponseWriter, r *http.Request) {

}
