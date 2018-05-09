package gen

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGen_Build(t *testing.T) {
	searchDir := "../testdata/simple"
	assert.NotPanics(t, func() {
		New().Build(searchDir, "./main.go", "../testdata/simple/docs/swagger", "snakecase")
	})

	if _, err := os.Stat(path.Join(searchDir, "docs", "docs.go")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple/docs/swagger", "swagger.json")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple/docs/swagger", "swagger.yaml")); os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestGen_BuildSnakecase(t *testing.T) {
	searchDir := "../testdata/simple2"
	assert.NotPanics(t, func() {
		New().Build(searchDir, "./main.go", "../testdata/simple2/docs/swagger", "snakecase")
	})

	if _, err := os.Stat(path.Join(searchDir, "docs", "docs.go")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple2/docs/swagger", "swagger.json")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple2/docs/swagger", "swagger.yaml")); os.IsNotExist(err) {
		t.Fatal(err)
	}
}
