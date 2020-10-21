package model

type MyStruct struct {
	Name string `json:"name"`
}

type MyPayload2 struct {
	My   MyStruct
	Name string `json:"name"`
}

type ErrorsResponse struct {
	NewTime MyPayload2
}
