//go:build !go1.21
// +build !go1.21

package swag

import (
	"go/ast"
	"strings"
)

// ast.IsGenerated is only supported for version of Go >= 1.21.
// Hence, for older Go version we always return false.
// See https://go.dev/doc/go1.21#goastpkggoast
func astFileIsGenerated(file *ast.File) bool {
	_, ok := generator(file)
	return ok
}

// Copied directly from https://cs.opensource.google/go/go/+/refs/tags/go1.25.0:src/go/ast/ast.go;l=1112-1137
// to make it available in Go versions < 1.21.
func generator(file *ast.File) (string, bool) {
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if comment.Pos() > file.Package {
				break // after package declaration
			}
			// opt: check Contains first to avoid unnecessary array allocation in Split.
			const prefix = "// Code generated "
			if strings.Contains(comment.Text, prefix) {
				for _, line := range strings.Split(comment.Text, "\n") {
					if rest, ok := strings.CutPrefix(line, prefix); ok {
						if gen, ok := strings.CutSuffix(rest, " DO NOT EDIT."); ok {
							return gen, true
						}
					}
				}
			}
		}
	}
	return "", false
}
