package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MyStruct struct {
	ID       int     `json:"id" example:"1" format:"int64"`
	Name     string  `json:"name" example:"poti"`
	Intvar   int     `json:"myint,string"`                   // integer as string
	Boolvar  bool    `json:",string"`                        // boolean as a string
	TrueBool bool    `json:"truebool,string" example:"true"` // boolean as a string
	Floatvar float64 `json:",string"`                        // float as a string
}

// @Summary Call DoSomething
// @Description Does something
// @Accept  json
// @Produce  json
// @Param body body MyStruct true "My Struct"
// @Success 200 {object} MyStruct
// @Failure 500
// @Router /do-something [post]
func DoSomething(c *gin.Context) {
	objectFromJSON := new(MyStruct)
	if err := c.BindJSON(&objectFromJSON); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.JSON(http.StatusOK, objectFromJSON)
}

// @title Swagger Example API
// @version 1.0
// @description This is a sample server.
// @host localhost:4000
// @basePath /
func main() {
	r := gin.New()
	r.POST("/do-something", DoSomething)
	r.Run()
}
