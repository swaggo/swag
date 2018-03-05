package swag

import (
	"errors"
	"sync"
)

// Name TODO: NEEDS COMMENT INFO
const Name = "swagger"

//  TODO: NEEDS COMMENT INFO
var (
	swaggerMu sync.RWMutex
	swag      Swagger
)

// Swagger TODO: NEEDS COMMENT INFO
type Swagger interface {
	ReadDoc() string
}

// Register TODO: NEEDS COMMENT INFO
func Register(name string, swagger Swagger) {
	swaggerMu.Lock()
	defer swaggerMu.Unlock()
	if swagger == nil {
		panic("swagger is nil")
	}

	if swag != nil {
		panic("Register called twice for swag: " + name)
	}
	swag = swagger
}

// ReadDoc TODO: NEEDS COMMENT INFO
func ReadDoc() (string, error) {
	if swag != nil {
		return swag.ReadDoc(), nil
	}
	return "", errors.New("not yet registered swag")
}
