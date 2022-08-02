package api

import (
	"net/http"

	"github.com/swaggo/swag/testdata/generics_nested/web"
)

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Param   data	body	web.GenericNestedBody[web.GenericInnerType[web.Post]]	true	"Some ID"
// @Success 200 {object} web.GenericNestedResponse[web.Post]
// @Success 201 {object} web.GenericNestedResponse[web.GenericInnerType[web.Post]]
// @Success 202 {object} web.GenericNestedResponseMulti[web.Post, web.GenericInnerMultiType[web.Post, web.Post]]
// @Success 203 {object} web.GenericNestedResponseMulti[web.Post, web.GenericInnerMultiType[web.Post, web.GenericInnerType[web.Post]]]
// @Success 222 {object} web.GenericNestedResponseMulti[web.GenericInnerType[web.Post], web.Post]
// @Router /posts [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericNestedResponse[web.Post]{}
}

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Param   data	body	web.GenericNestedBody[web.GenericInnerType[[]web.Post]]	true	"Some ID"
// @Success 200 {object} web.GenericNestedResponse[[]web.Post]
// @Success 201 {object} web.GenericNestedResponse[[]web.GenericInnerType[web.Post]]
// @Success 202 {object} web.GenericNestedResponse[[]web.GenericInnerType[[]web.Post]]
// @Success 203 {object} web.GenericNestedResponseMulti[[]web.Post, web.GenericInnerMultiType[[]web.Post, web.Post]]
// @Success 204 {object} web.GenericNestedResponseMulti[[]web.Post, []web.GenericInnerMultiType[[]web.Post, web.Post]]
// @Success 205 {object} web.GenericNestedResponseMulti[web.Post, web.GenericInnerMultiType[web.Post, []web.GenericInnerType[[][]web.Post]]]
// @Success 222 {object} web.GenericNestedResponseMulti[web.GenericInnerType[[]web.Post], []web.Post]
// @Router /posts-multis/ [get]
func GetPostArray(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericNestedResponse[web.Post]{}
}
