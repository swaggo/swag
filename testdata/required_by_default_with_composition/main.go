package main

import (
	"net/http"

	"github.com/swaggo/swag/testdata/required_by_default_with_composition/api"
)

// @title Swagger Example API
// @version 1.0
// @host localhost
// @BasePath /

func main() {
	http.HandleFunc("/parent", api.GetParent)
	http.ListenAndServe(":8080", nil)
}
