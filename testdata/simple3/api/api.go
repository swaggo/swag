package api

import "net/http"

// @Summary Add a new pet to the store
// @Description get string by ID
// @ID get-string-by-int
// @Accept  json
// @Produce  json
// @Param   some_id      path   int     true  "Some ID" Format(int64)
// @Param   some_id      body web.Pet true  "Some ID"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /testapi/get-string-by-int/{some_id} [get]
func GetStringByInt(w http.ResponseWriter, r *http.Request) {
	//write your code
}

// @Description get struct array by ID
// @ID get-struct-array-by-string
// @Accept  json
// @Produce  json
// @Param some_id path string true "Some ID"
// @Param category query int true "Category" Enums(1, 2, 3)
// @Param offset query int true "Offset" Minimum(0) default(0)
// @Param limit query int true "Limit" Maximum(50) default(10)
// @Param q query string true "q" Minlength(1) Maxlength(50) default("")
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Security ApiKeyAuth
// @Security BasicAuth
// @Security OAuth2Application[write]
// @Security OAuth2Implicit[read, admin]
// @Security OAuth2AccessCode[read]
// @Security OAuth2Password[admin]
// @Router /testapi/get-struct-array-by-string/{some_id} [get]
func GetStructArrayByString(w http.ResponseWriter, r *http.Request) {
	//write your code
}

// @Summary Upload file
// @Description Upload file
// @ID file.upload
// @Accept  multipart/form-data
// @Produce  json
// @Param   file formData file true  "this is a test file"
// @Success 200 {string} string "ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /file/upload [post]
func Upload(w http.ResponseWriter, r *http.Request) {
	//write your code
}

// @Summary use Anonymous field
// @Success 200 {object} web.RevValue "ok"
func AnonymousField() {

}

// @Summary use pet2
// @Success 200 {object} web.Pet2 "ok"
func Pet2() {

}

type Pet3 struct {
	ID int `json:"id"`
}
