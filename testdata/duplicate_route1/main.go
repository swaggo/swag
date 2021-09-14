package main

import (
	"net/http"

	"github.com/swaggo/swag/testdata/duplicate_route1/api"
)

func main() {
	http.HandleFunc("/testapi/endpoint", api.Function)
	http.ListenAndServe(":8080", nil)
}
