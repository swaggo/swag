package main

import (
	"net/http"

	"github.com/swaggo/swag/v2/testdata/v3/composition_inline/api"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server
// @termsOfService http://swagger.io/terms/

// @host petstore.swagger.io
// @BasePath /v2

func main() {
	http.HandleFunc("/testapi/get-server-metadata", api.GetServerMetadata)
	http.HandleFunc("/testapi/get-server-metadata-json-inline", api.GetServerMetadataJSONInline)
	http.HandleFunc("/testapi/get-server-metadata-with-ignored", api.GetServerMetadataWithIgnored)
	http.ListenAndServe(":8080", nil)
}
