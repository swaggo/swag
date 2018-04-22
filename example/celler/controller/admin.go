package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
	"github.com/swaggo/swag/example/celler/model"
)

// Auth godoc
// @Summary Auth admin
// @Description get admin info
// @Tags accounts,admin
// @Accept  json
// @Produce  json
// @Success 200 {object} model.Admin
// @Failure 400 {object} httputil.HTTPError
// @Failure 401 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Security ApiKeyAuth
// @Router /admin/auth [post]
func (c *Controller) Auth(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if len(authHeader) == 0 {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("please set Header Authorization"))
		return
	}
	if authHeader != "admin" {
		httputil.NewError(ctx, http.StatusUnauthorized, fmt.Errorf("this user isn't authorized to operation key=%s expected=admin", authHeader))
		return
	}
	admin := model.Admin{
		ID:   1,
		Name: "admin",
	}
	ctx.JSON(http.StatusOK, admin)
}
