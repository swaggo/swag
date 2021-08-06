package api

import (
	"net/http"

	_ "github.com/swaggo/swag/testdata/simple/web"
)

// @Router /testapi/endpoint [get]
func FunctionOne(w http.ResponseWriter, r *http.Request) {
	//write your code
}

// @Router /testapi/endpoint [get]
func FunctionTwo(w http.ResponseWriter, r *http.Request) {
	//write your code
}
