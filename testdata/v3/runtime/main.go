package main

import (
	"github.com/swaggo/swag/v2"
	_ "github.com/swaggo/swag/v2/testdata/v3/runtime/docs"
)

func ReadDoc() string {
	doc, _ := swag.ReadDoc("OpenAPI3Runtime")
	return doc
}

// @title OpenAPI 3 Runtime Example
// @version 1.0
// @description Runtime OpenAPI 3.1 document fixture.
// @servers.url https://example.com
func main() {}
