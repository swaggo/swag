package swagger

import (
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

	//fmt.Println(swaggers[Name].ReadDoc())
}

func ReadDoc() string {
	if val, ok := swaggers[Name]; ok {
		return val.ReadDoc()
	}
	return ""
}
