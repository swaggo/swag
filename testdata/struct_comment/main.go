package main

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/testdata/struct_comment/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
func main() {
	r := gin.New()
	r.GET("/posts/:post_id", api.GetPost)
	r.Run()
}
