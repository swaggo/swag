package main

import "github.com/griffnb/swag/testdata/alias_nested/pkg/good"

// @Success 200 {object} good.Gen
// @Router /api [get].
func main() {
	var _ good.Gen
}
