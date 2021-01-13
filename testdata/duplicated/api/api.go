package api

import "net/http"

// @Description get Foo
// @ID get-foo
// @Success 200 {string} string
// @Router /testapi/get-foo [get]
func GetFoo(w http.ResponseWriter, r *http.Request) {}

// @Description post Bar
// @ID get-foo
// @Success 200 {string} string
// @Router /testapi/post-bar [post]
func PostBar(w http.ResponseWriter, r *http.Request) {}
