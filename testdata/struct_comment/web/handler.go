package web

import (
	"time"
)

type Post struct {
	ID int `json:"id" example:"1" format:"int64"`
	// Name post name
	Name string `json:"name" example:"poti"`
	// Data post data
	Data interface{} `json:"data"`
}

type APIError struct {
	// Error an Api error
	Error     string
	CreatedAt time.Time
}
