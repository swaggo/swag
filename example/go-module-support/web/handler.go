package web

import (
	"time"
)

// Pet example
type Pet struct {
	ID       int `json:"id"`
	Category struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Name      string   `json:"name"`
	PhotoUrls []string `json:"photoUrls"`
	Tags      []Tag    `json:"tags"`
	Status    string   `json:"status"`
}

// Tag example
type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Pet2 example
type Pet2 struct {
	ID int `json:"id"`
}

// APIError example
type APIError struct {
	ErrorCode    int
	ErrorMessage string
	CreatedAt    time.Time
}

// RevValueBase example
type RevValueBase struct {
	Status bool `json:"Status"`

	Err int32 `json:"Err"`
}

// RevValue example
type RevValue struct {
	RevValueBase

	Data int `json:"Data"`
}
