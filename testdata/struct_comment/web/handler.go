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
	// Error an Api error
	Error     string
	CreatedAt time.Time
}
