package api

import (
	"net/http"

	_ "github.com/swaggo/swag/testdata/duplicate_route3"
)

// @Description duplicate_route2
// @Router /testapi/endpoint [get]
func Function(w http.ResponseWriter, r *http.Request) {
}

// @Description route2
// @Router /testapi/route2 [get]
func Function(w http.ResponseWriter, r *http.Request) {
}
