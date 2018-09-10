# swag

<p align="center">
    <img alt="swaggo" src="https://raw.githubusercontent.com/swaggo/swag/master/assets/swaggo.png" width="200">
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

## Content
 - [Getting started](#getting-started)
 - [Go web frameworks](#supported-web-frameworks)
 - [Supported Web Frameworks](#supported-web-frameworks)
 - [How to use it with `gin`?](#how-to-use-it-with-`gin`?)
 - [Implementation Status](#implementation-status)
 - [swag cli](#swag-cli)
 - [General API Info](#general-api-info)
 - [Security](#security)
 - [API Operation](#api-operation)
 - [TIPS](#tips)
 	- [User defined structure with an array type](#user-defined-structure-with-an-array-type)
	- [Use multiple path params](#use-multiple-path-params)
	- [Example value of struct](#example-value-of-struct)
	- [Description of struct](#description-of-struct)
- [About the Project](#about-the-project)

## Summary

Swag converts Go annotations to Swagger Documentation 2.0. We've created a variety of plugins for popular [Go web frameworks](#supported-web-frameworks). This allows you to quickly integrate with an existing Go project (using Swagger UI).

## Examples

[swaggo + gin](https://github.com/swaggo/swag/tree/master/example)


## Getting started

1. Add comments to your API source code, [See Declarative Comments Format](#general-api-info).

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

## Supported Web Frameworks

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

2. Add [General API](#general-api-info) annotations in `main.go` code:

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

3. Add [API Operation](#api-operation) annotations in `controller` code

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

```console
$ swag init
```

4.Run your app, and browse to http://localhost:8080/swagger/index.html. You will see Swagger 2.0 Api documents as shown below:

![swagger_index.html](https://raw.githubusercontent.com/swaggo/swag/master/assets/swagger-image.png)

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

# swag cli

```console
$ swag init -h
NAME:
   swag init - Create docs.go

USAGE:
   swag init [command options] [arguments...]

OPTIONS:
   --generalInfo value, -g value       Go file path in which 'swagger general API Info' is written (default: "main.go")
   --dir value, -d value               Directory you want to parse (default: "./")
   --swagger value, -s value           Output the swagger conf for json and yaml (default: "./docs/swagger")
   --propertyStrategy value, -p value  Property Naming Strategy like snakecase,camelcase,pascalcase (default: "camelcase")
```

# General API Info

**Example**  
[celler/main.go](https://github.com/swaggo/swag/blob/master/example/celler/main.go)

| annotation         | description                                                                                     | example                                                         |
|--------------------|-------------------------------------------------------------------------------------------------|-----------------------------------------------------------------|
| title              | **Required.** The title of the application.                                                     | // @title Swagger Example API                                   |
| version            | **Required.** Provides the version of the application API.                                      | // @version 1.0                                                 |
| description        | A short description of the application.                                                         | // @description This is a sample server celler server.          |
| termsOfService     | The Terms of Service for the API.                                                               | // @termsOfService http://swagger.io/terms/                     |
| contact.name       | The contact information for the exposed API.                                                    | // @contact.name API Support                                    |
| contact.url        | The URL pointing to the contact information. MUST be in the format of a URL.                    | // @contact.url http://www.swagger.io/support                   |
| contact.email      | The email address of the contact person/organization. MUST be in the format of an email address.| // @contact.email support@swagger.io                            |
| license.name       | **Required.** The license name used for the API.                                                | // @license.name Apache 2.0                                     |
| license.url        | A URL to the license used for the API. MUST be in the format of a URL.                          | // @license.url http://www.apache.org/licenses/LICENSE-2.0.html |
| host               | The host (name or ip) serving the API.                                                          | // @host localhost:8080                                         |
| BasePath           | The base path on which the API is served.                                                       | // @BasePath /api/v1                                            |

## Security

| annotation                              | description                                                                                    | parameters                        | example                                                      |
|-----------------------------------------|------------------------------------------------------------------------------------------------|-----------------------------------|--------------------------------------------------------------|
| securitydefinitions.basic               | [Basic](https://swagger.io/docs/specification/2-0/authentication/basic-authentication/) auth.  |                                   | // @securityDefinitions.basic BasicAuth                      |
| securitydefinitions.apikey              | [API key](https://swagger.io/docs/specification/2-0/authentication/api-keys/) auth.            | in, name                          | // @securityDefinitions.apikey ApiKeyAuth                    |
| securitydefinitions.oauth2.application  | [OAuth2 application](https://swagger.io/docs/specification/authentication/oauth2/) auth.       | tokenUrl, scope                   | // @securitydefinitions.oauth2.application OAuth2Application |
| securitydefinitions.oauth2.implicit     | [OAuth2 implicit](https://swagger.io/docs/specification/authentication/oauth2/) auth.          | authorizationUrl, scope           | // @securitydefinitions.oauth2.implicit OAuth2Implicit       |
| securitydefinitions.oauth2.password     | [OAuth2 password](https://swagger.io/docs/specification/authentication/oauth2/) auth.          | tokenUrl, scope                   | // @securitydefinitions.oauth2.password OAuth2Password       |
| securitydefinitions.oauth2.accessCode   | [OAuth2 access code](https://swagger.io/docs/specification/authentication/oauth2/) auth.       | tokenUrl, authorizationUrl, scope | // @securitydefinitions.oauth2.accessCode OAuth2AccessCode   |


| parameters annotation | example                                                  |
|-----------------------|----------------------------------------------------------|
| in                    | // @in header                                            |
| name                  | // @name Authorization                                   |
| tokenUrl              | // @tokenUrl https://example.com/oauth/token             |
| authorizationurl      | // @authorizationurl https://example.com/oauth/authorize |
| scope.hoge            | // @scope.write Grants write access                      |

# API Operation

**Example**  
[celler/controller](https://github.com/swaggo/swag/tree/master/example/celler/controller)


| annotation         | description                                                                                                                |
|--------------------|----------------------------------------------------------------------------------------------------------------------------|
| description        | A verbose explanation of the operation behavior.                                                                           |
| id                 | A unique string used to identify the operation. Must be unique among all API operations.                                   |
| tags               | A list of tags to each API operation that separated by commas.                                                             |
| summary            | A short summary of what the operation does.                                                                                |
| accept             | A list of MIME types the APIs can consume. Value MUST be as described under [Mime Types](#mime-types).                     |
| produce            | A list of MIME types the APIs can produce. Value MUST be as described under [Mime Types](#mime-types).                     |
| param              | Parameters that separated by spaces. `param name`,`param type`,`data type`,`is mandatory?`,`comment` `attribute(optional)` |
| security           | [Security](#security) to each API operation.                                                                               |
| success            | Success response that separated by spaces. `return code`,`{param type}`,`data type`,`comment`                              |
| failure            | Failure response that separated by spaces. `return code`,`{param type}`,`data type`,`comment`                              |
| router             | Path definition that separated by spaces. `path`,`[httpMethod]`                                                           |

## Mime Types

| Mime Type                         | annotation                                                |
|-----------------------------------|-----------------------------------------------------------|
| application/json                  | application/json, json                                    |
| text/xml                          | text/xml, xml                                             |
| text/plain                        | text/plain, plain                                         |
| html                              | text/html, html                                           |
| multipart/form-data               | multipart/form-data, mpfd                                 |
| application/x-www-form-urlencoded | application/x-www-form-urlencoded, x-www-form-urlencoded  |
| application/vnd.api+json          | application/vnd.api+json, json-api                        |
| application/x-json-stream         | application/x-json-stream, json-stream                    |
| application/octet-stream          | application/octet-stream, octet-stream                    |
| image/png                         | image/png, png                                            |
| image/jpeg                        | image/jpeg, jpeg                                          |
| image/gif                         | image/gif, gif                                            |

## Security

General API info.

```go
// @securityDefinitions.basic BasicAuth

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information
```

Each API operation.

```go
// @Security ApiKeyAuth
```

Make it AND condition

```go
// @Security ApiKeyAuth
// @Security OAuth2Application[write, admin]
```

## Param Type

- object (struct)
- string (string)
- integer (int, uint, uint32, uint64)
- number (float32)
- boolean (bool)
- array

## Data Type

- string (string)
- integer (int, uint, uint32, uint64)
- number (float32)
- boolean (bool)
- user defined struct

## Attribute

```go
// @Param enumstring query string false "string enums" Enums(A, B, C)
// @Param enumint query int false "int enums" Enums(1, 2, 3)
// @Param enumnumber query number false "int enums" Enums(1.1, 1.2, 1.3)
// @Param string query string false "string valid" minlength(5) maxlength(10)
// @Param int query int false "int valid" mininum(1) maxinum(10)
// @Param default query string false "string default" default(A)
```

### Available

Field Name | Type | Description
---|:---:|---
<a name="parameterDefault"></a>default | * | Declares the value of the parameter that the server will use if none is provided, for example a "count" to control the number of results per page might default to 100 if not supplied by the client in the request. (Note: "default" has no meaning for required parameters.)  See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-6.2. Unlike JSON Schema this value MUST conform to the defined [`type`](#parameterType) for this parameter.
<a name="parameterMaximum"></a>maximum | `number` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.1.2.
<a name="parameterMinimum"></a>minimum | `number` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.1.3.
<a name="parameterMaxLength"></a>maxLength | `integer` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.2.1.
<a name="parameterMinLength"></a>minLength | `integer` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.2.2.
<a name="parameterEnums"></a>enums | [\*] | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.5.1.

### Future

Field Name | Type | Description
---|:---:|---
<a name="parameterFormat"></a>format | `string` | The extending format for the previously mentioned [`type`](#parameterType). See [Data Type Formats](#dataTypeFormat) for further details.
<a name="parameterMultipleOf"></a>multipleOf | `number` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.1.1.
<a name="parameterPattern"></a>pattern | `string` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.2.3.
<a name="parameterMaxItems"></a>maxItems | `integer` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.3.2.
<a name="parameterMinItems"></a>minItems | `integer` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.3.3.
<a name="parameterUniqueItems"></a>uniqueItems | `boolean` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.3.4.
<a name="parameterCollectionFormat"></a>collectionFormat | `string` | Determines the format of the array if type array is used. Possible values are: <ul><li>`csv` - comma separated values `foo,bar`. <li>`ssv` - space separated values `foo bar`. <li>`tsv` - tab separated values `foo\tbar`. <li>`pipes` - pipe separated values <code>foo&#124;bar</code>. <li>`multi` - corresponds to multiple parameter instances instead of multiple values for a single instance `foo=bar&foo=baz`. This is valid only for parameters [`in`](#parameterIn) "query" or "formData". </ul> Default value is `csv`.

## TIPS

### User defined structure with an array type

```go
// @Success 200 {array} model.Account <-- This is a user defined struct.
```

```go
package model

type Account struct {
    ID   int    `json:"id" example:"1"`
    Name string `json:"name" example:"account name"`
}
```

### Use multiple path params

```go
/// ...
// @Param group_id path int true "Group ID"
// @Param account_id path int true "Account ID"
// ...
// @Router /examples/groups/{group_id}/accounts/{account_id} [get]
```

### Example value of struct

```go
type Account struct {
    ID   int    `json:"id" example:"1"`
    Name string `json:"name" example:"account name"`
    PhotoUrls []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
}
```

### Description of struct

```go
type Account struct {
    // ID this is userid
    ID   int    `json:"id"
}
```

## About the Project
This project was inspired by [yvasiyarov/swagger](https://github.com/yvasiyarov/swagger) but we simplified the usage and added support a variety of [web frameworks](#supported-web-frameworks).
