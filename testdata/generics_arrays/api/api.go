package api

import (
	"net/http"

	"github.com/swaggo/swag/testdata/generics_arrays/web"
)

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Param   data        body   web.GenericListBody[web.Post]    true  "Some ID"
// @Success 200 {object} web.GenericListResponse[web.Post]
// @Success 222 {object} web.GenericListResponseMulti[web.Post, web.Post]
// @Router /posts [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericListResponseMulti[web.Post, web.Post]{}
}

// @Summary Add new pets to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data        body   web.GenericListBodyMulti[web.Post, web.Post] true  "Some ID"
// @Success 200 {object} web.GenericListResponse[web.Post]
// @Success 222 {object} web.GenericListResponseMulti[web.Post, web.Post]
// @Router /posts-multi [get]
func GetPostMulti(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericListResponseMulti[web.Post, web.Post]{}
}

// @Summary Add new pets to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data        body   web.GenericListBodyMulti[web.Post, []web.Post] true  "Some ID"
// @Success 200 {object} web.GenericListResponse[[]web.Post]
// @Success 222 {object} web.GenericListResponseMulti[web.Post, []web.Post]
// @Router /posts-multis [get]
func GetPostArray(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericListResponseMulti[web.Post, []web.Post]{}
}
