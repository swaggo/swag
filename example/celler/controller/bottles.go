package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
	"github.com/swaggo/swag/example/celler/model"
)

// ShowBottle godoc
// @Summary Show a bottle
// @Description get string by ID
// @ID get-string-by-int
// @Tags bottles
// @Accept  json
// @Produce  json
// @Param  id path int true "Bottle ID"
// @Success 200 {object} model.Bottle
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /bottles/{id} [get]
func (c *Controller) ShowBottle(ctx *gin.Context) {
	id := ctx.Param("id")
	bid, err := strconv.Atoi(id)
	if err != nil {
		httputil.NewError(ctx, http.StatusBadRequest, err)
		return
	}
	bottle, err := model.BottleOne(bid)
	if err != nil {
		httputil.NewError(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, bottle)
}

// ListBottles godoc
// @Summary List bottles
// @Description get bottles
// @Tags bottles
// @Accept  json
// @Produce  json
// @Success 200 {array} model.Bottle
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /bottles [get]
func (c *Controller) ListBottles(ctx *gin.Context) {
	bottles, err := model.BottlesAll()
	if err != nil {
		httputil.NewError(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, bottles)
}
