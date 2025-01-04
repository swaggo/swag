package types

type GenericResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type GenericMultiResponse[T any, X any] struct {
	Status string `json:"status"`
	DataT  T      `json:"data_t"`
	DataX  X      `json:"data_x"`
}
