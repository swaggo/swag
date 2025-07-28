package main

import (
	"net/http"

	"github.com/swaggo/swag/testdata/description_escape_new_line/api"
)

// @title Swagger Example API
// @version 1.0
// @description Example long description \
// @description that should not be split \
// @description into multiple lines.
// @description This is a new line that\
// @description escapes new line without\
// @description adding a whitespace.
// @description
// @description Another line that has an \
// @description empty line above it.
// @host localhost:8080
func main() {
	http.HandleFunc("/a", api.EndpointA)
	http.HandleFunc("/b", api.EndpointB)
	http.ListenAndServe(":8080", nil)
}
