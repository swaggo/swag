package main

import (
	"net/http"

	"github.com/rampnow-io/swag/testdata/generics_function_scoped/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server.
// @host localhost:8080
// @basePath /api
func main() {
	http.HandleFunc("/", api.GetGeneric)
	http.HandleFunc("/renamed", api.GetGenericRenamed)
	http.HandleFunc("/multi", api.GetGenericMulti)
	http.HandleFunc("/multi-renamed", api.GetGenericMulti)
	http.ListenAndServe(":8080", nil)
}
