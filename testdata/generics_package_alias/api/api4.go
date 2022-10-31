package api

import (
	"github.com/swaggo/swag/testdata/external_models/external"
	_ "github.com/swaggo/swag/testdata/generics_package_alias/path1/v1"
)

// @Summary Create movie
// @Description models imported from an external package
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.ListResult[external.MyError] ""
// @Router /api11 [post]
func CreateMovie11() {
	var _ external.MyError
}
