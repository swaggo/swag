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
- [Supported Web Framework](#supported-web-framework)
- [Features](#features)

## Quick Start Guide

1. Add comments to your API source code, [see Declarative Comments Format](#declarative-comments-format)

2. Download Swag for Go by using:
```sh
$ go get -u github.com/swag-gonic/swag
```
3. Run the Swag in your Go root project folder which contains `main.go` file, Swag will parse your comments and generate required files(`docs` folder and `docs/doc.go`)
```sh
$ swag init
```
4. Open your `main.go` file, add import for Gin user
 `import github.com/swag-gonic/gin-swagger` 

TODO:

## Declarative Comments Format
TODO:

## Supported Web Framework
- [gin-swagger](http://github.com/swag-gonic/gin-swagger)
- [echo-swagger](http://github.com/swag-gonic/gin-swagger)