package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	SearchDir = "./testdata/format_test"
	Excludes  = "./testdata/format_test/web"
	MainFile  = "main.go"
)

func testFormat(t *testing.T, filename, contents, want string) {
	got, err := NewFormatter().Format(filename, []byte(contents))
	assert.NoError(t, err)
	assert.Equal(t, want, string(got))
}

func Test_FormatMain(t *testing.T) {
	contents := `package main
	// @title Swagger Example API
	// @version 1.0
	// @description This is a sample server Petstore server.
	// @termsOfService http://swagger.io/terms/

	// @contact.name API Support
	// @contact.url http://www.swagger.io/support
	// @contact.email support@swagger.io

	// @license.name Apache 2.0
	// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

	// @host petstore.swagger.io
	// @BasePath /v2

	// @securityDefinitions.basic BasicAuth

	// @securityDefinitions.apikey ApiKeyAuth
	// @in header
	// @name Authorization

	// @securitydefinitions.oauth2.application OAuth2Application
	// @tokenUrl https://example.com/oauth/token
	// @scope.write Grants write access
	// @scope.admin Grants read and write access to administrative information

	// @securitydefinitions.oauth2.implicit OAuth2Implicit
	// @authorizationurl https://example.com/oauth/authorize
	// @scope.write Grants write access
	// @scope.admin Grants read and write access to administrative information

	// @securitydefinitions.oauth2.password OAuth2Password
	// @tokenUrl https://example.com/oauth/token
	// @scope.read Grants read access
	// @scope.write Grants write access
	// @scope.admin Grants read and write access to administrative information

	// @securitydefinitions.oauth2.accessCode OAuth2AccessCode
	// @tokenUrl https://example.com/oauth/token
	// @authorizationurl https://example.com/oauth/authorize
	// @scope.admin Grants read and write access to administrative information
	func main() {}`

	want := `package main
	//	@title			Swagger Example API
	//	@version		1.0
	//	@description	This is a sample server Petstore server.
	//	@termsOfService	http://swagger.io/terms/

	//	@contact.name	API Support
	//	@contact.url	http://www.swagger.io/support
	//	@contact.email	support@swagger.io

	//	@license.name	Apache 2.0
	//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

	//	@host		petstore.swagger.io
	//	@BasePath	/v2

	//	@securityDefinitions.basic	BasicAuth

	//	@securityDefinitions.apikey	ApiKeyAuth
	//	@in							header
	//	@name						Authorization

	//	@securitydefinitions.oauth2.application	OAuth2Application
	//	@tokenUrl								https://example.com/oauth/token
	//	@scope.write							Grants write access
	//	@scope.admin							Grants read and write access to administrative information

	//	@securitydefinitions.oauth2.implicit	OAuth2Implicit
	//	@authorizationurl						https://example.com/oauth/authorize
	//	@scope.write							Grants write access
	//	@scope.admin							Grants read and write access to administrative information

	//	@securitydefinitions.oauth2.password	OAuth2Password
	//	@tokenUrl								https://example.com/oauth/token
	//	@scope.read								Grants read access
	//	@scope.write							Grants write access
	//	@scope.admin							Grants read and write access to administrative information

	//	@securitydefinitions.oauth2.accessCode	OAuth2AccessCode
	//	@tokenUrl								https://example.com/oauth/token
	//	@authorizationurl						https://example.com/oauth/authorize
	//	@scope.admin							Grants read and write access to administrative information
	func main() {}`
	testFormat(t, "main.go", contents, want)
}

func Test_FormatMultipleFunctions(t *testing.T) {
	contents := `package main

	// @Produce json
	// @Success 200 {object} string
	// @Failure 400 {object} string
	func A() {}

	// @Description Description of B.
	// @Produce json
	// @Success 200 {array} string
	// @Failure 400 {object} string
	func B() {}`

	want := `package main

	//	@Produce	json
	//	@Success	200	{object}	string
	//	@Failure	400	{object}	string
	func A() {}

	//	@Description	Description of B.
	//	@Produce		json
	//	@Success		200	{array}		string
	//	@Failure		400	{object}	string
	func B() {}`

	testFormat(t, "main.go", contents, want)
}

func Test_FormatApi(t *testing.T) {
	contents := `package api

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
	func GetStringByInt(w http.ResponseWriter, r *http.Request) {}`

	want := `package api

	import "net/http"

	//	@Summary		Add a new pet to the store
	//	@Description	get string by ID
	//	@ID				get-string-by-int
	//	@Accept			json
	//	@Produce		json
	//	@Param			some_id	path		int				true	"Some ID"	Format(int64)
	//	@Param			some_id	body		web.Pet			true	"Some ID"
	//	@Success		200		{string}	string			"ok"
	//	@Failure		400		{object}	web.APIError	"We need ID!!"
	//	@Failure		404		{object}	web.APIError	"Can not find ID"
	//	@Router			/testapi/get-string-by-int/{some_id} [get]
	func GetStringByInt(w http.ResponseWriter, r *http.Request) {}`

	testFormat(t, "api.go", contents, want)
}

func Test_NonSwagComment(t *testing.T) {
	contents := `package api
	// @Summary Add a new pet to the store
	// @Description get string by ID
	// @ID get-string-by-int
	// @ Accept json
	// This is not a @swag comment`
	want := `package api
	//	@Summary		Add a new pet to the store
	//	@Description	get string by ID
	//	@ID				get-string-by-int
	// @ Accept json
	// This is not a @swag comment`

	testFormat(t, "non_swag.go", contents, want)
}

func Test_EmptyComment(t *testing.T) {
	contents := `package empty
	// @Summary Add a new pet to the store
	// @Description  `
	want := `package empty
	//	@Summary	Add a new pet to the store
	//	@Description`

	testFormat(t, "empty.go", contents, want)
}

func Test_AlignAttribute(t *testing.T) {
	contents := `package align
	// @Summary Add a new pet to the store
	//  @Description Description`
	want := `package align
	//	@Summary		Add a new pet to the store
	//	@Description	Description`

	testFormat(t, "align.go", contents, want)

}

func Test_SyntaxError(t *testing.T) {
	contents := []byte(`package invalid
	func invalid() {`)

	_, err := NewFormatter().Format("invalid.go", contents)
	assert.Error(t, err)
}
