package swag

import (
	"reflect"
	"unicode"
	"unicode/utf8"
)

// FieldsFunc split a string s by a func splitter into max n parts
func FieldsFunc(s string, f func(rune2 rune) bool, n int) []string {
	// A span is used to record a slice of s of the form s[start:end].
	// The start index is inclusive and the end index is exclusive.
	type span struct {
		start int
		end   int
	}
	spans := make([]span, 0, 32)

	// Find the field start and end indices.
	// Doing this in a separate pass (rather than slicing the string s
	// and collecting the result substrings right away) is significantly
	// more efficient, possibly due to cache effects.
	start := -1 // valid span start if >= 0
	for end, rune := range s {
		if f(rune) {
			if start >= 0 {
				spans = append(spans, span{start, end})
				// Set start to a negative value.
				// Note: using -1 here consistently and reproducibly
				// slows down this code by a several percent on amd64.
				start = ^start
			}
		} else {
			if start < 0 {
				start = end
				if n > 0 && len(spans)+1 >= n {
					break
				}
			}
		}
	}

	// Last field might end at EOF.
	if start >= 0 {
		spans = append(spans, span{start, len(s)})
	}

	// Create strings from recorded field indices.
	a := make([]string, len(spans))
	for i, span := range spans {
		a[i] = s[span.start:span.end]
	}
	return a
}

// FieldsByAnySpace split a string s by any space character into max n parts
func FieldsByAnySpace(s string, n int) []string {
	return FieldsFunc(s, unicode.IsSpace, n)
}

// AppendUtf8Rune appends the UTF-8 encoding of r to the end of p and
// returns the extended buffer. If the rune is out of range,
// it appends the encoding of RuneError.
func AppendUtf8Rune(p []byte, r rune) []byte {
	return utf8.AppendRune(p, r)
}

// CanIntegerValue a wrapper of reflect.Value
type CanIntegerValue struct {
	reflect.Value
}

// CanInt reports whether Uint can be used without panicking.
func (v CanIntegerValue) CanInt() bool {
	return v.Value.CanInt()
}

// CanUint reports whether Uint can be used without panicking.
func (v CanIntegerValue) CanUint() bool {
	return v.Value.CanUint()
}
