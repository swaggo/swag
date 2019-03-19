package gen

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGen_Build(t *testing.T) {
	searchDir := "../testdata/simple"

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		PropNamingStrategy: "",
	}
	assert.NoError(t, New().Build(config))

	if _, err := os.Stat(path.Join("../testdata/simple/docs", "docs.go")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple/docs", "swagger.json")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple/docs", "swagger.yaml")); os.IsNotExist(err) {
		t.Fatal(err)
	}

	//TODO: remove gen files
}

func TestGen_BuildSnakecase(t *testing.T) {
	searchDir := "../testdata/simple2"
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple2/docs",
		PropNamingStrategy: "snakecase",
	}

	assert.NoError(t, New().Build(config))

	if _, err := os.Stat(path.Join("../testdata/simple2/docs", "docs.go")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple2/docs", "swagger.json")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple2/docs", "swagger.yaml")); os.IsNotExist(err) {
		t.Fatal(err)
	}

	//TODO: remove gen files
}

func TestGen_BuildLowerCamelcase(t *testing.T) {
	searchDir := "../testdata/simple3"
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple3/docs",
		PropNamingStrategy: "",
	}

	assert.NoError(t, New().Build(config))

	if _, err := os.Stat(path.Join("../testdata/simple3/docs", "docs.go")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple3/docs", "swagger.json")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple3/docs", "swagger.yaml")); os.IsNotExist(err) {
		t.Fatal(err)
	}

	//TODO: remove gen files
}

func TestGen_SearchDirIsNotExist(t *testing.T) {
	searchDir := "../isNotExistDir"

	var swaggerConfDir, propNamingStrategy string
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          swaggerConfDir,
		PropNamingStrategy: propNamingStrategy,
	}
	assert.EqualError(t, New().Build(config), "dir: ../isNotExistDir is not exist")
}

func TestGen_configWithOutputDir(t *testing.T) {
	searchDir := "../testdata/simple"

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		PropNamingStrategy: "",
	}

	assert.NoError(t, New().Build(config))

	if _, err := os.Stat(path.Join("../testdata/simple/docs", "docs.go")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple/docs", "swagger.json")); os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := os.Stat(path.Join("../testdata/simple/docs", "swagger.yaml")); os.IsNotExist(err) {
		t.Fatal(err)
	}

	//TODO: remove gen files
}
