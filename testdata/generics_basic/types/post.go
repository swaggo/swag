package types

type APIBase struct {
	APIUrl string `json:"@uri,omitempty"`
	ID     int    `json:"id" example:"1" format:"int64"`
}

type Post struct {
	APIBase
	ID int `json:"id" example:"1" format:"int64"`
	// Post name
	Name string `json:"name" example:"poti"`
	// Post data
	Data struct {
		// Post tag
		Tag []string `json:"name"`
	} `json:"data"`
}
