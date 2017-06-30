package swagger

import (
	"errors"
	"sync"
)

const Name = "swagger"

var (
	swaggerMu sync.RWMutex
	swaggers  = make(map[string]Swagger)
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

	if _, dup := swaggers[name]; dup {
		panic("Register called twice for swag doc: " + name)
	}
	swaggers[name] = swagger
}

func ReadDoc() (string, error) {
	if val, ok := swaggers[Name]; ok {
		return val.ReadDoc(), nil
	}

	return "", errors.New("Can't found swag doc")

}
