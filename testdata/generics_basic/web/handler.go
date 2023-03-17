package web

import (
	"time"
)

type GenericBody[T any] struct {
	Data T
}

type GenericBodyMulti[T any, X any] struct {
	Data T
	Meta X
}

type GenericResponse[T any] struct {
	Data T

	Status string
}

type GenericResponseMulti[T any, X any] struct {
	Data T
	Meta X

	Status string
}

// APIError
// @Description API error
// @Description with information about it
// Other some summary
type APIError struct {
	// Error an Api error
	Error string // Error this is Line comment
	// Error `number` tick comment
	ErrorNo   int64
	ErrorCtx  string    // Error `context` tick comment
	CreatedAt time.Time // Error time
}
