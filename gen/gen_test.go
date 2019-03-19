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
		config := &Config{
			SearchDir:          searchDir,
			MainAPIFile:        "./main.go",
			SwaggerConfDir:     "../testdata/simple/docs/swagger",
			PropNamingStrategy: "",
		}
		New().Build(config)
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
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		SwaggerConfDir:     "../testdata/simple2/docs/swagger",
		PropNamingStrategy: "snakecase",
	}

	assert.NoError(t, New().Build(config))

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

func TestGen_BuildLowerCamelcase(t *testing.T) {
	searchDir := "../testdata/simple3"
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		SwaggerConfDir:     "../testdata/simple3/docs/swagger",
		PropNamingStrategy: "",
	}

	assert.NoError(t, New().Build(config))

	if _, err := os.Stat(path.Join(searchDir, "docs", "docs.go")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple3/docs/swagger", "swagger.json")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple3/docs/swagger", "swagger.yaml")); os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestGen_SearchDirIsNotExist(t *testing.T) {
	searchDir := "../isNotExistDir"

	var swaggerConfDir, propNamingStrategy string
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		SwaggerConfDir:     swaggerConfDir,
		PropNamingStrategy: propNamingStrategy,
	}
	assert.EqualError(t, New().Build(config), "dir: ../isNotExistDir is not exist")
}
