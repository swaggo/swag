package api

import (
	"net/http"
)

// @Summary List Posts
// @Description Get All of the Posts
// @Accept  json
// @Produce  json
// @Param   data query  web.PostPager true "1"
// @Success 200 {object} web.PostResponse "ok"
// @Success 201 {object} web.PostResponses "ok"
// @Success 202 {object} web.StringResponse "ok"
// @Router /posts [get]
func GetPosts(w http.ResponseWriter, r *http.Request) {
}
