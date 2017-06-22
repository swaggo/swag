# swag
Automatically generate RESTful API documentation with Swagger 2.0 for Go (This project was stll in development). 
This project was inspired by [swagger](https://raw.githubusercontent.com/yvasiyarov/swagger)but simplified the usage of complexity and support a variety of [web framework]((#supported-web-framework)). Let you focus on writing [Declarative Comments Format](#declarative-comments-format).

[![Travis branch](https://img.shields.io/travis/swag-gonic/swag/master.svg)](https://travis-ci.org/swag-gonic/swag)
[![Codecov branch](https://img.shields.io/codecov/c/github/swag-gonic/swag/master.svg)](https://codecov.io/gh/swag-gonic/swag)
[![Go Report Card](https://goreportcard.com/badge/github.com/swag-gonic/swag)](https://goreportcard.com/report/github.com/swag-gonic/swag)
[![GoDoc](https://godoc.org/github.com/swag-gonic/swag?status.svg)](https://godoc.org/github.com/swag-gonic/swag)


## Contents
- [Quick Start Guide](#quick-start-guide)
- [Declarative Comments Format](#declarative-comments-format)
  - [General API info](#general-api-info)
  - [API Operation](#api-operation)
- [Supported Web Framework](#supported-web-framework)
- [Features](#features)

## Quick Start Guide

1. Add comments to your API source code, [see Declarative Comments Format](#declarative-comments-format)

2. Download Swag for Go by using:
```sh
$ go get -u github.com/swag-gonic/swag
```
3. Run the Swag in your Go project root folder which contains `main.go` file, Swag will parse your comments and generate required files(`docs` folder and `docs/doc.go`)
```sh
$ swag init
```
4. Open your `main.go` file, import it in your code:
                            
 `import github.com/swag-gonic/gin-swagger` 
TODO:

## Declarative Comments Format

### General API info
| annotation         | Description                                                                                               | 
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
| annotation         | Description                                                                                               | 
|--------------------|-----------------------------------------------------------------------------------------------------------|
| description        | A verbose explanation of the operation behavior.                                                          |
| summary            | A short summary of what the operation does.                                                               |
| accept             | A list of MIME types the APIs can consume. Now only `json` application type.                              | 
| produce            | A list of MIME types the APIs can produce. Now only `json` application type.                              | 
| param              | Parameters that separated by spaces. `param name`,`param type`,`data type`,`is mandatory?`,`comment`      | 
| success            | Success response that separated by spaces. `return code`,`{param type}`,`data type`,`comment`             | 
| failure            | Failure response that separated by spaces. `return code`,`{param type}`,`data type`,`comment`             | 
| router             | Failure response that separated by spaces. `path`,`[httpMethod]`                                          | 



## Supported Web Framework
- [gin-swagger](http://github.com/swag-gonic/gin-swagger)
- [echo-swagger](http://github.com/swag-gonic/gin-swagger)

## TODO
- [ ] support other Mime Types, eg: xml