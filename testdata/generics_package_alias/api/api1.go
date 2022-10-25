package api

import (
	myv1 "github.com/swaggo/swag/testdata/generics_package_alias/path1/v1"
)

// @Summary Create movie
// @Description Create a new movie production
// @Accept  json
// @Produce  json
// @Success 200 {object} myv1.ListResult[myv1.ProductDto] ""
// @Router /api1 [post]
func CreateMovie1() {
	_ = myv1.ListResult[myv1.ProductDto]{}
}

// @Summary Create movie
// @Description Create a new movie production
// @Accept  json
// @Produce  json
// @Success 200 {object} myv1.RenamedListResult[myv1.RenamedProductDto] ""
// @Router /api2 [post]
func CreateMovie2() {
	_ = myv1.ListResult[myv1.ProductDto]{}
}
