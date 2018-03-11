package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/model"
)

// Auth godoc
// @Summary Auth admin
// @Description get admin info
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} model.Admin
// @Failure 400 {object} controller.HTTPError
// @Failure 401 {object} controller.HTTPError
// @Failure 404 {object} controller.HTTPError
// @Failure 500 {object} controller.HTTPError
// @Router /admin/auth [post]
func (c *Controller) Auth(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if len(authHeader) == 0 {
		NewError(ctx, http.StatusBadRequest, errors.New("please set Header Authorization"))
		return
	}
	if authHeader != "admin" {
		NewError(ctx, http.StatusUnauthorized, fmt.Errorf("this user isn't authorized to operation key=%s", authHeader))
		return
	}
	admin := model.Admin{
		ID:   1,
		Name: "admin",
	}
	ctx.JSON(http.StatusOK, admin)
}
