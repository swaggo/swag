package api

import (
	. "github.com/swaggo/swag/testdata/generics_package_alias/external/external3"
	_ "github.com/swaggo/swag/testdata/generics_package_alias/internal/path1/v1"
)

// @Summary Create movie
// @Description models from an external package imported by mode dot
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.ListResult[Customer] ""
// @Router /api13 [post]
func CreateMovie13() {
	var _ Customer
}
