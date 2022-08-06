package types

type Field[T any] struct {
	Value T
}

type APIBase struct {
	APIUrl Field[string] `json:"@uri,omitempty"`
	ID     int           `json:"id" example:"1" format:"int64"`
}

type Post struct {
	APIBase
	// Post name
	Name string `json:"name" example:"poti"`
	// Post data
	Data struct {
		// Post tag
		Tag []string `json:"name"`
	} `json:"data"`
}
