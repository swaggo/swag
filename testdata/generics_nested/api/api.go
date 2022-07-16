package api

import (
	"net/http"

	"github.com/swaggo/swag/testdata/generics_nested/web"
)

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Success 200 {object} web.GenericListResponse[web.Post]
// @Success 201 {object} web.GenericListResponse[web.GenericListResponse[web.Post]]
// @Success 202 {object} web.GenericListResponseMulti[web.Post, web.GenericListResponse[web.Post]]
// @Success 222 {object} web.GenericListResponseMulti[web.GenericListResponse[web.Post], web.Post]
// @Router /posts [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericListResponse[web.Post]{}
}
