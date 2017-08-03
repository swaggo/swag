package swag

import (
	"errors"
	"sync"
)

const Name = "swagger"

var (
	swaggerMu sync.RWMutex
	swag      Swagger
)

type Swagger interface {
	ReadDoc() string
}

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

func ReadDoc() (string, error) {
	if swag != nil {
		return swag.ReadDoc(), nil
	}
	return "", errors.New("Not yet registered swag.")
}
