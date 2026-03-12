package api

import "net/http"

// Upload godoc
// @Summary      Upload
// @Description  Upload a file.
// @Tags         Some section
// @Accept	     multipart/form-data
// @Produce      text/plain
// @Param        data formData file true "attachment file"
// @Success      201  {string}  "OK"
// @Failure      400  {string}  "Bad Request"
// @Failure      422  {string}  "Validation Error"
// @Failure      500  {string}  "Internal Server Error"
// @Router       /upload [post]
func Upload(w http.ResponseWriter, r *http.Request) {
	//write your code
}
