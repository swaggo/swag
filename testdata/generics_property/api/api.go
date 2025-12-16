package api

import (
	"github.com/swaggo/swag/testdata/generics_property/types"
	"github.com/swaggo/swag/testdata/generics_property/web"
	"net/http"
)

type NestedResponse struct {
	web.GenericResponse[[]string, *uint8]
	Post types.Field[[]types.Post]
}

type Audience[T any] []T

type CreateMovie struct {
	Name           string
	MainActor      types.Field[Person]
	SupportingCast types.Field[[]Person]
	Directors      types.Field[*[]Person]
	CameraPeople   types.Field[[]*Person]
	Producer       types.Field[*Person]
	Audience       Audience[Person]
	AudienceNames  Audience[string]
	Detail1        types.Field[types.Field[Person]]
	Detail2        types.Field[types.Field[string]]
}

type Person struct {
	Name string
}

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Param   data query  web.PostPager true "1"
// @Success 200 {object} web.PostResponse "ok"
// @Success 201 {object} web.PostResponses "ok"
// @Success 202 {object} web.StringResponse "ok"
// @Success 203 {object} NestedResponse "ok"
// @Router /posts [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {
}

// @Summary Create movie
// @Description Create a new movie production
// @Accept  json
// @Produce  json
// @Param   data body  CreateMovie true "Movie Create-Payload"
// @Success 201 {object} CreateMovie "ok"
// @Router /movie [post]
func CreateMovieApi(w http.ResponseWriter, r *http.Request) {
}
