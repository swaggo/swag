package api

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/testdata/error_field/data"
)

// @Summary Get application
// @Description test get application
// @ID get-application
// @Accept  json
// @Produce  json
// @Success 200 {object} data.ErrorResponse	"ok"
// @Router /testapi/application [get]
func GetApplication(c *gin.Context) {
	var foo = data.ErrorResponse{
		Error: errors.New("just a random thing"),
	}
	log.Println(foo)
}
