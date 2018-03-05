package gen

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGen_Build(t *testing.T) {
	searchDir := "../example/simple"
	assert.NotPanics(t, func() {
		New().Build(searchDir, "./main.go", "../example/simple/docs/swagger")
	})

	if _, err := os.Stat(path.Join(searchDir, "docs", "docs.go")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../example/simple/docs/swagger", "swagger.json")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../example/simple/docs/swagger", "swagger.yaml")); os.IsNotExist(err) {
		t.Fatal(err)
	}
}
