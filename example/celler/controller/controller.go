package controller

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

type Message struct {
	Message string `json:"message" example:"message"`
}
