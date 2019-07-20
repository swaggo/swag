package api

import (
	"github.com/gin-gonic/gin"
)

// PostPosts godoc
// @Summary Add a lot of posts to somewhere
// @Description Post lot of posts to somewhere
// @Accept  json
// @Produce  json
// @Param posts body []web.Post true "This is a lots of posts"
// @Success 200 {string} web.Post
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /posts [post]
func PostPosts(c *gin.Context) {
	//write your code
}

// GetPost godoc
// @Summary Add a new pet to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   post_id      path   int     true  "Some ID" Format(int64)
// @Success 200 {string} web.Post
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /posts/{post_id} [get]
func GetPost(c *gin.Context) {
	//write your code
}
