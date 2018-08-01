package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/testdata/model_not_under_root/data"
)

// @Summary Add a new pet to the store
// @Description get string by ID
// @ID get-string-by-int
// @Accept  json
// @Produce  json
// @Param   some_id      path   int     true  "Some ID" Format(int64)
// @Success 200 {object} data.Foo	"ok"
// @Router /testapi/get-string-by-int/{some_id} [get]
func GetStringByInt(c *gin.Context) {
	var foo data.Foo
	log.Println(foo)
	//write your code
}

// @Summary Upload file
// @Description Upload file
// @ID file.upload
// @Accept  json
// @Produce  json
// @Param data body data.Foo true "Foo to create"
// @Success 200 {string} string "ok"
// @Router /file/upload [post]
func Upload(ctx *gin.Context) {
	//write your code
}
