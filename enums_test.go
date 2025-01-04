package swag

import (
	"encoding/json"
	"math/bits"
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
	assert.Equal(t, string(expected), string(b))
	constsPath := "github.com/swaggo/swag/testdata/enums/consts"
	assert.Equal(t, bits.UintSize, p.packages.packages[constsPath].ConstTable["uintSize"].Value)
	assert.Equal(t, int32(62), p.packages.packages[constsPath].ConstTable["maxBase"].Value)
	assert.Equal(t, 8, p.packages.packages[constsPath].ConstTable["shlByLen"].Value)
	assert.Equal(t, 255, p.packages.packages[constsPath].ConstTable["hexnum"].Value)
	assert.Equal(t, 15, p.packages.packages[constsPath].ConstTable["octnum"].Value)
	assert.Equal(t, `aa\nbb\u8888cc`, p.packages.packages[constsPath].ConstTable["nonescapestr"].Value)
	assert.Equal(t, "aa\nbb\u8888cc", p.packages.packages[constsPath].ConstTable["escapestr"].Value)
	assert.Equal(t, 1_000_000, p.packages.packages[constsPath].ConstTable["underscored"].Value)
	assert.Equal(t, 0b10001000, p.packages.packages[constsPath].ConstTable["binaryInteger"].Value)
}
