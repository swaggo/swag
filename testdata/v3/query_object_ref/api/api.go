package api

// Include is a named string type with enum constants.
type Include string

const (
	IncludeAuthor   Include = "author"
	IncludeComments Include = "comments"
)

// Nested is a named struct type.
type Nested struct {
	A string `json:"a"`
}

// Query is bound from the query string.
type Query struct {
	Includes []Include `json:"include"`
	Include  Include   `json:"single_include"`
	Items    []Nested  `json:"items"`
	Item     Nested    `json:"item"`
	Limit    int       `json:"limit"`
}

// ListPosts lists posts.
// @Summary List posts
// @Param query query Query true "query"
// @Success 200 {string} string
// @Router /posts [get]
func ListPosts() {}
