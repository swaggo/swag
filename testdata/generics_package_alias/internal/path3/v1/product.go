package v1

type ProductDto struct {
	Name3 string `json:"name3"`
}

type ListResult[T any] struct {
	Items3 []T `json:"items3,omitempty"`
}

type RenamedProductDto struct {
	Name33 string `json:"name33"`
} // @name ProductDtoV3

type RenamedListResult[T any] struct {
	Items33 []T `json:"items33,omitempty"`
} // @name ListResultV3
