package swag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGlobalEnums(t *testing.T) {
	searchDir := "testdata/enums"
	expected, err := os.ReadFile(filepath.Join(searchDir, "expected.json"))
	assert.NoError(t, err)

	p := New()
	err = p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(p.swagger, "", "    ")
	assert.NoError(t, err)
	fmt.Println(string(b))
	assert.Equal(t, string(expected), string(b))
}
