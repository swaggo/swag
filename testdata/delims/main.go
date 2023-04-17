package main

import (
	"github.com/swaggo/swag/v2"
	"github.com/swaggo/swag/v2/testdata/delims/api"
	_ "github.com/swaggo/swag/v2/testdata/delims/docs"
)

func ReadDoc() string {
	doc, _ := swag.ReadDoc("CustomDelims")
	return doc
}

// @title Swagger Example API
// @version 1.0
// @description Testing custom template delimeters
// @termsOfService http://swagger.io/terms/

func main() {
	api.MyFunc()
}
