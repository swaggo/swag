package main

import (
	"fmt"
	"github.com/easonlin404/gin-swagger"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.GET("/favicon.ico", handlerTest1)
	router.GET("/", handlerTest1)
	group := router.Group("/users")
	{
		group.GET("/", handlerTest2)
		group.GET("/:id", handlerTest1)
		group.POST("/:id", handlerTest2)
	}

	swagg := swagger.New(router.Routes())

	swagg.Build()

	fmt.Println(swagg.Routes())

	router.Run()

}

func handlerTest1(c *gin.Context) {}
func handlerTest2(c *gin.Context) {}
