package api

import (
	"net/http"

	_ "github.com/nguyennm96/swag/v2/testdata/conflict_name/model"
)

// @Tags Health
// @Description  Check if Health  of service it's OK!
// @ID health
// @Accept  json
// @Produce  json
// @Success 200 {object} model.ErrorsResponse
// @Router /health [get]
func Get1(w http.ResponseWriter, r *http.Request) {

}
