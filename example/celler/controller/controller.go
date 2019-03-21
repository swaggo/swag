package controller

// Controller example
type Controller struct {
}

// NewController example
func NewController() *Controller {
	return &Controller{}
}

// Message example
type Message struct {
	Message string `json:"message" example:"message"`
}
