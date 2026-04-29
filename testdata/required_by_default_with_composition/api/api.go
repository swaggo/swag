package api

import "net/http"

// Base has Name and Foo fields.
type Base struct {
	Name string `json:"name"`
	Foo  string `json:"foo"`
}

// Parent embeds Base and shadows Name with its own field.
type Parent struct {
	Base
	Name string `json:"name"`
	Bar  string `json:"bar"`
}

// GetParent godoc
// @Description get Parent
// @ID get-parent
// @Accept json
// @Produce json
// @Success 200 {object} api.Parent
// @Router /parent [get]
func GetParent(w http.ResponseWriter, r *http.Request) {}
