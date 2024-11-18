package api

import (
	"net/http"

	"github.com/rampnow-io/swag/testdata/generics_function_scoped/types"
)

// @Summary Generic Response
// @Produce  json
// @Success 200 {object} types.GenericResponse[api.GetGeneric.User]
// @Success 201 {object} types.GenericResponse[api.GetGeneric.Post]
// @Router / [get]
func GetGeneric(w http.ResponseWriter, r *http.Request) {
	type User struct {
		Username int    `json:"username"`
		Email    string `json:"email"`
	}
	type Post struct {
		Slug  int    `json:"slug"`
		Title string `json:"title"`
	}

	_ = types.GenericResponse[any]{}
}

// @Summary Generic Response With Custom Type Names
// @Produce  json
// @Success 200 {object} types.GenericResponse[api.GetGenericRenamed.User]
// @Success 201 {object} types.GenericResponse[api.GetGenericRenamed.Post]
// @Router /renamed [get]
func GetGenericRenamed(w http.ResponseWriter, r *http.Request) {
	type User struct {
		Username int    `json:"username"`
		Email    string `json:"email"`
	} // @Name RenamedUserData
	type Post struct {
		Slug  int    `json:"slug"`
		Title string `json:"title"`
	} // @Name RenamedPostData

	_ = types.GenericResponse[any]{}
}

// @Summary Multiple Generic Response
// @Produce  json
// @Success 200 {object} types.GenericMultiResponse[api.GetGenericMulti.MyStructA, api.GetGenericMulti.MyStructB]
// @Success 201 {object} types.GenericMultiResponse[api.GetGenericMulti.MyStructB, api.GetGenericMulti.MyStructA]
// @Router /multi [get]
func GetGenericMulti(w http.ResponseWriter, r *http.Request) {
	type MyStructA struct {
		SomeFieldA string `json:"some_field_a"`
	}
	type MyStructB struct {
		SomeFieldB string `json:"some_field_b"`
	}

	_ = types.GenericMultiResponse[any, any]{}
}

// @Summary Multiple Generic Response With Custom Type Names
// @Produce  json
// @Success 200 {object} types.GenericMultiResponse[api.GetGenericMultiRenamed.MyStructA, api.GetGenericMultiRenamed.MyStructB]
// @Success 201 {object} types.GenericMultiResponse[api.GetGenericMultiRenamed.MyStructB, api.GetGenericMultiRenamed.MyStructA]
// @Router /multi-renamed [get]
func GetGenericMultiRenamed(w http.ResponseWriter, r *http.Request) {
	type MyStructA struct {
		SomeFieldA string `json:"some_field_a"`
	} // @Name NameForMyStructA
	type MyStructB struct {
		SomeFieldB string `json:"some_field_b"`
	} // @Name NameForMyStructB

	_ = types.GenericMultiResponse[any, any]{}
}
