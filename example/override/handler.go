package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type ProductUpdates struct {
	Type        sql.NullString `json:"type"`
	Description sql.NullString `json:"description"`
	Stock       sql.NullInt64  `json:"stock"`
}

// UpdateProduct example
//
//	@Summary	Update product attributes
//	@ID			update-product
//	@Accept		json
//	@Param		product_id	path	int				true	"Product ID"
//	@Param		_			body	ProductUpdates	true	" "
//	@Router		/testapi/update-product/{product_id} [post]
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var pUpdates ProductUpdates
	if err := json.NewDecoder(r.Body).Decode(&pUpdates); err != nil {
		// write your code
		return
	}

	// write your code
}
