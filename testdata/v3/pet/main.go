package main

import (
	"net/http"

	"github.com/nguyennm96/swag/v2/testdata/v3/pet/web"
)

// @title Swagger Petstore
// @version 1.0
// @description This is a sample server Petstore server.  You can find out more about     Swagger at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).      For this sample, you can use the api key 'special-key' to test the authorization     filters.
// @termsOfService http://swagger.io/terms/

// @contact.email apiteam@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	http.HandleFunc("/testapi/pets", web.GetPets)
}
