package direct

import "github.com/swaggo/swag/testdata/golist_parse_depth/transitive"

type Direct struct {
	Value string `json:"value"`
}

func init() {
	transitive.Touch()
}
