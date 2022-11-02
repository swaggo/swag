package api

import (
	_ "github.com/swaggo/swag/testdata/generics_package_alias/internal/path1/v1"
	. "github.com/swaggo/swag/testdata/generics_package_alias/internal/path2/v1"
)

// @Summary Create movie
// @Description models imported from an unnamed package
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.ListResult[v1.ProductDto] ""
// @Router /api07 [post]
func CreateMovie07() {
	var _ ProductDto
}

// @Summary Create movie
// @Description models imported from an unnamed package
// @Accept  json
// @Produce  json
// @Success 200 {object} ListResult[ProductDto] ""
// @Router /api08 [post]
func CreateMovie08() {
	var _ ProductDto
}

// @Summary Create movie
// @Description models imported from an unnamed package
// @Accept  json
// @Produce  json
// @Success 200 {object} ListResult[v1.ProductDto] ""
// @Router /api09 [post]
func CreateMovie09() {
	var _ ProductDto
}

// @Summary Create movie
// @Description models imported from an unnamed package
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.ListResult[ProductDto] ""
// @Router /api10 [post]
func CreateMovie10() {
	var _ ProductDto
}
