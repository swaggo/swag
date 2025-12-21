package constants

type GlobalConfigKey string

const (
	GCK_ALLOW_SELF_SIGNED_CERTS GlobalConfigKey = "allow_self_signed_certs" // Allow self-signed SSL certificates
	GCK_API_RATE_LIMIT_ENABLED  GlobalConfigKey = "api_rate_limit_enabled"  // Enable API rate limiting
)
