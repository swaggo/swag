package main

// @title Grouped Security Definitions
// @version 1.0
// @securityDefinitions.bearerauth BearerAuth
// @description Bearer access token.
// @bearerformat JWT
// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name X-API-Key
// @description API key auth.
// @securityDefinitions.apikey SessionCookieAuth
// @in cookie
// @name session_cookie
// @description Session cookie auth.
func main() {}
