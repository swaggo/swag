//go:build !go1.21
// +build !go1.21

package swag

import "go/ast"

// ast.IsGenerated is only supported for version of Go >= 1.21.
// Hence, for older Go version we always return false.
// See https://go.dev/doc/go1.21#goastpkggoast
func astFileIsGenerated(*ast.File) bool {
	return false
}
