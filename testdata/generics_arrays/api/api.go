package api

import (
	"net/http"

	"github.com/extrame/swag/testdata/generics_arrays/web"
)

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Success 200 {object} web.GenericListResponse[web.Post]
// @Success 222 {object} web.GenericListResponseMulti[web.Post, web.Post]
// @Router /posts [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericListResponse[web.Post]{}
}
