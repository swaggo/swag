# swag
Automatically generate RESTful API documentation with Swagger 2.0 for Go.

[![Travis branch](https://img.shields.io/travis/xykong/swag/master.svg)](https://travis-ci.org/xykong/swag)
[![Codecov branch](https://img.shields.io/codecov/c/github/xykong/swag/master.svg)](https://codecov.io/gh/xykong/swag)
[![Go Report Card](https://goreportcard.com/badge/github.com/xykong/swag)](https://goreportcard.com/report/github.com/xykong/swag)
[![GoDoc](https://godoc.org/github.com/swaggo/swagg?status.svg)](https://godoc.org/github.com/xykong/swag)
 
## What is swag?
swag converts Go annotations to Swagger Documentation 2.0. And provides a variety of builtin [web framework](#supported-web-framework) lib. Let you can quickly integrated in existing golang project(using Swagger UI) .

## Contents
- [Generate Swagger 2.0 docs](#generate-swagger-20-docs)
- [How to use it with gin?](#how-to-use-it-with-gin)
- [Declarative Comments Format](#declarative-comments-format)
  - [General API info](#general-api-info)
  - [API Operation](#api-operation)
- [Supported Web Framework](#supported-web-framework)


## Generate Swagger 2.0 docs
1. Add comments to your API source code, [See Declarative Comments Format](#declarative-comments-format)

2. Download swag by using:
```sh
$ go get -u github.com/xykong/swag/cmd/swag
```
3. Run the [swag](#generate-swagger-20-docs) in project root folder which contains `main.go` file, The [swag](#generate-swagger-20-docs) will parse your comments and generate required files(`docs` folder and `docs/doc.go`).
```sh
$ swag init
```

## How to use it with `gin`? 
1. After using [swag](#generate-swagger-20-docs) to generate Swagger 2.0 docs. Import following packages:
```go
import "github.com/swaggo/gin-swagger" // gin-swagger middleware
import "github.com/swaggo/gin-swagger/swaggerFiles" // swagger embed files

```

2. Added [API Operation](#api-operation) annotations in `main.go` code:
```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	_ "github.com/swaggo/gin-swagger/example/docs" // docs is generated by Swag CLI, you have to import it.
)

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
func main() {
	r := gin.New()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run()
}
```

3. Added [General API Info](#api-operation) annotations in `handler/controller` code
``` go 
// @Summary Add a new pet to the store
// @Description get string by ID
// @ID get-string-by-int
// @Accept  json
// @Produce  json
// @Param   some_id     path    int     true        "Some ID"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /testapi/get-string-by-int/{some_id} [get]
func GetStringByInt(c *gin.Context) {
	//write your code
}

// @Description get struct array by ID
// @ID get-struct-array-by-string
// @Accept  json
// @Produce  json
// @Param   some_id     path    string     true        "Some ID"
// @Param   offset     query    int     true        "Offset"
// @Param   limit      query    int     true        "Offset"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /testapi/get-struct-array-by-string/{some_id} [get]
func GetStructArrayByString(c *gin.Context) {
	//write your code
}

type Pet3 struct {
	ID int `json:"id"`
}

```

4. Run it, and browser to http://localhost:8080/swagger/index.html. You will see Swagger 2.0 Api documents as bellow:

![swagger_index.html](https://user-images.githubusercontent.com/8943871/31943004-dd08a10e-b88c-11e7-9e77-19d2c759a586.png)



## Declarative Comments Format

### General API Info
| annotation         | description                                                                                               | 
|--------------------|-----------------------------------------------------------------------------------------------------------|
| title              | **Required.** The title of the application.                                                               |
| version            | **Required.** Provides the version of the application API.                                                |
| description        | A short description of the application.                                                                   |
| termsOfService     | The Terms of Service for the API.                                                                         |
| contact.name       | The contact information for the exposed API.                                                              |
| contact.url        | The URL pointing to the contact information. MUST be in the format of a URL.                              |
| contact.email      | The email address of the contact person/organization. MUST be in the format of an email address.          |
| license.name       | **Required.** The license name used for the API.                                                          |
| license.url        | A URL to the license used for the API. MUST be in the format of a URL.                                    |
| host               | The host (name or ip) serving the API.                                                                    |
| BasePath           | The base path on which the API is served.                                                                 |


### API Operation
| annotation         | description                                                                                               | 
|--------------------|-----------------------------------------------------------------------------------------------------------|
| description        | A verbose explanation of the operation behavior.                                                          |
| id                 | A unique string used to identify the operation. Must be unique among all API operations.                      |
| summary            | A short summary of what the operation does.                                                               |
| accept             | A list of MIME types the APIs can consume. Now only `json` application type.                              | 
| produce            | A list of MIME types the APIs can produce. Now only `json` application type.                              | 
| param              | Parameters that separated by spaces. `param name`,`param type`,`data type`,`is mandatory?`,`comment`      | 
| success            | Success response that separated by spaces. `return code`,`{param type}`,`data type`,`comment`             | 
| failure            | Failure response that separated by spaces. `return code`,`{param type}`,`data type`,`comment`             | 
| router             | Failure response that separated by spaces. `path`,`[httpMethod]`                                          | 



## Supported Web Framework
- [gin-swagger](http://github.com/swaggo/gin-swagger)

## TODO
- [ ] support other Mime Types, eg: xml
- [ ] supplement better documentation
- [ ] add more example
- [ ] support other web Framework


## About the Project
This project was inspired by [swagger](https://raw.githubusercontent.com/yvasiyarov/swagger) but simplified the usage of complexity and support a variety of [web framework]((#supported-web-framework)).

