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
	"github.com/swaggo/swag"
)

const searchDir = "../testdata/simple"

func TestGen_Build(t *testing.T) {
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
			require.NoError(t, err)
		}
		_ = os.Remove(expectedFile)
	}
}

func TestGen_BuildInstanceName(t *testing.T) {
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
		require.NoError(t, err)
	}
	if !strings.Contains(string(expectedCode), "swag.Register(\"swagger\", &s{})") {
		t.Fatal(errors.New("generated go code does not contain the correct default registration sequence"))
	}

	// Custom name
	config.InstanceName = "custom"
	assert.NoError(t, New().Build(config))
	expectedCode, err = ioutil.ReadFile(goSourceFile)
	if err != nil {
		require.NoError(t, err)
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
			require.NoError(t, err)
		}
		_ = os.Remove(expectedFile)
	}
}

func TestGen_BuildSnakeCase(t *testing.T) {
	config := &Config{
		SearchDir:          "../testdata/simple2",
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple2/docs",
		PropNamingStrategy: swag.SnakeCase,
	}

	assert.NoError(t, New().Build(config))

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			require.NoError(t, err)
		}
		_ = os.Remove(expectedFile)
	}
}

func TestGen_BuildLowerCamelcase(t *testing.T) {
	config := &Config{
		SearchDir:          "../testdata/simple3",
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
			require.NoError(t, err)
		}
		_ = os.Remove(expectedFile)
	}
}

func TestGen_BuildDescriptionWithQuotes(t *testing.T) {
	config := &Config{
		SearchDir:        "../testdata/quotes",
		MainAPIFile:      "./main.go",
		OutputDir:        "../testdata/quotes/docs",
		MarkdownFilesDir: "../testdata/quotes",
	}

	require.NoError(t, New().Build(config))

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			require.NoError(t, err)
		}
	}
	cmd := exec.Command("go", "build", "-buildmode=plugin", "github.com/swaggo/swag/testdata/quotes")
	cmd.Dir = config.SearchDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		require.NoError(t, err, string(output))
	}
	p, err := plugin.Open(filepath.Join(config.SearchDir, "quotes.so"))
	if err != nil {
		require.NoError(t, err)
	}
	defer os.Remove("quotes.so")

	readDoc, err := p.Lookup("ReadDoc")
	if err != nil {
		require.NoError(t, err)
	}
	jsonOutput := readDoc.(func() string)()
	var jsonDoc interface{}
	if err := json.Unmarshal([]byte(jsonOutput), &jsonDoc); err != nil {
		require.NoError(t, err)
	}
	expectedJSON, err := ioutil.ReadFile(filepath.Join(config.SearchDir, "expected.json"))
	if err != nil {
		require.NoError(t, err)
	}
	assert.JSONEq(t, string(expectedJSON), jsonOutput)
}

func TestGen_jsonIndent(t *testing.T) {
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
			require.NoError(t, err)
		}
		_ = os.Remove(expectedFile)
	}
}

func TestGen_SearchDirIsNotExist(t *testing.T) {
	var swaggerConfDir, propNamingStrategy string
	config := &Config{
		SearchDir:          "../isNotExistDir",
		MainAPIFile:        "./main.go",
		OutputDir:          swaggerConfDir,
		PropNamingStrategy: propNamingStrategy,
	}
	assert.EqualError(t, New().Build(config), "dir: ../isNotExistDir does not exist")
}

func TestGen_MainAPiNotExist(t *testing.T) {
	var swaggerConfDir, propNamingStrategy string
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./notExists.go",
		OutputDir:          swaggerConfDir,
		PropNamingStrategy: propNamingStrategy,
	}
	assert.Error(t, New().Build(config))
}

func TestGen_OutputIsNotExist(t *testing.T) {
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
		require.NoError(t, err)
	}

	_ = os.RemoveAll(filepath.Join(outputDir, "swagger.yaml"))
	err = os.Mkdir(filepath.Join(outputDir, "swagger.yaml"), 0755)
	if err != nil {
		require.NoError(t, err)
	}
	assert.Error(t, New().Build(config))

	_ = os.RemoveAll(filepath.Join(outputDir, "swagger.json"))
	err = os.Mkdir(filepath.Join(outputDir, "swagger.json"), 0755)
	if err != nil {
		require.NoError(t, err)
	}
	assert.Error(t, New().Build(config))

	_ = os.RemoveAll(filepath.Join(outputDir, "docs.go"))

	err = os.Mkdir(filepath.Join(outputDir, "docs.go"), 0755)
	if err != nil {
		require.NoError(t, err)
	}
	assert.Error(t, New().Build(config))

	err = os.RemoveAll(outputDir)
	if err != nil {
		require.NoError(t, err)
	}
}

func TestGen_configWithOutputDir(t *testing.T) {
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
			require.NoError(t, err)
		}
		_ = os.Remove(expectedFile)
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
fmt.Print("Hello world")
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
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		PropNamingStrategy: "",
	}

	assert.NoError(t, New().Build(config))
	goCMD, err := exec.LookPath("go")
	assert.NoError(t, err)

	cmd := exec.Command(goCMD, "build", filepath.Join(config.OutputDir, "docs.go"))
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
			require.NoError(t, err)
		}
		_ = os.Remove(expectedFile)
	}
}

func TestGen_cgoImports(t *testing.T) {
	config := &Config{
		SearchDir:          "../testdata/simple_cgo",
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
			require.NoError(t, err)
		}
		_ = os.Remove(expectedFile)
	}
}

func TestGen_duplicateRoute(t *testing.T) {
	config := &Config{
		SearchDir:          "../testdata/duplicate_route",
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
