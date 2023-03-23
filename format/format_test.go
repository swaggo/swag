package format

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat_Format(t *testing.T) {
	fx := setup(t)
	assert.NoError(t, New().Build(&Config{SearchDir: fx.basedir}))
	assert.True(t, fx.isFormatted("main.go"))
	assert.True(t, fx.isFormatted("api/api.go"))
}

func TestFormat_ExcludeDir(t *testing.T) {
	fx := setup(t)
	assert.NoError(t, New().Build(&Config{
		SearchDir: fx.basedir,
		Excludes:  filepath.Join(fx.basedir, "api"),
	}))
	assert.False(t, fx.isFormatted("api/api.go"))
}

func TestFormat_ExcludeFile(t *testing.T) {
	fx := setup(t)
	assert.NoError(t, New().Build(&Config{
		SearchDir: fx.basedir,
		Excludes:  filepath.Join(fx.basedir, "main.go"),
	}))
	assert.False(t, fx.isFormatted("main.go"))
}

func TestFormat_DefaultExcludes(t *testing.T) {
	fx := setup(t)
	assert.NoError(t, New().Build(&Config{SearchDir: fx.basedir}))
	assert.False(t, fx.isFormatted("api/api_test.go"))
	assert.False(t, fx.isFormatted("docs/docs.go"))
}

func TestFormat_ParseError(t *testing.T) {
	fx := setup(t)
	os.WriteFile(filepath.Join(fx.basedir, "parse_error.go"), []byte(`package main
		func invalid() {`), 0644)
	assert.Error(t, New().Build(&Config{SearchDir: fx.basedir}))
}

func TestFormat_ReadError(t *testing.T) {
	fx := setup(t)
	os.Chmod(filepath.Join(fx.basedir, "main.go"), 0)
	assert.Error(t, New().Build(&Config{SearchDir: fx.basedir}))
}

func TestFormat_WriteError(t *testing.T) {
	fx := setup(t)
	os.Chmod(fx.basedir, 0555)
	assert.Error(t, New().Build(&Config{SearchDir: fx.basedir}))
	os.Chmod(fx.basedir, 0755)
}

func TestFormat_InvalidSearchDir(t *testing.T) {
	formatter := New()
	assert.Error(t, formatter.Build(&Config{SearchDir: "no_such_dir"}))
}

type fixture struct {
	t       *testing.T
	basedir string
}

func setup(t *testing.T) *fixture {
	fx := &fixture{
		t:       t,
		basedir: t.TempDir(),
	}
	for filename, contents := range testFiles {
		fullpath := filepath.Join(fx.basedir, filepath.Clean(filename))
		if err := os.MkdirAll(filepath.Dir(fullpath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullpath, contents, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return fx
}

func (fx *fixture) isFormatted(file string) bool {
	contents, err := os.ReadFile(filepath.Join(fx.basedir, filepath.Clean(file)))
	if err != nil {
		fx.t.Fatal(err)
	}
	return !bytes.Equal(testFiles[file], contents)
}

var testFiles = map[string][]byte{
	"api/api.go": []byte(`package api

		import "net/http"

		// @Summary Add a new pet to the store
		// @Description get string by ID
		func GetStringByInt(w http.ResponseWriter, r *http.Request) {
			//write your code
		}`),
	"api/api_test.go": []byte(`package api
		// @Summary API Test
		// @Description Should not be formatted
		func TestApi(t *testing.T) {}`),
	"docs/docs.go": []byte(`package docs
		// @Summary Documentation package
		// @Description Should not be formatted`),
	"main.go": []byte(`package main

		import (
			"net/http"

			"github.com/swaggo/swag/format/testdata/api"
		)

		// @title Swagger Example API
		// @version 1.0
		func main() {
			http.HandleFunc("/testapi/get-string-by-int/", api.GetStringByInt)
		}`),
	"README.md": []byte(`# Format test`),
}
