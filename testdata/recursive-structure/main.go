package main

type Extension struct {
	Element
	Url string `json:"url"`
}

type Element struct {
	Id        string      `json:"id"`
	Extension []Extension `json:"extension,omitempty"`
}

type Resource struct {
	Element
}

// Resource godoc
// @Summary get
// @Description get resource
//
// @Accept json
// @Param Authorization header string true "access token sent using Bearer prefix"
// @Param ID path string true "ID"
//
// @Success 200 {object} Resource
// @Failure 400
// @Tags resource
// @Router /Resource/{ID} [get]
func _() {}

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
//
// @BasePath /
func main() {}
