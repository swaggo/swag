package api

import "net/http"

// @Description add Foo
// @Deprecated
// @Success 200 {string} string
// @Router /testapi/foo1 [put]
// @Router /testapi/foo1 [post]
// @Router /test/api/foo1 [post]
func AddFoo(w http.ResponseWriter, r *http.Request) {}

// @Description get Foo
// @Success 200 {string} string
// @Router /testapi/foo1 [get]
// @DeprecatedRouter /test/api/foo1 [get]
func GetFoo(w http.ResponseWriter, r *http.Request) {}
