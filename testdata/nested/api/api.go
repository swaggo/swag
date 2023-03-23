package api

import (
	"net/http"

	"github.com/swaggo/swag/testdata/nested2"
)

type Foo struct {
	Field1      string `validate:"required"`
	OutsideData *nested2.Body
	InsideData  Bar      `validate:"required"`
	ArrayField1 []string `validate:"required"`
	ArrayField2 []Bar    `validate:"required"`
}

type Bar struct {
	Field string
}

// @Description get Foo
// @ID get-foo
// @Accept json
// @Produce json
// @Success 200 {object} api.Foo
// @Router /testapi/get-foo [get]
func GetFoo(w http.ResponseWriter, r *http.Request) {
	//write your code
	var _ = Foo{}
}
