package api

import (
	_ "github.com/swaggo/swag/testdata/external_models/external"
	_ "github.com/swaggo/swag/testdata/generics_package_alias/path1/v1"
)

// @Summary Create movie
// @Description models imported from an unnamed external package
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.ListResult[external.MyError] ""
// @Router /api14 [post]
func CreateMovie14() {

}
