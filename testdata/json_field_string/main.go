package main

import (
	"github.com/gin-gonic/gin"
)

type MyStruct struct {
	ID int `json:"id" example:"1" format:"int64"`
	// Post name
	Name string `json:"name" example:"poti"`
	// Post data
	Data struct {
		// Post tag
		Tag []string `json:"name"`
	} `json:"data"`
	// Integer represented by a string
	MyInt int `json:"myint,string"`
}

// @Summary Call DoSomething
// @Description Does something, but internal (non-exported) fields inside a struct won't be marshaled into JSON
// @Accept  json
// @Produce  json
// @Success 200 {string} MyStruct
// @Router /so-something [get]
func DoSomething(c *gin.Context) {
	//write your code
}

// @title Swagger Example API
// @version 1.0
// @description This is a sample server.
// @host localhost:4000
// @basePath /api
func main() {
	r := gin.New()
	r.GET("/do-something", api.DoSomething)
	r.Run()
}
