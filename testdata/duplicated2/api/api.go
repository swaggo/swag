package api

import "net/http"

// @Description put Foo
// @ID put-foo
// @Success 200 {string} string
// @Router /testapi/put-foo [put]
func PutFoo(w http.ResponseWriter, r *http.Request) {}

// @Description head Foo
// @ID head-foo
// @Success 200 {string} string
// @Router /testapi/head-foo [head]
func HeadFoo(w http.ResponseWriter, r *http.Request) {}

// @Description options Foo
// @ID options-foo
// @Success 200 {string} string
// @Router /testapi/options-foo [options]
func OptionsFoo(w http.ResponseWriter, r *http.Request) {}

// @Description patch Foo
// @ID patch-foo
// @Success 200 {string} string
// @Router /testapi/patch-foo [patch]
func PatchFoo(w http.ResponseWriter, r *http.Request) {}

// @Description delete Foo
// @ID put-foo
// @Success 200 {string} string
// @Router /testapi/delete-foo [delete]
func DeleteFoo(w http.ResponseWriter, r *http.Request) {}
