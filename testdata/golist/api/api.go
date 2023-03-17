package api

/*
#include "foo.h"
*/
import "C"
import (
	"fmt"
	"net/http"
)

func PrintInt(i, j int) {
	res := C.add(C.int(i), C.int(j))
	fmt.Println(res)
}

type Foo struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	PhotoUrls []string `json:"photoUrls"`
	Status    string   `json:"status"`
}

// GetFoo example
// @Summary Get foo
// @Description get foo
// @ID foo
// @Accept  json
// @Produce  json
// @Param   some_id      query   int     true  "Some ID"
// @Param	some_foo	 formData Foo true "Foo"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /testapi/foo [get]
func GetFoo(w http.ResponseWriter, r *http.Request) {
	// write your code
}
