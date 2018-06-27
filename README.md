# swag

<p align="center">
  <a href="https://swaggo.github.io/swaggo.io/">
    <img alt="swaggo" src="https://raw.githubusercontent.com/swaggo/swaggo.io/master/images/swaggo.png" width="200">
  </a>
</p>

<p align="center">
  Automatically generate RESTful API documentation with Swagger 2.0 for Go.
</p>

<p align="center">
  <a href="https://travis-ci.org/swaggo/swag"><img alt="Travis Status" src="https://img.shields.io/travis/swaggo/swag/master.svg"></a>
  <a href="https://codecov.io/gh/swaggo/swag"><img alt="Coverage Status" src="https://img.shields.io/codecov/c/github/swaggo/swag/master.svg"></a>
  <a href="https://goreportcard.com/badge/github.com/swaggo/swag"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/swaggo/swag"></a>
  <a href="https://codebeat.co/projects/github-com-swaggo-swag-master"><img alt="codebeat badge" src="https://codebeat.co/badges/71e2f5e5-9e6b-405d-baf9-7cc8b5037330" /></a>
  <a href="https://godoc.org/github.com/swaggo/swag"><img alt="Go Doc" src="https://godoc.org/github.com/swaggo/swagg?status.svg"></a>
</p>

<p align="center">gopher image source is <a href="https://github.com/tenntenn/gopher-stickers">tenntenn/gopher-stickers.</a> It has licenses <a href="http://creativecommons.org/licenses/by/3.0/deed.en">creative commons licensing.</a></p>

## Summary
Swag converts Go annotations to Swagger Documentation 2.0. We've created a variety of plugins for popular [Go web frameworks](#supported-web-frameworks). This allows you to quickly integrate with an existing Go project (using Swagger UI).

## Documentation

For a comprehensive explanation of what swag can do, check out the [swag documentation](https://swaggo.github.io/swaggo.io/).

## Getting started

1. Add comments to your API source code, [See Declarative Comments Format](https://swaggo.github.io/swaggo.io/declarative_comments_format/).

2. Download swag by using:
```sh
$ go get -u github.com/swaggo/swag/cmd/swag
```

3. Run `swag init` in the project's root folder which contains the `main.go` file. This will parse your comments and generate the required files (`docs` folder and `docs/docs.go`).
```sh
$ swag init
```

4. In order to serve these files, you can utilize one of our supported plugins. For go's core library, check out [net/http](https://github.com/swaggo/http-swagger).

  * Make sure to import the generated `docs/docs.go` so that your specific configuration gets `init`'ed.
  * If your General API annotation do not live in `main.go`, you can let swag know with `-g`.
  ```sh
  swag init -g http/api.go
  ```

## <a name="supported-web-frameworks"></a> Supported Web Frameworks

- [gin](http://github.com/swaggo/gin-swagger)
- [echo](http://github.com/swaggo/echo-swagger)
- [net/http](https://github.com/swaggo/http-swagger)

## How to use it with `gin`?

Find the example source code [here](https://github.com/swaggo/swag/tree/master/example/celler).

1. After using `swag init` to generate Swagger 2.0 docs, import the following packages:
```go
import "github.com/swaggo/gin-swagger" // gin-swagger middleware
import "github.com/swaggo/gin-swagger/swaggerFiles" // swagger embed files
```

2. Add [General API](https://swaggo.github.io/swaggo.io/declarative_comments_format/api_operation.html) annotations in `main.go` code:

```go
// @title Swagger Example API
// @version 1.0
// @description This is a sample server celler server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

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

func main() {
	r := gin.Default()

	c := controller.NewController()

	v1 := r.Group("/api/v1")
	{
		accounts := v1.Group("/accounts")
		{
			accounts.GET(":id", c.ShowAccount)
			accounts.GET("", c.ListAccounts)
			accounts.POST("", c.AddAccount)
			accounts.DELETE(":id", c.DeleteAccount)
			accounts.PATCH(":id", c.UpdateAccount)
			accounts.POST(":id/images", c.UploadAccountImage)
		}
    //...
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8080")
}
//...

```

3. Add [API Operation](https://swaggo.github.io/swaggo.io/declarative_comments_format/general_api_info.html) annotations in `controller` code

``` go
package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/celler/httputil"
	"github.com/swaggo/swag/example/celler/model"
)

// ShowAccount godoc
// @Summary Show a account
// @Description get string by ID
// @ID get-string-by-int
// @Accept  json
// @Produce  json
// @Param id path int true "Account ID"
// @Success 200 {object} model.Account
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /accounts/{id} [get]
func (c *Controller) ShowAccount(ctx *gin.Context) {
	id := ctx.Param("id")
	aid, err := strconv.Atoi(id)
	if err != nil {
		httputil.NewError(ctx, http.StatusBadRequest, err)
		return
	}
	account, err := model.AccountOne(aid)
	if err != nil {
		httputil.NewError(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, account)
}

// ListAccounts godoc
// @Summary List accounts
// @Description get accounts
// @Accept  json
// @Produce  json
// @Param q query string false "name search by q"
// @Success 200 {array} model.Account
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /accounts [get]
func (c *Controller) ListAccounts(ctx *gin.Context) {
	q := ctx.Request.URL.Query().Get("q")
	accounts, err := model.AccountsAll(q)
	if err != nil {
		httputil.NewError(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, accounts)
}

//...
```

4.Run it, and browse to http://localhost:8080/swagger/index.html. You will see Swagger 2.0 Api documents as shown below:

![swagger_index.html](https://user-images.githubusercontent.com/8943871/31943004-dd08a10e-b88c-11e7-9e77-19d2c759a586.png)

## Examples

[swaggo + gin](https://github.com/swaggo/swag/tree/master/example)

## Implementation Status

[Swagger 2.0 document](https://swagger.io/docs/specification/2-0/basic-structure/)

- [x] Basic Structure
- [x] API Host and Base Path
- [x] Paths and Operations
- [x] Describing Parameters
- [x] Describing Request Body
- [x] Describing Responses
- [x] MIME Types
- [x] Authentication
  - [x] Basic Authentication
  - [x] API Keys
- [x] Adding Examples
- [x] File Upload
- [x] Enums
- [x] Grouping Operations With Tags
- [ ] Swagger Extensions

## About the Project
This project was inspired by [swagger](https://github.com/yvasiyarov/swagger) but we simplified the usage and added support a variety of [web frameworks](#supported-web-frameworks).
