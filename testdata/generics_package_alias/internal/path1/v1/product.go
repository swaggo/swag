package v1

type ProductDto struct {
	Name1 string `json:"name1"`
}

type ListResult[T any] struct {
	Items1 []T `json:"items1,omitempty"`
}

type RenamedProductDto struct {
	Name11 string `json:"name11"`
} // @name ProductDtoV1

type RenamedListResult[T any] struct {
	Items11 []T `json:"items11,omitempty"`
} // @name ListResultV1
