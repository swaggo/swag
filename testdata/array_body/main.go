package main

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/testdata/array_body/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server.
// @host localhost:4000
// @basePath /api
func main() {
	r := gin.New()
	r.GET("/posts/:post_id", api.GetPost)
	r.POST("/posts", api.PostPosts)
	r.Run()
}
