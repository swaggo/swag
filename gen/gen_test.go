package gen

import (
	"github.com/go-openapi/spec"
	"os"
	"os/exec"
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

func TestGen_formatSource(t *testing.T) {
	src := `package main

import "net

func main() {}
`
	g := New()

	res := g.formatSource([]byte(src))
	assert.Equal(t, []byte(src), res, "Should return same content due to fmt fail")

	src2 := `package main

import "fmt"

func main() {
fmt.Print("Helo world")
}
`
	res = g.formatSource([]byte(src2))
	assert.NotEqual(t, []byte(src2), res, "Should return fmt code")
}

type mocWriter struct{}

func (w *mocWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func TestGen_writeGoDoc(t *testing.T) {
	gen := New()

	swapTemplate := packageTemplate

	packageTemplate = `{{{`
	err := gen.writeGoDoc(nil, nil)
	assert.Error(t, err)

	packageTemplate = `{{.Data}}`
	swagger := &spec.Swagger{
		VendorExtensible: spec.VendorExtensible{},
		SwaggerProps: spec.SwaggerProps{
			Info: &spec.Info{},
		},
	}
	err = gen.writeGoDoc(&mocWriter{}, swagger)
	assert.Error(t, err)

	packageTemplate = swapTemplate

}

func TestGen_GeneratedDoc(t *testing.T) {

	searchDir := "../testdata/simple"

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		PropNamingStrategy: "",
	}

	assert.NoError(t, New().Build(config))
	gocmd, err := exec.LookPath("go")
	assert.NoError(t, err)

	cmd := exec.Command(gocmd, "build", filepath.Join(config.OutputDir, "docs.go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	assert.NoError(t, cmd.Run())
}
