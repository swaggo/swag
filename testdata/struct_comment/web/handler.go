package web

import (
	"time"
)

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

type APIError struct {
	Error     string    // Error an Api error
	CreatedAt time.Time // Error time
}
