package web

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type Pet struct {
	ID       int `example:"1" format:"int64"`
	Category struct {
		ID            int      `example:"1"`
		Name          string   `example:"category_name"`
		PhotoURLs     []string `example:"http://test/image/1.jpg,http://test/image/2.jpg" format:"url"`
		SmallCategory struct {
			ID        int      `example:"1"`
			Name      string   `example:"detail_category_name"`
			PhotoURLs []string `example:"http://test/image/1.jpg,http://test/image/2.jpg"`
		}
	}
	Name      string   `example:"poti"`
	PhotoURLs []string `example:"http://test/image/1.jpg,http://test/image/2.jpg"`
	Tags      []Tag
	Pets      *[]Pet2
	Pets2     []*Pet2
	Status    string
	Price     float32 `example:"3.25"`
	IsAlive   bool    `example:"true"`
	Data      interface{}
	Hidden    string `json:"-"`
	UUID      uuid.UUID
	Decimal   decimal.Decimal
}

type Tag struct {
	ID   int `format:"int64"`
	Name string
	Pets []Pet
}

type Pet2 struct {
	ID         int
	MiddleName *string
	DeletedAt  *time.Time
}

type APIError struct {
	ErrorCode    int
	ErrorMessage string
	CreatedAt    time.Time
}

type RevValueBase struct {
	Status bool

	Err int32
}
type RevValue struct {
	RevValueBase

	Data int
}
