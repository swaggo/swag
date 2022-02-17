package main

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag/example/markdown/api"
	_ "github.com/swaggo/swag/example/markdown/docs"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @description.markdown
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @tag.name admin
// @tag.description.markdown

// @BasePath /v2

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/admin/user/", api.ListUsers).Methods("GET")
	router.HandleFunc("/admin/user/{id}", api.GetUser).Methods("GET")
	router.HandleFunc("/admin/user/", api.AddUser).Methods("POST")
	router.HandleFunc("/admin/user/{id}", api.UpdateUser).Methods("PUT")

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	http.ListenAndServe(":8080", router)
}
