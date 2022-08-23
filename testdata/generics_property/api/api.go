package api

import (
	"github.com/swaggo/swag/testdata/generics_property/web"
	"net/http"
)

type NestedResponse struct {
	web.GenericResponse[[]string, *uint8]
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
