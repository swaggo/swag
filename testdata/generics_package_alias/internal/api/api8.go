package api

import (
	_ "github.com/swaggo/swag/testdata/generics_package_alias/internal/path1/v1"
	_ "github.com/swaggo/swag/testdata/generics_package_alias/internal/path2/v1"
)

// @Summary Create movie
// @Description model from a package whose name conflicts with other packages
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.UniqueProduct ""
// @Router /api15 [post]
func CreateMovie15() {

}
