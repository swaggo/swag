package global_security

// @title Swagger Example API
// @version 1.0

// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name Authorization

// @securityDefinitions.basic  BasicAuth

// @securityDefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @security APIKeyAuth || OAuth2Application
func main() {}
