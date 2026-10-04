package types

// Language represents supported languages
type Language string

const (
	// English language
	En Language = "EN"
	// German language
	De Language = "DE"
	// Chinese language
	Zh Language = "ZH"
	// DefaultLanguage is an alias for English (should be excluded via deduplication)
	DefaultLanguage = En
	// @swaggerignore
	Fr Language = "FR"
)
