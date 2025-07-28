package api

import (
	"net/http"
)

// @Summary Endpoint A
// @Description This is a mock endpoint description \
// @Description which is long and descriptions that \
// @Description end with backslash do not add a new line.
// @Description This sentence is in a new line.
// @Description
// @Description And this have an empty line above it.
// @Description Lorem ipsum dolor sit amet \
// @Description consectetur adipiscing elit, \
// @Description sed do eiusmod tempor incididunt \
// @Description ut labore et dolore magna aliqua.
// @Success 200
// @Router /a [get]
func EndpointA(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// @Summary Endpoint B
// @Description Something something.
// @Description
// @Description A new line, \
// @Description continue to the line.
// @Success 200
// @Router /b [get]
func EndpointB(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
