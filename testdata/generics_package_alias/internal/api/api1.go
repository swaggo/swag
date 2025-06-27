package api

import (
	myv1 "github.com/swaggo/swag/testdata/generics_package_alias/internal/path1/v1"
)

// @Summary Create movie
// @Description Create a new movie production
// @Accept  json
// @Produce  json
// @Success 200 {object} myv1.ListResult[myv1.ProductDto] ""
// @Router /api01 [post]
func CreateMovie01() {
	_ = myv1.ListResult[myv1.ProductDto]{}
}

// @Summary Create movie
// @Description Create a new movie production
// @Accept  json
// @Produce  json
// @Success 200 {object} myv1.RenamedListResult[myv1.RenamedProductDto] ""
// @Router /api02 [post]
func CreateMovie02() {
	_ = myv1.ListResult[myv1.ProductDto]{}
}
