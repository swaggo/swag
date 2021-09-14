package api

import (
	"net/http"
)

// @Description duplicate_route3
// @Router /testapi/endpoint [get]
func Function(w http.ResponseWriter, r *http.Request) {
}

// @Description route3
// @Router /testapi/route3 [get]
func Function(w http.ResponseWriter, r *http.Request) {
}
