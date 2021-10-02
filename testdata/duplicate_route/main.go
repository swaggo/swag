package main

import (
	"net/http"

	"github.com/swaggo/swag/testdata/duplicate_route/api"
)

func main() {
	http.HandleFunc("/testapi/endpoint", api.FunctionOne)
	http.HandleFunc("/testapi/endpoint", api.FunctionTwo)
	http.ListenAndServe(":8080", nil)
}
