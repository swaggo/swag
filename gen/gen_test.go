package gen

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-openapi/spec"
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

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
		os.Remove(expectedFile)
	}
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

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
		os.Remove(expectedFile)
	}
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

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
		os.Remove(expectedFile)
	}
}

func TestGen_jsonIndent(t *testing.T) {
	searchDir := "../testdata/simple"

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		PropNamingStrategy: "",
	}
	gen := New()
	gen.jsonIndent = func(data interface{}) ([]byte, error) {
		return nil, errors.New("fail")
	}
	assert.Error(t, gen.Build(config))
}

func TestGen_jsonToYAML(t *testing.T) {
	searchDir := "../testdata/simple"

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		PropNamingStrategy: "",
	}
	gen := New()
	gen.jsonToYAML = func(data []byte) ([]byte, error) {
		return nil, errors.New("fail")
	}
	assert.Error(t, gen.Build(config))

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
		os.Remove(expectedFile)
	}
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

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
		os.Remove(expectedFile)
	}
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

type mockWriter struct {
	hook func([]byte)
}

func (w *mockWriter) Write(data []byte) (int, error) {
	if w.hook != nil {
		w.hook(data)
	}
	return len(data), nil
}

func TestGen_writeGoDoc(t *testing.T) {
	gen := New()

	swapTemplate := packageTemplate

	packageTemplate = `{{{`
	err := gen.writeGoDoc("docs", nil, nil, &Config{})
	assert.Error(t, err)

	packageTemplate = `{{.Data}}`
	swagger := &spec.Swagger{
		VendorExtensible: spec.VendorExtensible{},
		SwaggerProps: spec.SwaggerProps{
			Info: &spec.Info{},
		},
	}
	err = gen.writeGoDoc("docs", &mockWriter{}, swagger, &Config{})
	assert.Error(t, err)

	packageTemplate = `{{ if .GeneratedTime }}Fake Time{{ end }}`
	err = gen.writeGoDoc("docs",
		&mockWriter{
			hook: func(data []byte) {
				assert.Equal(t, "Fake Time", string(data))
			},
		}, swagger, &Config{GeneratedTime: true})
	assert.NoError(t, err)
	err = gen.writeGoDoc("docs",
		&mockWriter{
			hook: func(data []byte) {
				assert.Equal(t, "", string(data))
			},
		}, swagger, &Config{GeneratedTime: false})
	assert.NoError(t, err)

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

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
		os.Remove(expectedFile)
	}
}

func TestGen_cgoImports(t *testing.T) {
	searchDir := "../testdata/simple_cgo"

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple_cgo/docs",
		PropNamingStrategy: "",
		ParseDependency:    true,
	}

	assert.NoError(t, New().Build(config))

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
		os.Remove(expectedFile)
	}
}

func TestGen_multipleDefs(t *testing.T) {
	searchDir := "../testdata/multiple_def"

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/multiple_def/docs",
		PropNamingStrategy: "",
		ParseDependency:    true,
		Categories:         "bar",
	}

	assert.NoError(t, New().Build(config))

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
		os.Remove(expectedFile)
	}
}
