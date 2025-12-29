// Package main for enums_duplicate tests
// @title Enum Deduplication API
// @version 1.0
// @description Test API for enum deduplication and @swaggerignore
// @host localhost:8080
// @basePath /api/v1
// @schemes http
package main

import "github.com/swaggo/swag/testdata/enums_duplicate/types"

// GetLanguage retrieves a language
// @Summary Get a language
// @Accept json
// @Produce json
// @Param language query types.Language true "Language parameter"
// @Success 200 {string} types.Language
// @Router /language [get]
func GetLanguage(language types.Language) string {
	return string(language)
}

func main() {
}
