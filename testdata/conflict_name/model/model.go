package model

type MyStruct struct {
	Name string `json:"name"`
}

type MyPayload struct {
	My   MyStruct
	Name string `json:"name"`
}

type ErrorsResponse struct {
	NewTime MyPayload
}
