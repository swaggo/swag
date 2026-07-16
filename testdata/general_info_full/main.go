package main

// @title           Some API Title
// @version         0.0.1
// @description     This is a sample server for a pet store.
// @termsOfService  http://swagger.io/terms/
//
// @host                     localhost:8080
// @BasePath                 /api/v1
// @accept                   json
// @produce                  json
// @schemes                  http https
// @query.collection.format  multi
//
// @securityDefinitions.basic  BasicAuth
// @x-basic-foo                "some basic value"
//
// @securityDefinitions.apikey  Bearer
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and JWT token.
// @x-bearer-foo                "some bearer value"
//
// @securitydefinitions.oauth2.application  OAuth2Application
// @description                             OAuth protects our entity endpoints
// @tokenUrl                                https://example.com/oauth/token
// @scope.write                             Grants write access
// @scope.admin                             Grants read and write access to administrative information
// @x-oa2-application-foo                   "some application value"
//
// @securitydefinitions.oauth2.implicit  OAuth2Implicit
// @authorizationurl                     https://example.com/oauth/authorize
// @scope.write                          Grants write access
// @scope.admin                          Grants read and write access to administrative information
// @x-oa2-implicit-foo                   "some implicit value"
//
// @securitydefinitions.oauth2.password  OAuth2Password
// @tokenUrl                             https://example.com/oauth/token
// @scope.write                          Grants write access
// @scope.admin                          Grants read and write access to administrative information
// @x-oa2-password-foo                   "some password value"
//
// @securitydefinitions.oauth2.accessCode  OAuth2AccessCode
// @tokenUrl                               https://example.com/oauth/token
// @authorizationurl                       https://example.com/oauth/authorize
// @scope.write                            Grants write access
// @scope.admin                            Grants read and write access to administrative information
// @x-oa2-accessCode-foo                   "some accessCode value"
//
// @tag.name         dogs
// @tag.description  Dogs are cool
//
// @tag.name apes
// @tag.description.markdown
//
// @tag.name              cats
// @tag.description       Cats are the devil
// @tag.docs.url          https://example.com/cats
// @tag.docs.description  Everything about cats
//
// @externalDocs.description  Some description here
// @externalDocs.url          http://some.url.com/path/here/
//
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
//
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
//
// @x-foo "some value"
func main() {}
