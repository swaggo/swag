package main

import (
	"net/http"
)

//	@title			Swagger Example API
//	@version		1.0
//	@description	This is a sample server with null types overridden with primitive types.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		product_info.swagger.io
//	@BasePath	/v2

func main() {
	http.HandleFunc("/testapi/update-product", UpdateProduct)
	http.ListenAndServe(":8080", nil)
}
