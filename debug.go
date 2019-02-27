package swag

import (
	"log"
)

const (
	test = iota
	release
)

var swagMode = release

func isRelease() bool {
	return swagMode == release
}
func Println(v ...interface{}) {
	if isRelease() {
		log.Println(v...)
	}
}

func Printf(format string, v ...interface{}) {
	if isRelease() {
		log.Printf(format, v...)
	}
}
