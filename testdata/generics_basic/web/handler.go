package web

import (
	"time"
)

type GenericResponse[T any] struct {
	Data T

	Status string
}

type GenericResponseMulti[T any, X any] struct {
	Data T
	Meta X

	Status string
}

type Post struct {
	ID int `json:"id" example:"1" format:"int64"`
	// Post name
	Name string `json:"name" example:"poti"`
	// Post data
	Data struct {
		// Post tag
		Tag []string `json:"name"`
	} `json:"data"`
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
