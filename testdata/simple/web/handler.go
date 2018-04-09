package web

import (
	"time"
)

type Pet struct {
	ID       int `json:"id" example:"1"`
	Category struct {
		ID            int      `json:"id" example:"1"`
		Name          string   `json:"name" example:"category_name"`
		PhotoUrls     []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
		SmallCategory struct {
			ID        int      `json:"id" example:"1"`
			Name      string   `json:"name" example:"detail_category_name"`
			PhotoUrls []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
		} `json:"small_category"`
	} `json:"category"`
	Name      string      `json:"name" example:"poti"`
	PhotoUrls []string    `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
	Tags      []Tag       `json:"tags"`
	Status    string      `json:"status"`
	Price     float32     `json:"price" example:"3.25"`
	IsAlive   bool        `json:"is_alive" example:"true"`
	Data      interface{} `json:"data"`
	Hidden    string      `json:"-"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Pet2 struct {
	ID         int        `json:"id"`
	MiddleName *string    `json:"middlename"`
	DeletedAt  *time.Time `json:"deleted_at"`
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
