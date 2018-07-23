package web

import (
	"time"

	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/swaggo/swag/testdata/simple/cross"
)

type Pet struct {
	ID       int `json:"id" example:"1" format:"int64"`
	Category struct {
		ID            int      `json:"id" example:"1"`
		Name          string   `json:"name" example:"category_name"`
		PhotoUrls     []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg" format:"url"`
		SmallCategory struct {
			ID        int      `json:"id" example:"1"`
			Name      string   `json:"name" example:"detail_category_name" binding:"required"`
			PhotoUrls []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
		} `json:"small_category"`
	} `json:"category"`
	Name      string          `json:"name" example:"poti"`
	PhotoUrls []string        `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg" binding:"required"`
	Tags      []Tag           `json:"tags"`
	Pets      *[]Pet2         `json:"pets"`
	Pets2     []*Pet2         `json:"pets2"`
	Status    string          `json:"status"`
	Price     float32         `json:"price" example:"3.25"`
	IsAlive   bool            `json:"is_alive" example:"true"`
	Data      interface{}     `json:"data"`
	Hidden    string          `json:"-"`
	UUID      uuid.UUID       `json:"uuid"`
	Decimal   decimal.Decimal `json:"decimal"`
}

type Tag struct {
	ID   int    `json:"id" format:"int64"`
	Name string `json:"name"`
	Pets []Pet  `json:"pets"`
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

	Err int32 `json:"Err,omitempty"`
}
type RevValue struct {
	RevValueBase `json:"rev_value_base"`

	Data    int           `json:"Data"`
	Cross   cross.Cross   `json:"cross"`
	Crosses []cross.Cross `json:"crosses"`
}
