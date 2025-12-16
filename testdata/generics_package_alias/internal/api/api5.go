package api

import (
	myexternal "github.com/swaggo/swag/testdata/generics_package_alias/external/external2"
	_ "github.com/swaggo/swag/testdata/generics_package_alias/internal/path1/v1"
)

// @Summary Create movie
// @Description models imported from a named external package
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.ListResult[myexternal.Customer] ""
// @Router /api12 [post]
func CreateMovie12() {
	var _ myexternal.Customer
}
