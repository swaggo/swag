package api

import (
	"net/http"
)

// GetPosts
// @Summary Test Generics with multi level nesting
// @Description Test one of the edge cases found in generics
// @Accept  json
// @Produce  json
// @Success 200 {object} web.TestResponse
// @Router /use-struct-and-generics-with-multi-level-nesting [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {

}
