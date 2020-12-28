package foo

import (
	"net/http"

	"github.com/swaggo/swag/testdata/multiple_def/models/mbar"
)

// @Description get Baz
// @ID get-baz
// @Accept json
// @Produce json
// @Category foo, bar
// @Success 200 {object} mbar.Bar
// @Router /testapi/get-baz [get]
func GetBar(w http.ResponseWriter, r *http.Request) {
	//write your code
	var _ = mbar.Bar{}
}
