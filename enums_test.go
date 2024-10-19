package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGlobalEnums(t *testing.T) {
	searchDir := "testdata/enums"

	p := New()
	err := p.ParseAPI(searchDir, mainAPIFile, defaultParseDepth)
	require.NoError(t, err)

	const constsPath = "github.com/swaggo/swag/v2/testdata/enums/consts"
	table := p.packages.packages[constsPath].ConstTable
	require.NotNil(t, table, "const table must not be nil")

	assert.Equal(t, 64, table["uintSize"].Value)
	assert.Equal(t, int32(62), table["maxBase"].Value)
	assert.Equal(t, 8, table["shlByLen"].Value)
	assert.Equal(t, 255, table["hexnum"].Value)
	assert.Equal(t, 15, table["octnum"].Value)
	assert.Equal(t, `aa\nbb\u8888cc`, table["nonescapestr"].Value)
	assert.Equal(t, "aa\nbb\u8888cc", table["escapestr"].Value)
	assert.Equal(t, '\u8888', table["escapechar"].Value)
}
