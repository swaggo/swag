package consts

import "github.com/swaggo/swag/testdata/enums/types"

const Base = 1

const uintSize = 32 << (^uint(uintptr(0)) >> 63)
const maxBase = 10 + ('z' - 'a' + 1) + ('Z' - 'A' + 1)
const shlByLen = 1 << len("aaa")
const hexnum = 0xFF
const octnum = 017
const nonescapestr = `aa\nbb\u8888cc`
const escapestr = "aa\nbb\u8888cc"
const escapechar = '\u8888'
const underscored = 1_000_000
const binaryInteger = 0b10001000

const lenArrayLit = len([3]int{})

type arrayAlias = [2][3]int

const lenArrayAlias1 = len(arrayAlias{})
const lenArrayAlias2 = len(arrayAlias{}[0])

const lenArrayFieldsOfAnonymousStruct = len(struct {
	Items [3]int
}{}.Items)
const lenArrayFieldsOfNamedStruct = len([2]types.MyStruct{}[0].A.Items)
