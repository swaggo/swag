package response

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
