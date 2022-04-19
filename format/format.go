package format

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/swaggo/swag"
)

// Format implements `fmt` command for formatting swag comments in Go source
// files.
type Format struct {
	formatter *swag.Formatter

	// exclude exclude dirs and files in SearchDir
	exclude map[string]bool
}

// New creates a new Format instance
func New() *Format {
	return &Format{
		exclude:   map[string]bool{},
		formatter: swag.NewFormatter(),
	}
}

// Config specifies configuration for a format run
type Config struct {
	// SearchDir the swag would be parse
	SearchDir string

	// excludes dirs and files in SearchDir,comma separated
	Excludes string

	// MainFile (DEPRECATED)
	MainFile string
}

var defaultExcludes = []string{"docs", "vendor"}

// Build runs formatter according to configuration in config
func (f *Format) Build(config *Config) error {
	searchDirs := strings.Split(config.SearchDir, ",")
	for _, searchDir := range searchDirs {
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return fmt.Errorf("fmt: %w", err)
		}
		for _, d := range defaultExcludes {
			f.exclude[filepath.Join(searchDir, d)] = true
		}
	}
	for _, fi := range strings.Split(config.Excludes, ",") {
		if fi = strings.TrimSpace(fi); fi != "" {
			f.exclude[filepath.Clean(fi)] = true
		}
	}
	for _, searchDir := range searchDirs {
		err := filepath.Walk(searchDir, f.visit)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Format) visit(path string, fileInfo os.FileInfo, err error) error {
	if fileInfo.IsDir() {
		return f.skipDir(path, fileInfo)
	}
	if f.exclude[path] ||
		strings.HasSuffix(strings.ToLower(path), "_test.go") ||
		filepath.Ext(path) != ".go" {
		return nil
	}
	if err := f.formatter.Format(path); err != nil {
		return fmt.Errorf("fmt: %w", err)
	}
	return nil
}

func (f *Format) skipDir(path string, info os.FileInfo) error {
	if f.exclude[path] ||
		len(info.Name()) > 1 && info.Name()[0] == '.' { // exclude hidden folders
		return filepath.SkipDir
	}
	return nil
}
