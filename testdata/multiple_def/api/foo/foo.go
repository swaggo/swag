package foo

import (
	"net/http"

	"github.com/swaggo/swag/testdata/multiple_def/models/mfoo"
)

// @Description get Foo
// @ID get-foo
// @Accept json
// @Produce json
// @Category foo
// @Success 200 {object} mfoo.Foo
// @Router /testapi/get-foo [get]
func GetBar(w http.ResponseWriter, r *http.Request) {
	//write your code
	var _ = mfoo.Foo{}
}
