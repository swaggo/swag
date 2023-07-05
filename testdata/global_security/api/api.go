package api

import (
	"net/http"
)

// @Summary Get application
// @Description test get application
// @Success 200
// @Router /testapi/application [get]
func GetApplication(w http.ResponseWriter, r *http.Request) {
	//write your code
}

// Summary Get no security
// @Description override global security
// @Security
// @Success 200
// @Router /testapi/nosec [get]
func GetNoSec(w http.ResponseWriter, r *http.Request) {
	//write your code
}
