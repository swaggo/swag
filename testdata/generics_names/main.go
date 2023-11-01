package main

import (
	"net/http"

	"github.com/nguyennm96/swag/v2/testdata/generics_names/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @host localhost:4000
// @basePath /api
func main() {
	http.HandleFunc("/posts/", api.GetPost)
	http.HandleFunc("/posts-multi/", api.GetPostMulti)
	http.HandleFunc("/posts-multis/", api.GetPostArray)
	http.ListenAndServe(":8080", nil)
}
