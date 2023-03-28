package main

// @Router /without-extension [get]
func Fun() {}

// @Router /with-another-extension [get]
// @x-another-extension {"address": "http://backend"}
func Fun2() {}

// @Router /with-correct-extension [get]
// @x-google-backend {"address": "http://backend"}
func Fun3() {}

// @Router /with-empty-comment-line [get]
func FunEmptyCommentLine() {}
