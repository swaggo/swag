package api

import (
	"net/http"

	shared "github.com/swaggo/swag/testdata/dependency_local_type/pkg/a"
	_ "github.com/swaggo/swag/testdata/dependency_local_type/pkg/b"
)

// Get godoc
// @Summary get dependency local type
// @Tags dependency-local-type
// @Success 200 {object} shared.Outer
// @Router /test [get]
func Get(w http.ResponseWriter, r *http.Request) {
}
