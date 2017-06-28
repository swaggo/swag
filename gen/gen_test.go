package gen

import (
	"testing"
)

func TestGen_Build(t *testing.T) {
	searchDir := "/Users/easonlin/gocode/src/api"
	New().Build(searchDir, "./main.go")
}
