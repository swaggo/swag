package web

import (
	"time"
)

type GenericBody[T any] struct {
	Data T
} // @name Body

type GenericBodyMulti[T any, X any] struct {
	Data T
	Meta X
} // @name MultiBody

type GenericResponse[T any] struct {
	Data T

	Status string
} // @name Response

type GenericResponseMulti[T any, X any] struct {
	Data T
	Meta X

	Status string
} // @name MultiResponse

type Post struct {
	ID int `json:"id" example:"1" format:"int64"`
	// Post name
	Name string `json:"name" example:"poti"`
	// Post data
	Data struct {
		// Post tag
		Tag []string `json:"name"`
	} `json:"data"`
} // @name Post

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
