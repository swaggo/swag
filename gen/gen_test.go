package gen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
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

var outputTypes = []string{"go", "json", "yaml"}

func TestGen_Build(t *testing.T) {
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		OutputTypes:        outputTypes,
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

func TestGen_SpecificOutputTypes(t *testing.T) {
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		OutputTypes:        []string{"go", "unknownType"},
		PropNamingStrategy: "",
	}
	assert.NoError(t, New().Build(config))

	tt := []struct {
		expectedFile string
		shouldExist  bool
	}{
		{filepath.Join(config.OutputDir, "docs.go"), true},
		{filepath.Join(config.OutputDir, "swagger.json"), false},
		{filepath.Join(config.OutputDir, "swagger.yaml"), false},
	}
	for _, tc := range tt {
		_, err := os.Stat(tc.expectedFile)
		if tc.shouldExist {
			if os.IsNotExist(err) {
				require.NoError(t, err)
			}
		} else {
			require.Error(t, err)
			require.True(t, errors.Is(err, os.ErrNotExist))
		}

		_ = os.Remove(tc.expectedFile)
	}
}

func TestGen_BuildInstanceName(t *testing.T) {
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		OutputTypes:        outputTypes,
		PropNamingStrategy: "",
	}
	assert.NoError(t, New().Build(config))

	goSourceFile := filepath.Join(config.OutputDir, "docs.go")

	// Validate default registration name
	expectedCode, err := os.ReadFile(goSourceFile)
	if err != nil {
		require.NoError(t, err)
	}

	if !strings.Contains(
		string(expectedCode),
		"swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)",
	) {
		t.Fatal(errors.New("generated go code does not contain the correct default registration sequence"))
	}

	if !strings.Contains(
		string(expectedCode),
		"var SwaggerInfo =",
	) {
		t.Fatal(errors.New("generated go code does not contain the correct default variable declaration"))
	}

	// Custom name
	config.InstanceName = "Custom"
	goSourceFile = filepath.Join(config.OutputDir, config.InstanceName+"_"+"docs.go")
	assert.NoError(t, New().Build(config))

	expectedCode, err = os.ReadFile(goSourceFile)
	if err != nil {
		require.NoError(t, err)
	}

	if !strings.Contains(
		string(expectedCode),
		"swag.Register(SwaggerInfoCustom.InstanceName(), SwaggerInfoCustom)",
	) {
		t.Fatal(errors.New("generated go code does not contain the correct registration sequence"))
	}

	if !strings.Contains(
		string(expectedCode),
		"var SwaggerInfoCustom =",
	) {
		t.Fatal(errors.New("generated go code does not contain the correct variable declaration"))
	}

	// cleanup
	expectedFiles := []string{
		filepath.Join(config.OutputDir, config.InstanceName+"_"+"docs.go"),
		filepath.Join(config.OutputDir, config.InstanceName+"_"+"swagger.json"),
		filepath.Join(config.OutputDir, config.InstanceName+"_"+"swagger.yaml"),
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
		OutputTypes:        outputTypes,
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
		OutputTypes:        outputTypes,
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
		OutputTypes:      outputTypes,
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

	expectedJSON, err := os.ReadFile(filepath.Join(config.SearchDir, "expected.json"))
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
		OutputTypes:        outputTypes,
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
		OutputTypes:        outputTypes,
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
		OutputTypes:        outputTypes,
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
		OutputTypes:        outputTypes,
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
		OutputTypes:        outputTypes,
		PropNamingStrategy: propNamingStrategy,
	}
	assert.Error(t, New().Build(config))
}

func TestGen_FailToWrite(t *testing.T) {
	outputDir := filepath.Join(os.TempDir(), "swagg", "test")
	outputTypes := []string{"go", "json", "yaml"}

	var propNamingStrategy string
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          outputDir,
		OutputTypes:        outputTypes,
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
		OutputTypes:        outputTypes,
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

func TestGen_configWithOutputTypesAll(t *testing.T) {
	searchDir := "../testdata/simple"
	outputTypes := []string{"go", "json", "yaml"}

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		OutputTypes:        outputTypes,
		PropNamingStrategy: "",
	}

	assert.NoError(t, New().Build(config))

	expectedFiles := []string{
		path.Join(config.OutputDir, "docs.go"),
		path.Join(config.OutputDir, "swagger.json"),
		path.Join(config.OutputDir, "swagger.yaml"),
	}
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Fatal(err)
		}

		_ = os.Remove(expectedFile)
	}
}

func TestGen_configWithOutputTypesSingle(t *testing.T) {
	searchDir := "../testdata/simple"
	outputTypes := []string{"go", "json", "yaml"}

	for _, outputType := range outputTypes {
		config := &Config{
			SearchDir:          searchDir,
			MainAPIFile:        "./main.go",
			OutputDir:          "../testdata/simple/docs",
			OutputTypes:        []string{outputType},
			PropNamingStrategy: "",
		}

		assert.NoError(t, New().Build(config))

		outFileName := "swagger"
		if outputType == "go" {
			outFileName = "docs"
		}

		expectedFiles := []string{
			path.Join(config.OutputDir, outFileName+"."+outputType),
		}
		for _, expectedFile := range expectedFiles {
			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				t.Fatal(err)
			}

			_ = os.Remove(expectedFile)
		}
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
		OutputTypes:        outputTypes,
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
		OutputTypes:        outputTypes,
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

func TestGen_parseOverrides(t *testing.T) {
	testCases := []struct {
		Name          string
		Data          string
		Expected      map[string]string
		ExpectedError error
	}{
		{
			Name: "replace",
			Data: `replace github.com/foo/bar baz`,
			Expected: map[string]string{
				"github.com/foo/bar": "baz",
			},
		},
		{
			Name: "skip",
			Data: `skip github.com/foo/bar`,
			Expected: map[string]string{
				"github.com/foo/bar": "",
			},
		},
		{
			Name: "generic-simple",
			Data: `replace types.Field[string] string`,
			Expected: map[string]string{
				"types.Field[string]": "string",
			},
		},
		{
			Name: "generic-double",
			Data: `replace types.Field[string,string] string`,
			Expected: map[string]string{
				"types.Field[string,string]": "string",
			},
		},
		{
			Name: "comment",
			Data: `// this is a comment
			replace foo bar`,
			Expected: map[string]string{
				"foo": "bar",
			},
		},
		{
			Name: "ignore whitespace",
			Data: `

			replace foo bar`,
			Expected: map[string]string{
				"foo": "bar",
			},
		},
		{
			Name:          "unknown directive",
			Data:          `foo`,
			ExpectedError: fmt.Errorf("could not parse override: 'foo'"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			overrides, err := parseOverrides(strings.NewReader(tc.Data))
			assert.Equal(t, tc.Expected, overrides)
			assert.Equal(t, tc.ExpectedError, err)
		})
	}
}

func TestGen_TypeOverridesFile(t *testing.T) {
	customPath := "/foo/bar/baz"

	tmp, err := os.CreateTemp("", "")
	require.NoError(t, err)

	defer os.Remove(tmp.Name())

	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		PropNamingStrategy: "",
	}

	t.Run("Default file is missing", func(t *testing.T) {
		open = func(path string) (*os.File, error) {
			assert.Equal(t, DefaultOverridesFile, path)

			return nil, os.ErrNotExist
		}
		defer func() {
			open = os.Open
		}()

		config.OverridesFile = DefaultOverridesFile
		err := New().Build(config)
		assert.NoError(t, err)
	})

	t.Run("Default file is present", func(t *testing.T) {
		open = func(path string) (*os.File, error) {
			assert.Equal(t, DefaultOverridesFile, path)

			return tmp, nil
		}
		defer func() {
			open = os.Open
		}()

		config.OverridesFile = DefaultOverridesFile
		err := New().Build(config)
		assert.NoError(t, err)
	})

	t.Run("Different file is missing", func(t *testing.T) {
		open = func(path string) (*os.File, error) {
			assert.Equal(t, customPath, path)

			return nil, os.ErrNotExist
		}
		defer func() {
			open = os.Open
		}()

		config.OverridesFile = customPath
		err := New().Build(config)
		assert.EqualError(t, err, "could not open overrides file: file does not exist")
	})

	t.Run("Different file is present", func(t *testing.T) {
		open = func(path string) (*os.File, error) {
			assert.Equal(t, customPath, path)

			return tmp, nil
		}
		defer func() {
			open = os.Open
		}()

		config.OverridesFile = customPath
		err := New().Build(config)
		assert.NoError(t, err)
	})
}
func TestGen_Debugger(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		SearchDir:          searchDir,
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/simple/docs",
		OutputTypes:        outputTypes,
		PropNamingStrategy: "",
		Debugger:           log.New(&buf, "", log.LstdFlags),
	}
	assert.True(t, buf.Len() == 0)
	assert.NoError(t, New().Build(config))
	assert.True(t, buf.Len() > 0)

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

func TestGen_ErrorAndInterface(t *testing.T) {
	config := &Config{
		SearchDir:          "../testdata/error",
		MainAPIFile:        "./main.go",
		OutputDir:          "../testdata/error/docs",
		OutputTypes:        outputTypes,
		PropNamingStrategy: "",
	}

	assert.NoError(t, New().Build(config))

	expectedFiles := []string{
		filepath.Join(config.OutputDir, "docs.go"),
		filepath.Join(config.OutputDir, "swagger.json"),
		filepath.Join(config.OutputDir, "swagger.yaml"),
	}
	t.Cleanup(func() {
		for _, expectedFile := range expectedFiles {
			_ = os.Remove(expectedFile)
		}
	})

	// check files
	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			require.NoError(t, err)
		}
	}

	// check content
	jsonOutput, err := os.ReadFile(filepath.Join(config.OutputDir, "swagger.json"))
	if err != nil {
		require.NoError(t, err)
	}
	expectedJSON, err := os.ReadFile(filepath.Join(config.SearchDir, "expected.json"))
	if err != nil {
		require.NoError(t, err)
	}

	assert.JSONEq(t, string(expectedJSON), string(jsonOutput))
}
