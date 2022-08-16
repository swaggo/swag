package api

import "net/http"

// @Description get Foo
// @ID get-foo
// @Success 200 {object} api.GetFoo.response
// @Router /testapi/get-foo [get]
func GetFoo(w http.ResponseWriter, r *http.Request) {
	type response struct {
	}
}
