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

// ErrDocIsNotRegistered is returned when a swag document is not registered.
var ErrDocIsNotRegistered = errors.New("no swag has yet been registered")

// Swagger is an interface to read swagger document.
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

// Unregister removes a registered swagger instance by name.
func Unregister(name string) {
	swaggerMu.Lock()
	defer swaggerMu.Unlock()

	if swags == nil {
		return
	}

	delete(swags, name)
}

// Clear removes all registered swagger instances.
func Clear() {
	swaggerMu.Lock()
	defer swaggerMu.Unlock()

	swags = nil
}

// GetSwagger returns the swagger instance for given name.
// If not found, returns nil.
func GetSwagger(name string) Swagger {
	swaggerMu.RLock()
	defer swaggerMu.RUnlock()

	return swags[name]
}

// ReadDoc reads swagger document. An optional name parameter can be passed to read a specific document.
// The default name is "swagger".
func ReadDoc(optionalName ...string) (string, error) {
	swaggerMu.RLock()
	defer swaggerMu.RUnlock()

	if swags == nil {
		return "", ErrDocIsNotRegistered
	}

	name := Name
	if len(optionalName) != 0 && optionalName[0] != "" {
		name = optionalName[0]
	}

	swag, ok := swags[name]
	if !ok {
		return "", fmt.Errorf("no swag named \"%s\" was registered: %w", name, ErrDocIsNotRegistered)
	}

	return swag.ReadDoc(), nil
}
