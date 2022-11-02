package v1

type ProductDto struct {
	Name2 string `json:"name2"`
}

type ListResult[T any] struct {
	Items2 []T `json:"items2,omitempty"`
}

type RenamedProductDto struct {
	Name22 string `json:"name22"`
} // @name ProductDtoV2

type RenamedListResult[T any] struct {
	Items22 []T `json:"items22,omitempty"`
} // @name ListResultV2

type UniqueProduct struct {
	UniqueProductName string `json:"unique_product_name"`
}
