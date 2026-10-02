package shared

type Outer struct {
	Items []*Inner `json:"items"`
}

type Inner struct {
	ID int `json:"id"`
}
