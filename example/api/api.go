package api

import (
	"fmt"
	"gin"
)



//
// @Title GetStringByInt
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Param   some_id     path    int     true        "Some ID"
// @Success 200 {object} string
// @Failure 400 {object} APIError "We need ID!!"
// @Failure 404 {object} APIError "Can not find ID"
// @Router /testapi/get-string-by-int/{some_id} [get]
func GetStringByInt(c *gin.Context) {

}

type Pet3 struct {
	ID int `json:"id"`
}
