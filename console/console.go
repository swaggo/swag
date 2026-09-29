// Package console provides utilities for styled console output with ANSI colors and formatting.
// It offers two approaches: simple functions (Red, Bold, etc.) and a fluent builder API (ColorBuilder).
// Also supports template syntax via Sprintf: $Bold{text}, $Red{text}, etc. with nesting support.
package console

import (
	"fmt"
	"strings"
)

// Styles
const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	underline = "\033[4m"
)

// Colors
const (
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

// Emoji constants for console output.
const (
	Check   = "✅"  // Check indicates success or completion
	Fire    = "🔥"  // Fire indicates something hot or important
	X       = "❌"  // X indicates failure or error
	Info    = "ℹ️" // Info indicates informational message
	Warning = "⚠️" // Warning indicates a warning message
	Star    = "⭐"  // Star indicates something special or highlighted
)

// ColorBuilder provides a fluent interface for building styled console text.
// Use Format to create a builder, chain style methods (Bold, Red, etc.), then call String or Println.
// TODO: Needs Examples
type ColorBuilder struct {
	text  string
	codes []string
}

// Format creates a new ColorBuilder with formatted text using fmt.Sprintf syntax.
// Chain style methods like Bold(), Red(), etc. before calling String() or Println().
func Format(format string, messages ...any) *ColorBuilder {
	return &ColorBuilder{text: fmt.Sprintf(format, messages...)}
}

// Bold applies bold style to the text.
func (c *ColorBuilder) Bold() *ColorBuilder {
	c.codes = append(c.codes, bold)
	return c
}

// Underline applies underline style to the text.
func (c *ColorBuilder) Underline() *ColorBuilder {
	c.codes = append(c.codes, underline)
	return c
}

// Red applies red color to the text.
func (c *ColorBuilder) Red() *ColorBuilder {
	c.codes = append(c.codes, red)
	return c
}

// Green applies green color to the text.
func (c *ColorBuilder) Green() *ColorBuilder {
	c.codes = append(c.codes, green)
	return c
}

// Yellow applies yellow color to the text.
func (c *ColorBuilder) Yellow() *ColorBuilder {
	c.codes = append(c.codes, yellow)
	return c
}

// Blue applies blue color to the text.
func (c *ColorBuilder) Blue() *ColorBuilder {
	c.codes = append(c.codes, blue)
	return c
}

// Magenta applies magenta color to the text.
func (c *ColorBuilder) Magenta() *ColorBuilder {
	c.codes = append(c.codes, magenta)
	return c
}

// Cyan applies cyan color to the text.
func (c *ColorBuilder) Cyan() *ColorBuilder {
	c.codes = append(c.codes, cyan)
	return c
}

// White applies white color to the text.
func (c *ColorBuilder) White() *ColorBuilder {
	c.codes = append(c.codes, white)
	return c
}

// String returns the text with all applied ANSI escape codes.
func (c *ColorBuilder) String() string {
	return fmt.Sprintf("%s%s%s", join(c.codes), c.text, reset)
}

// Println prints the styled text to stdout with a newline.
func (c *ColorBuilder) Println() {
	fmt.Println(c.String())
}

// helper
func join(parts []string) string {
	out := ""
	for _, p := range parts {
		out += p
	}
	return out
}

// Printf formats and prints text with template syntax support.
// Uses Sprintf for formatting (see console.go:204).
func Printf(format string, args ...any) {
	fmt.Print(Sprintf(format, args...))
}

// parseTemplate processes template strings with $Keyword{content} syntax
// Supports: $Bold{}, $Red{}, $Green{}, $Yellow{}, $Blue{}, $Magenta{}, $Cyan{}, $White{}, $Underline{}
// Allows nesting: $Bold{$Red{text}}
func parseTemplate(format string) string {
	// Map of keywords to ANSI codes
	styleMap := map[string]string{
		"Bold":      bold,
		"Underline": underline,
		"Red":       red,
		"Green":     green,
		"Yellow":    yellow,
		"Blue":      blue,
		"Magenta":   magenta,
		"Cyan":      cyan,
		"White":     white,
	}

	return parseTemplateWithContext(format, styleMap, []string{})
}

// parseTemplateWithContext processes templates while maintaining a stack of active styles
func parseTemplateWithContext(format string, styleMap map[string]string, activeStyles []string) string {
	result := format
	changed := true

	for changed {
		changed = false

		// Look for the first occurrence of any template pattern
		firstStart := -1
		firstKeyword := ""
		firstPattern := ""

		for keyword := range styleMap {
			pattern := "$" + keyword + "{"
			start := strings.Index(result, pattern)
			if start != -1 && (firstStart == -1 || start < firstStart) {
				firstStart = start
				firstKeyword = keyword
				firstPattern = pattern
			}
		}

		if firstStart != -1 {
			code := styleMap[firstKeyword]

			// Find the matching closing brace
			braceCount := 0
			pos := firstStart + len(firstPattern)
			end := -1

			for i := pos; i < len(result); i++ {
				if result[i] == '{' {
					braceCount++
				} else if result[i] == '}' {
					if braceCount == 0 {
						end = i
						break
					}
					braceCount--
				}
			}

			if end != -1 {
				// Extract content between braces
				content := result[pos:end]

				// Process the content recursively with this style added to the stack
				newActiveStyles := append(activeStyles, code)
				processedContent := parseTemplateWithContext(content, styleMap, newActiveStyles)

				// Build the replacement
				replacement := code + processedContent + reset

				// If there are active parent styles, restore them
				if len(activeStyles) > 0 {
					for _, parentStyle := range activeStyles {
						replacement += parentStyle
					}
				}

				result = result[:firstStart] + replacement + result[end+1:]
				changed = true
			}
		}
	}

	return result
}

// Sprintf processes template strings with $Keyword{content} syntax and standard fmt formatting.
// Supports: $Bold{}, $Red{}, $Green{}, $Yellow{}, $Blue{}, $Magenta{}, $Cyan{}, $White{}, $Underline{}.
// Allows nesting (e.g., $Bold{$Red{%s}}) and standard fmt.Sprintf verbs.
// TODO: Needs Examples
func Sprintf(format string, args ...any) string {
	// First process the template syntax
	processed := parseTemplate(format)

	// Then apply standard fmt.Sprintf
	return fmt.Sprintf(processed, args...)
}

// Red returns text with red color ANSI codes.
func Red(text string) string {
	return fmt.Sprintf("%s%s%s", red, text, reset)
}

// Green returns text with green color ANSI codes.
func Green(text string) string {
	return fmt.Sprintf("%s%s%s", green, text, reset)
}

// Yellow returns text with yellow color ANSI codes.
func Yellow(text string) string {
	return fmt.Sprintf("%s%s%s", yellow, text, reset)
}

// Blue returns text with blue color ANSI codes.
func Blue(text string) string {
	return fmt.Sprintf("%s%s%s", blue, text, reset)
}

// Magenta returns text with magenta color ANSI codes.
func Magenta(text string) string {
	return fmt.Sprintf("%s%s%s", magenta, text, reset)
}

// Cyan returns text with cyan color ANSI codes.
func Cyan(text string) string {
	return fmt.Sprintf("%s%s%s", cyan, text, reset)
}

// White returns text with white color ANSI codes.
func White(text string) string {
	return fmt.Sprintf("%s%s%s", white, text, reset)
}

// Bold returns text with bold style ANSI codes.
func Bold(text string) string {
	return fmt.Sprintf("%s%s%s", bold, text, reset)
}

// Underline returns text with underline style ANSI codes.
func Underline(text string) string {
	return fmt.Sprintf("%s%s%s", underline, text, reset)
}
