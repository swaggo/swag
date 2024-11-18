package api

import (
	"net/http"

	. "github.com/rampnow-io/swag/testdata/error/errors"
	_ "github.com/rampnow-io/swag/testdata/error/web"
)

// Upload do something
// @Summary Upload file
// @Description Upload file
// @ID file.upload
// @Accept  multipart/form-data
// @Produce  json
// @Param   file formData file true  "this is a test file"
// @Success 200 {string} string "ok"
// @Failure 400 {object} web.CrossErrors "Abort !!"
// @Router /file/upload [post]
func Upload(w http.ResponseWriter, r *http.Request) {
	//write your code
	_ = Errors{}
}
