package api

import (
	"net/http"

	"github.com/swaggo/swag/testdata/generics_basic/web"
)

// @Summary Add a new pet to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data        body   web.GenericBody[web.Post]    true  "Some ID"
// @Success 200 {object} web.GenericResponse[web.Post]
// @Success 222 {object} web.GenericResponseMulti[web.Post, web.Post]
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /posts/ [post]
func GetPost(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericResponse[web.Post]{}
}

// @Summary Add new pets to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data	body	web.GenericBodyMulti[web.Post, web.Post]	true	"Some ID"
// @Success 200 {object} web.GenericResponse[web.Post]
// @Success 222 {object} web.GenericResponseMulti[web.Post, web.Post]
// @Router /posts-multi/ [post]
func GetPostMulti(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericResponse[web.Post]{}
}

// @Summary Add new pets to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data	body	web.GenericBodyMulti[[]web.Post, [][]web.Post]	true	"Some ID"
// @Success 200 {object} web.GenericResponse[[]web.Post]
// @Success 222 {object} web.GenericResponseMulti[[]web.Post, [][]web.Post]
// @Router /posts-multis/ [post]
func GetPostArray(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericResponse[web.Post]{}
}
