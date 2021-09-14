package api

import (
	"net/http"
)

// @Description duplicate_route1
// @Router /testapi/endpoint [get]
func Function(w http.ResponseWriter, r *http.Request) {
}

// @Description route1
// @Router /testapi/route1 [get]
func Function(w http.ResponseWriter, r *http.Request) {
}
