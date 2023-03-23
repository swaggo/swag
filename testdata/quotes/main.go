package main

import (
	"github.com/swaggo/swag"
	"github.com/swaggo/swag/testdata/quotes/api"
	_ "github.com/swaggo/swag/testdata/quotes/docs"
)

func ReadDoc() string {
	doc, _ := swag.ReadDoc()
	return doc
}

// @title Swagger Example API
// @version 1.0
// @description.markdown
// @tag.name api
// @tag.description.markdown
// @termsOfService http://swagger.io/terms/

func main() {
	api.RandomFunc()
}
