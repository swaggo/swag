package api

import (
	_ "github.com/swaggo/swag/testdata/generics_package_alias/external/external4"
	_ "github.com/swaggo/swag/testdata/generics_package_alias/internal/path1/v1"
)

// @Summary Create movie
// @Description models imported from an unnamed external package
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.ListResult[external4.Customer] ""
// @Router /api14 [post]
func CreateMovie14() {

}
