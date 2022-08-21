package web

import (
	"time"
)

// GenericNestedBody[T]
// @Description Some Generic Body
type GenericNestedBody[T any] struct {
	// Items from the list response
	Items T
	// Status of some other stuff
	Status string
}

// GenericInnerType[T]
// @Description Some Generic Body
type GenericInnerType[T any] struct {
	// Items from the list response
	Items T
}

// GenericInnerMultiType[T, X]
// @Description Some Generic Body
type GenericInnerMultiType[T any, X any] struct {
	// ItemsOne is the first thing
	ItemOne T
	// ItemsTwo is the second thing
	ItemsTwo []X
}

// GenericNestedResponse[T]
// @Description Some Generic List Response
type GenericNestedResponse[T any] struct {
	// Items from the list response
	Items []T
	// Status of some other stuff
	Status string
}

// GenericNestedResponseMulti[T, X]
// @Description this contains a few things
type GenericNestedResponseMulti[T any, X any] struct {
	// ItemsOne is the first thing
	ItemOne T
	// ItemsTwo is the second thing
	ItemsTwo []X

	// Status of the things
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
