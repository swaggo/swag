package web

import "github.com/swaggo/swag/testdata/generics_property/types"

type PostSelector func(selector func())

type Filter interface {
	~func(selector func())
}

type query[T any, F Filter] interface {
	Where(ps ...F) T
}

type Pager[T query[T, F], F Filter] struct {
	Rows   uint8   `json:"rows" form:"rows"`
	Page   int     `json:"page" form:"page"`
	NextID *string `json:"next_id" form:"next_id"`
	PrevID *string `json:"prev_id" form:"prev_id"`
	query  T
}

type String string

func (String) Where(ps ...PostSelector) String {
	return ""
}

type PostPager struct {
	Pager[String, PostSelector]
	Search types.Field[string] `json:"search" form:"search"`
}

type PostResponse struct {
	GenericResponse[types.Post, types.Post]
}

type PostResponses struct {
	GenericResponse[[]types.Post, types.Post]
}

type StringResponse struct {
	GenericResponse[[]string, *uint8]
}

type GenericResponse[T any, T2 any] struct {
	Items  T
	Items2 T2
}
