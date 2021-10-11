package gen

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGen_BuildRegistrationName(t *testing.T) {
	searchDir := "../testdata/simple"

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		PropNamingStrategy: "",
	}
	assert.NoError(t, New().Build(config))

	goSourceFile := filepath.Join(config.OutputDir, "docs.go")

	// Validate default registration name
	expectedCode, err := ioutil.ReadFile(goSourceFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(expectedCode), "swag.Register(swag.Name, &s{})") {
		t.Fatal(errors.New("generated go code does not contain the correct default registration sequence"))
	}

	// Custom name
	config.RegistrationName = "custom"
	assert.NoError(t, New().Build(config))
	expectedCode, err = ioutil.ReadFile(goSourceFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(expectedCode), "swag.Register(\"custom\", &s{})") {
		t.Fatal(errors.New("generated go code does not contain the correct registration sequence"))
	}

	// cleanup
	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
		_ = os.Remove(expectedFile)
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

func TestGen_BuildDescriptionWithQuotes(t *testing.T) {
	searchDir := "../testdata/quotes"
	config := &Config{
		SearchDir:        searchDir,
		MainAPIFile:      "./main.go",
		OutputDir:        "../testdata/quotes/docs",
		MarkdownFilesDir: searchDir,
	}

	require.NoError(t, New().Build(config))

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "build", "-buildmode=plugin", "github.com/swaggo/swag/testdata/quotes")
	cmd.Dir = config.SearchDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(output))
	}
	p, err := plugin.Open(filepath.Join(config.SearchDir, "quotes.so"))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("quotes.so")

	readDoc, err := p.Lookup("ReadDoc")
	if err != nil {
		t.Fatal(err)
	}
	jsonOutput := []byte(readDoc.(func() string)())
	var jsonDoc interface{}
	if err := json.Unmarshal(jsonOutput, &jsonDoc); err != nil {
		t.Fatal(err, string(jsonOutput))
	}
	expectedJSON, err := ioutil.ReadFile(filepath.Join(config.SearchDir, "expected.json"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedJSON, jsonOutput)
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
	assert.EqualError(t, New().Build(config), "dir: ../isNotExistDir does not exist")
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

func TestGen_duplicateRoute(t *testing.T) {
	searchDir := "../testdata/duplicate_route"

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/duplicate_route/docs",
		PropNamingStrategy: "",
		ParseDependency:    true,
	}
	err := New().Build(config)
	assert.NoError(t, err)

	// with Strict enabled should cause an error instead of warning about the duplicate route
	config.Strict = true
	err = New().Build(config)
	assert.EqualError(t, err, "route GET /testapi/endpoint is declared multiple times")
}
