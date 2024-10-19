package consts

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
