package api

import (
	"net/http"

	"github.com/swaggo/swag/testdata/generics_nested/types"
	"github.com/swaggo/swag/testdata/generics_nested/web"
)

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Param   data	body	web.GenericNestedBody[web.GenericInnerType[types.Post]]	true	"Some ID"
// @Success 200 {object} web.GenericNestedResponse[types.Post]
// @Success 201 {object} web.GenericNestedResponse[web.GenericInnerType[types.Post]]
// @Success 202 {object} web.GenericNestedResponseMulti[types.Post, web.GenericInnerMultiType[types.Post, types.Post]]
// @Success 203 {object} web.GenericNestedResponseMulti[types.Post, web.GenericInnerMultiType[types.Post, web.GenericInnerType[types.Post]]]
// @Success 222 {object} web.GenericNestedResponseMulti[web.GenericInnerType[types.Post], types.Post]
// @Router /posts [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericNestedResponse[types.Post]{}
}

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Param   data	body	web.GenericNestedBody[web.GenericInnerType[[]types.Post]]	true	"Some ID"
// @Success 200 {object} web.GenericNestedResponse[[]types.Post]
// @Success 201 {object} web.GenericNestedResponse[[]web.GenericInnerType[types.Post]]
// @Success 202 {object} web.GenericNestedResponse[[]web.GenericInnerType[[]types.Post]]
// @Success 203 {object} web.GenericNestedResponseMulti[[]types.Post, web.GenericInnerMultiType[[]types.Post, types.Post]]
// @Success 204 {object} web.GenericNestedResponseMulti[[]types.Post, []web.GenericInnerMultiType[[]types.Post, types.Post]]
// @Success 205 {object} web.GenericNestedResponseMulti[types.Post, web.GenericInnerMultiType[types.Post, []web.GenericInnerType[[][]types.Post]]]
// @Success 222 {object} web.GenericNestedResponseMulti[web.GenericInnerType[[]types.Post], []types.Post]
// @Router /posts-multis/ [get]
func GetPostArray(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericNestedResponse[types.Post]{}
}
