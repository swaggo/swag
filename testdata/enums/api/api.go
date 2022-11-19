package api

import "github.com/swaggo/swag/testdata/enums/types"

//	 enum example
//
//		@Summary      enums
//		@Description  enums
//		@Failure      400   {object}  types.Person  "ok"
//		@Router       /students [post]
func API() {
	_ = types.Person{}
}
