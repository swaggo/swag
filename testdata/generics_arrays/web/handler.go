package web

import (
	"time"
)

// GenericListResponse[T]
// @Description Some Generic List Response
type GenericListResponse[T any] struct {
	// Items from the list response
	Items []T
	// Status of some other stuff
	Status string
}

// GenericListResponseMulti[T, X]
// @Description this contains a few things
type GenericListResponseMulti[T any, X any] struct {
	// ItemsOne is the first thing
	ItemsOne []T
	// ItemsTwo is the second thing
	ItemsTwo []X

	// Status of the things
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
