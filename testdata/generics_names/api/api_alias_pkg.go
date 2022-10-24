package api

import (
	"net/http"

	mytypes "github.com/swaggo/swag/testdata/generics_names/types"
	myweb "github.com/swaggo/swag/testdata/generics_names/web"
)

// @Summary Add a new pet to the store
// @Description get string by ID
// @Accept  json
// @Produce  json
// @Success 200 {object} myweb.AliasPkgGenericResponse[mytypes.Post]
// @Router /posts/aliaspkg [post]
func GetPostFromAliasPkg(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = myweb.AliasPkgGenericResponse[mytypes.Post]{}
}
