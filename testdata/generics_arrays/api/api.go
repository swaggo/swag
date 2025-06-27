package api

import (
	"net/http"

	"github.com/swaggo/swag/testdata/generics_arrays/types"
	"github.com/swaggo/swag/testdata/generics_arrays/web"
)

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Param   data        body   web.GenericListBody[types.Post]    true  "Some ID"
// @Success 200 {object} web.GenericListResponse[types.Post]
// @Success 222 {object} web.GenericListResponseMulti[types.Post, types.Post]
// @Router /posts [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericListResponseMulti[types.Post, types.Post]{}
}

// @Summary Add new pets to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data        body   web.GenericListBodyMulti[types.Post, types.Post] true  "Some ID"
// @Success 200 {object} web.GenericListResponse[types.Post]
// @Success 222 {object} web.GenericListResponseMulti[types.Post, types.Post]
// @Router /posts-multi [get]
func GetPostMulti(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericListResponseMulti[types.Post, types.Post]{}
}

// @Summary Add new pets to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data        body   web.GenericListBodyMulti[types.Post, []types.Post] true  "Some ID"
// @Success 200 {object} web.GenericListResponse[[]types.Post]
// @Success 222 {object} web.GenericListResponseMulti[types.Post, []types.Post]
// @Router /posts-multis [get]
func GetPostArray(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericListResponseMulti[types.Post, []types.Post]{}
}
