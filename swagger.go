package swag

import (
	"errors"
	"fmt"
	"sync"
)

// Name is a unique name be used to register swag instance.
const Name = "swagger"

var (
	swaggerMu sync.RWMutex
	swags     map[string]Swagger
)

// Swagger is a interface to read swagger document.
type Swagger interface {
	ReadDoc() string
}

// Register registers swagger for given name.
func Register(name string, swagger Swagger) {
	swaggerMu.Lock()
	defer swaggerMu.Unlock()

	if swagger == nil {
		panic("swagger is nil")
	}

	if swags == nil {
		swags = make(map[string]Swagger)
	}

	if _, ok := swags[name]; ok {
		panic("Register called twice for swag: " + name)
	}
	swags[name] = swagger
}

// ReadDoc reads swagger document.
func ReadDoc() (string, error) {
	return ReadDocName(Name)
}

// ReadDocName reads swagger document with specific name.
func ReadDocName(name string) (string, error) {
	if swags == nil {
		return "", errors.New("no swag has yet been registered")
	} else if swag, ok := swags[name]; !ok {
		return "", fmt.Errorf("no swag named \"%s\" was registered", name)
	} else {
		return swag.ReadDoc(), nil
	}
}
