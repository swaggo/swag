package web

import (
	"time"
)

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

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Pet2 struct {
	ID int `json:"id"`
}

type APIError struct {
	ErrorCode    int
	ErrorMessage string
	CreatedAt    time.Time
}

type RevValueBase struct {
	Status bool `json:"Status"`

	Err int32 `json:"Err"`
}
type RevValue struct {
	RevValueBase

	Data int `json:"Data"`
}
