package api

import (
	"net/http"

	_ "github.com/swaggo/swag/v2/testdata/conflict_name/model"
)

// @Description  Check if Health  of service it's OK!
// @Router /health [get]
// @x-codeSamples file
func Get1(w http.ResponseWriter, r *http.Request) {

}
