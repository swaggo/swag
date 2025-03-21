# swag

🌍 *[English](README.md) ∙ [简体中文](README_zh-CN.md) ∙ [Português](README_pt.md)*

<img align="right" width="180px" src="https://raw.githubusercontent.com/swaggo/swag/master/assets/swaggo.png">

[![Build Status](https://github.com/swaggo/swag/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/features/actions)
[![Coverage Status](https://img.shields.io/codecov/c/github/swaggo/swag/master.svg)](https://codecov.io/gh/swaggo/swag)
[![Go Report Card](https://goreportcard.com/badge/github.com/swaggo/swag)](https://goreportcard.com/report/github.com/swaggo/swag)
[![codebeat badge](https://codebeat.co/badges/71e2f5e5-9e6b-405d-baf9-7cc8b5037330)](https://codebeat.co/projects/github-com-swaggo-swag-master)
[![Go Doc](https://godoc.org/github.com/swaggo/swagg?status.svg)](https://godoc.org/github.com/swaggo/swag)
[![Backers on Open Collective](https://opencollective.com/swag/backers/badge.svg)](#backers)
[![Sponsors on Open Collective](https://opencollective.com/swag/sponsors/badge.svg)](#sponsors) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fswaggo%2Fswag.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fswaggo%2Fswag?ref=badge_shield)
[![Release](https://img.shields.io/github/release/swaggo/swag.svg?style=flat-square)](https://github.com/swaggo/swag/releases)

Swag converte anotações Go para Documentação Swagger 2.0. Criámos uma variedade de plugins para populares [Go web frameworks](#supported-web-frameworks). Isto permite uma integração rápida com um projecto Go existente (utilizando a Swagger UI).

## Conteúdo
- [Começando](#começando)
 - [Estruturas Web Suportadas](#estruturas-web-suportadas)
 - [Como utilizá-lo com Gin](#como-como-ser-como-gin)
 - [O formatador de swag](#a-formatação-de-swag)
 - [Estado de Implementação](#implementação-estado)
 - [Formato dos comentários declarativos](#formato-dos-comentarios-declarativos)
	- [Informações Gerais API](#informações-gerais-api)
	- [Operação API](#api-operacao)
	- [Segurança](#seguranca)
 - [Exemplos](#exemplos)
    - [Descrições em múltiplas linhas](#descricoes-sobre-múltiplas-linhas)
	- [Estrutura definida pelo utilizador com um tipo de matriz](#-estrutura-definida-pelo-utilizador-com-um-um-tipo)
	- [Declaração de estruturação de funções](#function-scoped-struct-declaration)
	- [Composição do modelo em resposta](#model-composição-em-resposta)
	- [Adicionar um cabeçalho em resposta](#add-a-headers-in-response)
	- [Utilizar parâmetros de caminhos múltiplos](#use-multiple-path-params)
	- [Exemplo de valor de estrutura](#exemplo-do-valor-de-estrutura)
	- [Schema Exemplo do corpo](#schemaexample-of-body)
	- [Descrição da estrutura](#descrição-da-estrutura)
	- [Usar etiqueta do tipo swaggertype para suportar o tipo personalizado](#use-swaggertype-tag-to-supported-custom-type)
	- [Utilizar anulações globais para suportar um tipo personalizado](#use-global-overrides-to-support-a-custom-type)
	- [Use swaggerignore tag para excluir um campo](#use-swaggerignore-tag-to-excluir-um-campo)
	- [Adicionar informações de extensão ao campo de estruturação](#add-extension-info-to-struct-field)
	- [Renomear modelo a expor](#renome-modelo-a-exibir)
	- [Como utilizar as anotações de segurança](#como-utilizar-as-anotações-de-segurança)
	- [Adicionar uma descrição para enumerar artigos](#add-a-description-for-enum-items)
	- [Gerar apenas tipos de ficheiros de documentos específicos](#generate-only-specific-docs-file-file-types)
    - [Como usar tipos genéricos](#como-usar-tipos-genéricos)
- [Sobre o projecto](#sobre-o-projecto)

## Começando

1. Adicione comentários ao código-fonte da API, consulte [Formato dos comentários declarativos](#declarative-comments-format).

2. Descarregue o swag utilizando:
```sh
go install github.com/swaggo/swag/cmd/swag@latest
```
Para construir a partir da fonte é necessário [Go](https://golang.org/dl/) (1.19 ou mais recente).

Ou descarregar um binário pré-compilado a partir da [página de lançamento](https://github.com/swaggo/swag/releases).

3. Executar `swag init` na pasta raiz do projecto que contém o ficheiro `main.go`. Isto irá analisar os seus comentários e gerar os ficheiros necessários (pasta `docs` e `docs/docs.go`).
```sh
swag init
```

Certifique-se de importar os `docs/docs.go` gerados para que a sua configuração específica fique "init" ed. Se as suas anotações API gerais não viverem em `main.go`, pode avisar a swag com a bandeira `-g`.
```sh
swag init -g http/api.go
```

4. (opcional) Utilizar o formato `swag fmt` no comentário SWAG. (Por favor, actualizar para a versão mais recente)

```sh
swag fmt
```

## swag cli

```sh
swag init -h
NOME:
   swag init - Criar docs.go

UTILIZAÇÃO:
   swag init [opções de comando] [argumentos...]

OPÇÕES:
   --quiet, -q Fazer o logger ficar quiet (por padrão: falso)
   --generalInfo valor, -g valor Go caminho do ficheiro em que 'swagger general API Info' está escrito (por padrão: "main.go")
   --dir valor, -d valor Os directórios que deseja analisar, separados por vírgulas e de informação geral devem estar no primeiro (por padrão: "./")
   --exclude valor Excluir directórios e ficheiros ao pesquisar, separados por vírgulas
   -propertyStrategy da estratégia, -p valor da propriedadeEstratégia de nomeação de propriedades como snakecase,camelcase,pascalcase (por padrão: "camelcase")
   --output de saída, -o valor directório de saída para todos os ficheiros gerados(swagger.json, swagger.yaml e docs.go) (por padrão: "./docs")
   --outputTypes valor de saídaTypes, -- valor de saída Tipos de ficheiros gerados (docs.go, swagger.json, swagger.yaml) como go,json,yaml (por padrão: "go,json,yaml")
   --parseVendor ParseVendor Parse go files na pasta 'vendor', desactivado por padrão (padrão: falso)
   --parseInternal Parse go ficheiros em pacotes internos, desactivados por padrão (padrão: falso)
   --generatedTime Gerar timestamp no topo dos docs.go, desactivado por padrão (padrão: falso)
   --parteDepth value Dependência profundidade parse (por padrão: 100)
   --templateDelims value, --td value fornecem delimitadores personalizados para a geração de modelos Go. O formato é leftDelim,rightDelim. Por exemplo: "[[,]]"
   ...

   --help, -h mostrar ajuda (por padrão: falso)
```

```bash
swag fmt -h
NOME:
   swag fmt - formato swag comentários

UTILIZAÇÃO:
   swag fmt [opções de comando] [argumentos...]

OPÇÕES:
   --dir valor, -d valor Os directórios que pretende analisar, separados por vírgulas e de informação geral devem estar no primeiro (por padrão: "./")
   --excluir valor Excluir directórios e ficheiros ao pesquisar, separados por vírgulas
   --generalInfo value, -g value Go file path in which 'swagger general API Info' is written (por padrão: "main.go")
   --ajuda, -h mostrar ajuda (por padrão: falso)

```

## Estruturas Web Suportadas

- [gin](http://github.com/swaggo/gin-swagger)
- [echo](http://github.com/swaggo/echo-swagger)
- [buffalo](https://github.com/swaggo/buffalo-swagger)
- [net/http](https://github.com/swaggo/http-swagger)
- [gorilla/mux](https://github.com/swaggo/http-swagger)
- [go-chi/chi](https://github.com/swaggo/http-swagger)
- [flamingo](https://github.com/i-love-flamingo/swagger)
- [fiber](https://github.com/gofiber/swagger)
- [atreugo](https://github.com/Nerzal/atreugo-swagger)
- [hertz](https://github.com/hertz-contrib/swagger)

## Como utilizá-lo com Gin

Encontrar o código fonte de exemplo [aqui](https://github.com/swaggo/swag/tree/master/example/celler).

1. Depois de utilizar `swag init` para gerar os documentos Swagger 2.0, importar os seguintes pacotes:
```go
import "github.com/swaggo/gin-swagger" // gin-swagger middleware
import "github.com/swaggo/files" // swagger embed files
```

2. Adicionar [Informações Gerais API](#general-api-info) anotações em código `main.go`:


```go
// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
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

Além disso, algumas informações API gerais podem ser definidas de forma dinâmica. O pacote de código gerado `docs` exporta a variável `SwaggerInfo` que podemos utilizar para definir programticamente o título, descrição, versão, hospedeiro e caminho base. Exemplo utilizando Gin:

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"./docs" // docs is generated by Swag CLI, you have to import it.
)

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
func main() {

	// programmatically set swagger info
	docs.SwaggerInfo.Title = "Swagger Example API"
	docs.SwaggerInfo.Description = "This is a sample server Petstore server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "petstore.swagger.io"
	docs.SwaggerInfo.BasePath = "/v2"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	r := gin.New()

	// use ginSwagger middleware to serve the API docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run()
}
```

3. Adicionar [Operação API](#api-operacao) anotações em código `controller`

```go
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
// @Summary      Show an account
// @Description  get string by ID
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  model.Account
// @Failure      400  {object}  httputil.HTTPError
// @Failure      404  {object}  httputil.HTTPError
// @Failure      500  {object}  httputil.HTTPError
// @Router       /accounts/{id} [get]
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
// @Summary      List accounts
// @Description  get accounts
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        q    query     string  false  "name search by q"  Format(email)
// @Success      200  {array}   model.Account
// @Failure      400  {object}  httputil.HTTPError
// @Failure      404  {object}  httputil.HTTPError
// @Failure      500  {object}  httputil.HTTPError
// @Router       /accounts [get]
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
swag init
```

4. Execute a sua aplicação, e navegue para http://localhost:8080/swagger/index.html. Verá os documentos Swagger 2.0 Api, como mostrado abaixo:

![swagger_index.html](https://raw.githubusercontent.com/swaggo/swag/master/assets/swagger-image.png)

## O formatador de swag

Os Swag Comments podem ser formatados automaticamente, assim como 'go fmt'.
Encontre o resultado da formatação [aqui](https://github.com/swaggo/swag/tree/master/example/celler).

Usage:
```shell
swag fmt
```

Exclude folder：
```shell
swag fmt -d ./ --exclude ./internal
```

Ao utilizar `swag fmt`, é necessário assegurar-se de que tem um comentário doc para a função a fim de assegurar uma formatação correcta.
Isto deve-se ao `swag fmt` que traça comentários swag com separadores, o que só é permitido *após* um comentário doc padrão.

Por exemplo, utilizar

```go
// ListAccounts lists all existing accounts
//
//  @Summary      List accounts
//  @Description  get accounts
//  @Tags         accounts
//  @Accept       json
//  @Produce      json
//  @Param        q    query     string  false  "name search by q"  Format(email)
//  @Success      200  {array}   model.Account
//  @Failure      400  {object}  httputil.HTTPError
//  @Failure      404  {object}  httputil.HTTPError
//  @Failure      500  {object}  httputil.HTTPError
//  @Router       /accounts [get]
func (c *Controller) ListAccounts(ctx *gin.Context) {
```

## Estado de Implementação

[Documento Swagger 2.0](https://swagger.io/docs/specification/2-0/basic-structure/)

- [x] Estrutura básica
- [x] Hospedeiro API e Caminho Base
- [x] Caminhos e operações
- [x] Descrição dos parâmetros
- [x] Descrever o corpo do pedido
- [x] Descrição das respostas
- [x] Tipos MIME
- [x] Autenticação
  - [x] Autenticação básica
  - [x] Chaves API
- [x] Acrescentar exemplos
- [x] Carregamento de ficheiros
- [x] Enums
- [x] Operações de Agrupamento com Etiquetas
- Extensões Swagger

## Formato dos comentários declarativos

## Informações Gerais API

**Exemplo**
[celler/main.go](https://github.com/swaggo/swag/blob/master/example/celler/main.go)

| anotação | descrição | exemplo |
|-------------|--------------------------------------------|---------------------------------|
| title | **Obrigatório.** O título da aplicação.| // @title Swagger Example API |
| version | **Obrigatório.** Fornece a versão da aplicação API.| // @version 1.0 |
| description | Uma breve descrição da candidatura.    |// @descrição Este é um servidor servidor de celas de amostra.         																 |
| tag.name | Nome de uma tag.| // @tag.name Este é o nome da tag |
| tag.description | Descrição da tag | // @tag.description Cool Description |
| tag.docs.url | Url da Documentação externa da tag | // @tag.docs.url https://example.com|
| tag.docs.description | Descrição da documentação externa da tag| // @tag.docs.description Melhor exemplo de documentação |
| TermsOfService | Os Termos de Serviço para o API.| // @termsOfService http://swagger.io/terms/ |
| contact.name | A informação de contacto para a API exposta.| // @contacto.name Suporte API |
| contact.url | O URL que aponta para as informações de contacto. DEVE estar no formato de um URL.  | // @contact.url http://www.swagger.io/support|
| contact.email| O endereço de email da pessoa/organização de contacto. DEVE estar no formato de um endereço de correio electrónico.| // @contact.email support@swagger.io |
| license.name | **Obrigatório.** O nome da licença utilizada para a API.|// @licença.name Apache 2.0|
| license.url | Um URL para a licença utilizada para a API. DEVE estar no formato de um URL.                       | // @license.url http://www.apache.org/licenses/LICENSE-2.0.html |
| host | O anfitrião (nome ou ip) que serve o API.     | // @host localhost:8080 |
| BasePath | O caminho de base sobre o qual o API é servido. | // @BasePath /api/v1 |
| accept | Uma lista de tipos de MIME que os APIs podem consumir. Note que accept só afecta operações com um organismo de pedido, tais como POST, PUT e PATCH.  O valor DEVE ser o descrito em [Tipos de Mime](#mime-types).                     | // @accept json |
| produce | Uma lista de tipos de MIME que os APIs podem produce. O valor DEVE ser o descrito em [Tipos de Mime](#mime-types).                     | // @produce json |
| query.collection.format | O formato padrão de param de colecção(array) em query,enums:csv,multi,pipes,tsv,ssv. Se não definido, csv é o padrão.| // @query.collection.format multi
| schemes | O protocolo de transferência para a operação que separou por espaços. | // @schemes http https |
| externalDocs.description | Descrição do documento externo. | // @externalDocs.description OpenAPI |
| externalDocs.url | URL do documento externo. | // @externalDocs.url https://swagger.io/resources/open-api/ |
| x-name | A chave de extensão, deve ser iniciada por x- e tomar apenas o valor json | // @x-example-key {"chave": "valor"} |

### Usando descrições de remarcação para baixo
Quando uma pequena sequência na sua documentação é insuficiente, ou precisa de imagens, exemplos de códigos e coisas do género, pode querer usar descrições de marcação. Para utilizar as descrições markdown, utilize as seguintes anotações.

| anotação | descrição | exemplo |
|-------------|--------------------------------------------|---------------------------------|
| title | **Obrigatório.** O título da aplicação.| // @title Swagger Example API |
| version | **Obrigatório.** Fornece a versão da aplicação API.| // @versão 1.0 |
| description.markdown | Uma breve descrição da candidatura. Parsed a partir do ficheiro api.md. Esta é uma alternativa a @description |// @description.markdown Sem valor necessário, isto analisa a descrição do ficheiro api.md |.
| tag.name | Nome de uma tag.| // @tag.name Este é o nome da tag |
| tag.description.markdown | Descrição da tag esta é uma alternativa à tag.description. A descrição será lida a partir de um ficheiro nomeado como tagname.md | // @tag.description.markdown |

## Operação API

**Exemplo**
[celler/controller](https://github.com/swaggo/swag/tree/master/example/celler/controller)

| anotação | descrição |
|-------------|----------------------------------------------------------------------------------------------------------------------------|
| descrição | Uma explicação verbosa do comportamento da operação.                                                                           |
| description.markdown | Uma breve descrição da candidatura. A descrição será lida a partir de um ficheiro.  Por exemplo, `@description.markdown details` irá carregar `details.md`| // @description.file endpoint.description.markdown |
| id | Um fio único utilizado para identificar a operação. Deve ser única entre todas as operações API.                                   |
| tags | Uma lista de tags para cada operação API que separou por vírgulas.                                                             |
| summary | Um breve resumo do que a operação faz.                                                                                |
| accept | Uma lista de tipos de MIME que os APIs podem consumir. Note que accept só afecta operações com um organismo de pedido, tais como POST, PUT e PATCH.  O valor DEVE ser o descrito em [Tipos de Mime](#mime-types).                     |
| produce | Uma lista de tipos de MIME que os APIs podem produce. O valor DEVE ser o descrito em [Tipos de Mime](#mime-types).                     |
| param | Parâmetros que se separaram por espaços. `param name`,`param type`,`data type`,`is mandatory?`,`comment` `attribute(optional)` |
| security | [Segurança](#security) para cada operação API.                                                                               |
| success | resposta de sucesso que separou por espaços. `return code or default`,`{param type}`,`data type`,`comment` |.
| failure | Resposta de falha que separou por espaços. `return code or default`,`{param type}`,`data type`,`comment` |
| response | Igual ao `sucesso` e `falha` |
| header | Cabeçalho em resposta que separou por espaços. `código de retorno`,`{tipo de parâmetro}`,`tipo de dados`,`comentário` |.
| router | Definição do caminho que separou por espaços. caminho",`path`,`[httpMethod]` |[httpMethod]` |
| x-name | A chave de extensão, deve ser iniciada por x- e tomar apenas o valor json.                                                           |
| x-codeSample | Optional Markdown use. tomar `file` como parâmetro. Isto irá então procurar um ficheiro nomeado como o resumo na pasta dada.                                      |
| deprecated | Marcar o ponto final como depreciado.                                                                                               |

## Mime Types

`swag` aceita todos os tipos MIME que estão no formato correcto, ou seja, correspondem `*/*`.
Além disso, `swag` também aceita pseudónimos para alguns tipos de MIME, como se segue:


| Alias                 | MIME Type                         |
|-----------------------|-----------------------------------|
| json                  | application/json                  |
| xml                   | text/xml                          |
| plain                 | text/plain                        |
| html                  | text/html                         |
| mpfd                  | multipart/form-data               |
| x-www-form-urlencoded | application/x-www-form-urlencoded |
| json-api              | application/vnd.api+json          |
| json-stream           | application/x-json-stream         |
| octet-stream          | application/octet-stream          |
| png                   | image/png                         |
| jpeg                  | image/jpeg                        |
| gif                   | image/gif                         |
| event-stream          | text/event-stream                 |



## Tipo de parâmetro

- query
- path
- header
- body
- formData

## Tipo de dados

- string (string)
- integer (int, uint, uint32, uint64)
- number (float32)
- boolean (bool)
- file (param data type when uploading)
- user defined struct

## Segurança
| anotação | descrição | parâmetros | exemplo |
|------------|-------------|------------|---------|
| securitydefinitions.basic | [Basic](https://swagger.io/docs/specification/2-0/authentication/basic-authentication/) auth.  | | // @securityDefinitions.basicAuth | [Básico]()
| securitydefinitions.apikey | [chave API](https://swagger.io/docs/specification/2-0/authentication/api-keys/) auth.            | in, name, description | // @securityDefinitions.apikey ApiKeyAuth |
| securitydefinitions.oauth2.application | [Aplicação OAuth2](https://swagger.io/docs/specification/authentication/oauth2/) auth.       | tokenUrl, scope, description | // @securitydefinitions.oauth2.application OAuth2Application |
| securitydefinitions.oauth2.implicit | [OAuth2 implicit](https://swagger.io/docs/specification/authentication/oauth2/) auth.          | authorizationUrl, scope, description | // @securitydefinitions.oauth2.implicit OAuth2Implicit | [OAuth2Implicit]()
| securitydefinitions.oauth2.password | [OAuth2 password](https://swagger.io/docs/specification/authentication/oauth2/) auth.          | tokenUrl, scope, description | // @securitydefinitions.oauth2.password OAuth2Password |
| securitydefinitions.oauth2.accessCode | [código de acesso OAuth2](https://swagger.io/docs/specification/authentication/oauth2/) auth.       | tokenUrl, authorizationUrl, scope, description | // @securitydefinitions.oauth2.accessCode OAuth2AccessCode | [código de acesso OAuth2.accessCode]()


| anotação de parâmetros | exemplo |
|---------------------------------|-------------------------------------------------------------------------|
| in | // @in header |
| name | // @name Authorization |
| tokenUrl | // @tokenUrl https://example.com/oauth/token |
| authorizationurl | // @authorizationurl https://example.com/oauth/authorize |
| scope.hoge | // @scope.write Grants write access |
| description | // @descrição OAuth protege os pontos finais da nossa entidade |

## Atributo

```go
// @Param   enumstring  query     string     false  "string enums"       Enums(A, B, C)
// @Param   enumint     query     int        false  "int enums"          Enums(1, 2, 3)
// @Param   enumnumber  query     number     false  "int enums"          Enums(1.1, 1.2, 1.3)
// @Param   string      query     string     false  "string valid"       minlength(5)  maxlength(10)
// @Param   int         query     int        false  "int valid"          minimum(1)    maximum(10)
// @Param   default     query     string     false  "string default"     default(A)
// @Param   example     query     string     false  "string example"     example(string)
// @Param   collection  query     []string   false  "string collection"  collectionFormat(multi)
// @Param   extensions  query     []string   false  "string collection"  extensions(x-example=test,x-nullable)
```

It also works for the struct fields:

```go
type Foo struct {
    Bar string `minLength:"4" maxLength:"16" example:"random string"`
    Baz int `minimum:"10" maximum:"20" default:"15"`
    Qux []string `enums:"foo,bar,baz"`
}
```

### Disponível

Nome do campo | Tipo | Descrição
---|:---:|---
<a name="validate"></a>validate | `string` | Determina a validação para o parâmetro. Os valores possíveis são: `required,optional`.
<a name="parameterDefault"></a>default | * | Declara o valor do parâmetro que o servidor utilizará se nenhum for fornecido, por exemplo, uma "contagem" para controlar o número de resultados por página poderá ser por defeito de 100 se não for fornecido pelo cliente no pedido. (Nota: "por defeito" não tem significado para os parâmetros requeridos).
See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-6.2. Ao contrário do esquema JSON, este valor DEVE estar em conformidade com o definido [`type`](#parameterType) para este parâmetro.
<a name="parameterMaximum"></a>maximum | `number` | Ver https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.1.2.
<a name="parameterMinimum"></a>minimum | `number` | Ver https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.1.3.
<a name="parameterMultipleOf"></a>multipleOf | `number` | Ver https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.1.1.
<a name="parameterMaxLength"></a>maxLength | `integer` | Ver https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.2.1.
<a name="parameterMinLength"></a>minLength | `integer` | Ver https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.2.2.
<a name="parameterEnums"></a>enums | [\*] | Ver https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.5.1.
<a name="parameterFormat"></a>format | `string` | O formato de extensão para o anteriormente mencionado [`type`](#parameterType). Ver [Data Type Formats](https://swagger.io/specification/v2/#dataTypeFormat) para mais detalhes.
<a name="parameterCollectionFormat"></a>collectionFormat | `string` |Determina o formato da matriz se for utilizada uma matriz de tipos. Os valores possíveis são: <ul><li>`csv` - valores separados por vírgulas `foo,bar`. <li>`ssv` - valores separados por espaço `foo bar`. <li>`tsv` - valores separados por tabulação `foo\tbar`. <li>`pipes` - valores separados por tubo <code>foo&#124;bar</code>. <li>`multi` - corresponde a múltiplas instâncias de parâmetros em vez de múltiplos valores para uma única instância `foo=bar&foo=baz`. This is valid only for parameters [`in`](#parameterIn) "query" or "formData". </ul> Default value is `csv`.
<a name="parameterExample"></a>example | * | Declara o exemplo para o valor do parâmetro
<a name="parameterExtensions"></a>extensions | `string` | Acrescentar extensão aos parâmetros.

### Futuro

Nome do campo | Tipo | Description
---|:---:|---
<a name="parameterPattern"></a>pattern | `string` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.2.3.
<a name="parameterMaxItems"></a>maxItems | `integer` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.3.2.
<a name="parameterMinItems"></a>minItems | `integer` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.3.3.
<a name="parameterUniqueItems"></a>uniqueItems | `boolean` | See https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.3.4.

## Exemplos


### Descrições em múltiplas linhas

É possível acrescentar descrições que abranjam várias linhas tanto na descrição geral da api como em definições de rotas como esta:

```go
// @description This is the first line
// @description This is the second line
// @description And so forth.
```

### Estrutura definida pelo utilizador com um tipo de matriz

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


### Declaração de estruturação de funções

Pode declarar as estruturas de resposta do seu pedido dentro de um corpo funcional.
Deve ter de seguir a convenção de nomeação
`<package-name>.<function-name>.<struct-name> `.

```go
package main

// @Param request body main.MyHandler.request true "query params"
// @Success 200 {object} main.MyHandler.response
// @Router /test [post]
func MyHandler() {
	type request struct {
		RequestField string
	}

	type response struct {
		ResponseField string
	}
}
```


### Composição do modelo em resposta
```go
// JSONResult's data field will be overridden by the specific type proto.Order
@success 200 {object} jsonresult.JSONResult{data=proto.Order} "desc"
```

```go
type JSONResult struct {
    Code    int          `json:"code" `
    Message string       `json:"message"`
    Data    interface{}  `json:"data"`
}

type Order struct { //in `proto` package
    Id  uint            `json:"id"`
    Data  interface{}   `json:"data"`
}
```

- também suportam uma variedade de objectos e tipos primitivos como resposta aninhada
```go
@success 200 {object} jsonresult.JSONResult{data=[]proto.Order} "desc"
@success 200 {object} jsonresult.JSONResult{data=string} "desc"
@success 200 {object} jsonresult.JSONResult{data=[]string} "desc"
```

- campos múltiplos que se sobrepõem. campo será adicionado se não existir
```go
@success 200 {object} jsonresult.JSONResult{data1=string,data2=[]string,data3=proto.Order,data4=[]proto.Order} "desc"
```
- overriding deep-level fields
```go
type DeepObject struct { //in `proto` package
	...
}
@success 200 {object} jsonresult.JSONResult{data1=proto.Order{data=proto.DeepObject},data2=[]proto.Order{data=[]proto.DeepObject}} "desc"
```

### Adicionar um cabeçalho em resposta

```go
// @Success      200              {string}  string    "ok"
// @failure      400              {string}  string    "error"
// @response     default          {string}  string    "other error"
// @Header       200              {string}  Location  "/entity/1"
// @Header       200,400,default  {string}  Token     "token"
// @Header       all              {string}  Token2    "token2"
```


### Utilizar parâmetros de caminhos múltiplos

```go
/// ...
// @Param group_id   path int true "Group ID"
// @Param account_id path int true "Account ID"
// ...
// @Router /examples/groups/{group_id}/accounts/{account_id} [get]
```

### Adicionar múltiplos caminhos

```go
/// ...
// @Param group_id path int true "Group ID"
// @Param user_id  path int true "User ID"
// ...
// @Router /examples/groups/{group_id}/user/{user_id}/address [put]
// @Router /examples/user/{user_id}/address [put]
```

### Exemplo de valor de estrutura

```go
type Account struct {
    ID   int    `json:"id" example:"1"`
    Name string `json:"name" example:"account name"`
    PhotoUrls []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
}
```

### Schema Exemplo do corpo

```go
// @Param email body string true "message/rfc822" SchemaExample(Subject: Testmail\r\n\r\nBody Message\r\n)
```

### Descrição da estrutura

```go
// Account model info
// @Description User account information
// @Description with user id and username
type Account struct {
	// ID this is userid
	ID   int    `json:"id"`
	Name string `json:"name"` // This is Name
}
```

[#708](https://github.com/swaggo/swag/issues/708) O analisador trata apenas de comentários estruturais a partir de `@Description` attribute.

Assim, gerou o doc. de swagger como se segue:
```json
"Account": {
  "type":"object",
  "description": "User account information with user id and username"
  "properties": {
    "id": {
      "type": "integer",
      "description": "ID this is userid"
    },
    "name": {
      "type":"string",
      "description": "This is Name"
    }
  }
}
```

### Usar etiqueta do tipo swaggertype para suportar o tipo personalizado
[#201](https://github.com/swaggo/swag/issues/201#issuecomment-475479409)

```go
type TimestampTime struct {
    time.Time
}

///implement encoding.JSON.Marshaler interface
func (t *TimestampTime) MarshalJSON() ([]byte, error) {
    bin := make([]byte, 16)
    bin = strconv.AppendInt(bin[:0], t.Time.Unix(), 10)
    return bin, nil
}

func (t *TimestampTime) UnmarshalJSON(bin []byte) error {
    v, err := strconv.ParseInt(string(bin), 10, 64)
    if err != nil {
        return err
    }
    t.Time = time.Unix(v, 0)
    return nil
}
///

type Account struct {
    // Override primitive type by simply specifying it via `swaggertype` tag
    ID     sql.NullInt64 `json:"id" swaggertype:"integer"`

    // Override struct type to a primitive type 'integer' by specifying it via `swaggertype` tag
    RegisterTime TimestampTime `json:"register_time" swaggertype:"primitive,integer"`

    // Array types can be overridden using "array,<prim_type>" format
    Coeffs []big.Float `json:"coeffs" swaggertype:"array,number"`
}
```

[#379](https://github.com/swaggo/swag/issues/379)
```go
type CerticateKeyPair struct {
	Crt []byte `json:"crt" swaggertype:"string" format:"base64" example:"U3dhZ2dlciByb2Nrcw=="`
	Key []byte `json:"key" swaggertype:"string" format:"base64" example:"U3dhZ2dlciByb2Nrcw=="`
}
```
generated swagger doc as follows:
```go
"api.MyBinding": {
  "type":"object",
  "properties":{
    "crt":{
      "type":"string",
      "format":"base64",
      "example":"U3dhZ2dlciByb2Nrcw=="
    },
    "key":{
      "type":"string",
      "format":"base64",
      "example":"U3dhZ2dlciByb2Nrcw=="
    }
  }
}

```

### Utilizar anulações globais para suportar um tipo personalizado

Se estiver a utilizar ficheiros gerados, as etiquetas [`swaggertype`](#use-swaggertype-tag-to-supported-custom-type) ou `swaggerignore` podem não ser possíveis.

Ao passar um mapeamento para swag com `--overridesFile` pode dizer swag para utilizar um tipo no lugar de outro onde quer que apareça. Por defeito, se um ficheiro `.swaggo` estiver presente no directório actual, será utilizado.

Go code:
```go
type MyStruct struct {
  ID     sql.NullInt64 `json:"id"`
  Name   sql.NullString `json:"name"`
}
```

`.swaggo`:
```
// Substituir todos os NullInt64 por int
replace database/sql.NullInt64 int

// Não inclua quaisquer campos do tipo base de database/sql.
NullString no swagger docs
skip    database/sql.NullString
```

As directivas possíveis são comentários (começando por `//`), `replace path/to/a.type path/to/b.type`, e `skip path/to/a.type`.

(Note que os caminhos completos para qualquer tipo nomeado devem ser fornecidos para evitar problemas quando vários pacotes definem um tipo com o mesmo nome)

Entregue em:
```go
"types.MyStruct": {
  "id": "integer"
}

### Use swaggerignore tag para excluir um campo

```go
type Account struct {
    ID   string    `json:"id"`
    Name string     `json:"name"`
    Ignored int     `swaggerignore:"true"`
}
```


### Adicionar informações de extensão ao campo de estruturação

```go
type Account struct {
    ID   string    `json:"id"   extensions:"x-nullable,x-abc=def,!x-omitempty"` // extensions fields must start with "x-"
}
```

gerar doc. de swagger como se segue:

```go
"Account": {
    "type": "object",
    "properties": {
        "id": {
            "type": "string",
            "x-nullable": true,
            "x-abc": "def",
            "x-omitempty": false
        }
    }
}
```


### Renomear modelo a expor

```golang
type Resp struct {
	Code int
}//@name Response
```

### Como utilizar as anotações de segurança

Informações API gerais.

```go
// @securityDefinitions.basic BasicAuth

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information
```

Cada operação API.

```go
// @Security ApiKeyAuth
```

Faça-o OR condicione-o

```go
// @Security ApiKeyAuth
// @Security OAuth2Application[write, admin]
```

Faça-o AND condição

```go
// @Security ApiKeyAuth && firebase
// @Security OAuth2Application[write, admin] && APIKeyAuth
```



### Adicionar uma descrição para enumerar artigos

```go
type Example struct {
	// Sort order:
	// * asc - Ascending, from A to Z.
	// * desc - Descending, from Z to A.
	Order string `enums:"asc,desc"`
}
```

### Gerar apenas tipos de ficheiros de documentos específicos

Por defeito, o comando `swag` gera especificação Swagger em três tipos diferentes de ficheiros/arquivos:
- docs.go
- swagger.json
- swagger.yaml

Se desejar limitar um conjunto de tipos de ficheiros que devem ser gerados pode utilizar a bandeira `--outputTypes` (short `-ot`). O valor por defeito é `go,json,yaml` - tipos de saída separados por vírgula. Para limitar a saída apenas a ficheiros `go` e `yaml`, escrever-se-ia `go,yaml'. Com comando completo que seria `swag init --outputTypes go,yaml`.

### Como usar tipos genéricos

```go
// @Success 200 {object} web.GenericNestedResponse[types.Post]
// @Success 204 {object} web.GenericNestedResponse[types.Post, Types.AnotherOne]
// @Success 201 {object} web.GenericNestedResponse[web.GenericInnerType[types.Post]]
func GetPosts(w http.ResponseWriter, r *http.Request) {
	_ = web.GenericNestedResponse[types.Post]{}
}
```
Para mais detalhes e outros exemplos, veja [esse arquivo](https://github.com/swaggo/swag/blob/master/testdata/generics_nested/api/api.go)

### Alterar os delimitadores de acção padrão Go Template
[#980](https://github.com/swaggo/swag/issues/980)
[#1177](https://github.com/swaggo/swag/issues/1177)

Se as suas anotações ou campos estruturantes contêm "{{" or "}}", a geração de modelos irá muito provavelmente falhar, uma vez que estes são os delimitadores por defeito para [go templates](https://pkg.go.dev/text/template#Template.Delims).

Para que a geração funcione correctamente, pode alterar os delimitadores por defeito com `-td'. Por exemplo:
``console
swag init -g http/api.go -td "[[,]"
```

O novo delimitador é um fio com o formato "`<left delimiter>`,`<right delimiter>`".

## Sobre o projecto
Este projecto foi inspirado por [yvasiyarov/swagger](https://github.com/yvasiyarov/swagger) mas simplificámos a utilização e acrescentámos apoio a uma variedade de [frameworks web](#estruturas-web-suportadas). A fonte de imagem Gopher é [tenntenn/gopher-stickers](https://github.com/tenntenn/gopher-stickers). Tem licenças [creative commons licensing](http://creativecommons.org/licenses/by/3.0/deed.en).

## Contribuidores

Este projecto existe graças a todas as pessoas que contribuem. [[Contribute](CONTRIBUTING.md)].
<a href="https://github.com/swaggo/swag/graphs/contributors"><img src="https://opencollective.com/swag/contributors.svg?width=890&button=false" /></a>


## Apoios

Obrigado a todos os nossos apoiantes! 🙏 [[Become a backer](https://opencollective.com/swag#backer)]

<a href="https://opencollective.com/swag#backers" target="_blank"><img src="https://opencollective.com/swag/backers.svg?width=890"></a>


## Patrocinadores

Apoiar este projecto tornando-se um patrocinador. O seu logótipo aparecerá aqui com um link para o seu website. [[Become a sponsor](https://opencollective.com/swag#sponsor)]


<a href="https://opencollective.com/swag/sponsor/0/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/0/avatar.svg"></a>
<a href="https://opencollective.com/swag/sponsor/1/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/1/avatar.svg"></a>
<a href="https://opencollective.com/swag/sponsor/2/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/2/avatar.svg"></a>
<a href="https://opencollective.com/swag/sponsor/3/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/3/avatar.svg"></a>
<a href="https://opencollective.com/swag/sponsor/4/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/4/avatar.svg"></a>
<a href="https://opencollective.com/swag/sponsor/5/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/5/avatar.svg"></a>
<a href="https://opencollective.com/swag/sponsor/6/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/6/avatar.svg"></a>
<a href="https://opencollective.com/swag/sponsor/7/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/7/avatar.svg"></a>
<a href="https://opencollective.com/swag/sponsor/8/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/8/avatar.svg"></a>
<a href="https://opencollective.com/swag/sponsor/9/website" target="_blank"><img src="https://opencollective.com/swag/sponsor/9/avatar.svg"></a>


## Licença
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fswaggo%2Fswag.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fswaggo%2Fswag?ref=badge_large)
