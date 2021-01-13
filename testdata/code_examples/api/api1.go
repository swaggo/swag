package api

import (
	_ "github.com/Nerzal/swag/testdata/conflict_name/model"
	"net/http"
)

// @Description  Check if Health  of service it's OK!
// @Router /health [get]
// @x-codeSamples file
func Get1(w http.ResponseWriter, r *http.Request) {

}
