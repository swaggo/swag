package web

// @hello
// yo yo yo yo
type Pet struct {
	ID       int `json:"id"`
	Category struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Name      string   `json:"name"`
	PhotoUrls []string `json:"photoUrls"`
	Tags      []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
	Status string `json:"status"`
}

type Pet2 struct {
	ID int `json:"id"`
}

type APIError struct {
	ErrorCode    int
	ErrorMessage string
}
