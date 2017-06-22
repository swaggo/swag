package gen

import (
	"testing"
)

func TestGen_Build(t *testing.T) {
	searchDir := "../example"
	New().Build(searchDir, "./main.go")
}
