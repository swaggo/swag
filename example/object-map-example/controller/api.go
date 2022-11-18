package controller

import "github.com/gin-gonic/gin"

// GetMap godoc
//
//	@Summary		Get Map Example
//	@Description	get map
//	@ID				get-map
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Response
//	@Router			/test [get]
func (c *Controller) GetMap(ctx *gin.Context) {
	ctx.JSON(200, Response{
		Title: map[string]string{
			"en": "Map",
		},
		CustomType: map[string]interface{}{
			"key": "value",
		},
		Object: Data{
			Text: "object text",
		},
	})
}
