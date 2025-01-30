package main

import (
	"github.com/yalochat/swag/testdata/simple_async/api"
)

// @title Swagger Example AsyncAPI
// @version 1.0
// @description This is a sample server Petstore server.
func main() {
	api.ConfigEventDrivenChannel()
}
