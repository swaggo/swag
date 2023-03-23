package api

import (
	"net/http"

	"github.com/swaggo/swag/testdata/generics_basic/types"
	"github.com/swaggo/swag/testdata/generics_basic/web"
)

type Response[T any, X any] struct {
	Data T
	Meta X

	Status string
}

type StringStruct struct {
	Data string
}

// @Summary Add a new pet to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data        body   web.GenericBody[types.Post]    true  "Some ID"
// @Success 200 {object} web.GenericResponse[types.Post]
// @Success 201 {object} web.GenericResponse[types.Hello]
// @Success 202 {object} web.GenericResponse[types.Field[string]]
// @Success 203 {object} web.GenericResponse[types.Field[int]]
// @Success 204 {object} Response[string, types.Field[int]]
// @Success 205 {object} Response[StringStruct, types.Field[int]]
// @Success 222 {object} web.GenericResponseMulti[types.Post, types.Post]
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /posts/ [post]
func GetPost(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericResponse[types.Post]{}
}

// @Summary Add new pets to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data	body	web.GenericBodyMulti[types.Post, types.Post]	true	"Some ID"
// @Success 200 {object} web.GenericResponse[types.Post]
// @Success 201 {object} web.GenericResponse[types.Hello]
// @Success 202 {object} web.GenericResponse[types.Field[string]]
// @Success 222 {object} web.GenericResponseMulti[types.Post, types.Post]
// @Router /posts-multi/ [post]
func GetPostMulti(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericResponse[types.Post]{}
}

// @Summary Add new pets to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   data	body	web.GenericBodyMulti[[]types.Post, [][]types.Post]	true	"Some ID"
// @Success 200 {object} web.GenericResponse[[]types.Post]
// @Success 201 {object} web.GenericResponse[[]types.Hello]
// @Success 222 {object} web.GenericResponseMulti[[]types.Post, [][]types.Post]
// @Router /posts-multis/ [post]
func GetPostArray(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = web.GenericResponse[types.Post]{}
}
