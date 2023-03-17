package types

type SubField1[T any, T2 any] struct {
	SubValue1 T
	SubValue2 T2
}

type Field[T any] struct {
	Value  T
	Value2 *T
	Value3 []T
	Value4 SubField1[T, string]
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
