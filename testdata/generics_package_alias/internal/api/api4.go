package api

import (
	"github.com/rampnow-io/swag/testdata/generics_package_alias/external/external1"
	_ "github.com/rampnow-io/swag/testdata/generics_package_alias/internal/path1/v1"
)

// @Summary Create movie
// @Description models imported from an external package
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.ListResult[external1.Customer] ""
// @Router /api11 [post]
func CreateMovie11() {
	var _ external1.Customer
}
