package gen

import (
	"os"
	"path"
	"path/filepath"
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

func TestGen_MainAPiNotExist(t *testing.T) {
	searchDir := "../testdata/simple"

	var swaggerConfDir, propNamingStrategy string
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./notexists.go",
		OutputDir:          swaggerConfDir,
		PropNamingStrategy: propNamingStrategy,
	}
	assert.Error(t, New().Build(config))
}

func TestGen_OutputIsNotExist(t *testing.T) {
	searchDir := "../testdata/simple"

	var propNamingStrategy string
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "/dev/null",
		PropNamingStrategy: propNamingStrategy,
	}
	assert.Error(t, New().Build(config))
}

func TestGen_FailToWrite(t *testing.T) {
	searchDir := "../testdata/simple"

	outputDir := filepath.Join(os.TempDir(), "swagg", "test")

	var propNamingStrategy string
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          outputDir,
		PropNamingStrategy: propNamingStrategy,
	}

	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll(filepath.Join(outputDir, "swagger.yaml"))
	err = os.Mkdir(filepath.Join(outputDir, "swagger.yaml"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	assert.Error(t, New().Build(config))

	os.RemoveAll(filepath.Join(outputDir, "swagger.json"))
	err = os.Mkdir(filepath.Join(outputDir, "swagger.json"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	assert.Error(t, New().Build(config))

	os.RemoveAll(filepath.Join(outputDir, "docs.go"))

	err = os.Mkdir(filepath.Join(outputDir, "docs.go"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	assert.Error(t, New().Build(config))

	err = os.RemoveAll(outputDir)
	if err != nil {
		t.Fatal(err)
	}
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
