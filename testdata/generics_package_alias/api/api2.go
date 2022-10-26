package api

import (
	myv1 "github.com/swaggo/swag/testdata/generics_package_alias/path1/v1"
	myv2 "github.com/swaggo/swag/testdata/generics_package_alias/path2/v1"
)

// @Summary Create movie
// @Description Create a new movie production
// @Accept  json
// @Produce  json
// @Success 200 {object} myv2.ListResult[myv2.ProductDto] ""
// @Router /api3 [post]
func CreateMovie3() {
	_ = myv2.ListResult[myv2.ProductDto]{}
}

// @Summary Create movie
// @Description Create a new movie production
// @Accept  json
// @Produce  json
// @Success 200 {object} myv2.RenamedListResult[myv2.RenamedProductDto] ""
// @Router /api4 [post]
func CreateMovie4() {
	_ = myv2.ListResult[myv2.ProductDto]{}
}

// @Summary Create movie
// @Description Create a new movie production
// @Accept  json
// @Produce  json
// @Success 200 {object} myv1.ListResult[myv2.ProductDto] ""
// @Router /api5 [post]
func CreateMovie5() {
	_ = myv1.ListResult[myv2.ProductDto]{}
}

// @Summary Create movie
// @Description Create a new movie production
// @Accept  json
// @Produce  json
// @Success 200 {object} myv1.RenamedListResult[myv2.RenamedProductDto] ""
// @Router /api6 [post]
func CreateMovie6() {
	_ = myv1.ListResult[myv2.ProductDto]{}
}
