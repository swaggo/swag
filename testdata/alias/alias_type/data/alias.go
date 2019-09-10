package data

import (
	"time"

	"github.com/swaggo/swag/testdata/alias/alias_type/types"
)

type TimeContainer struct {
	Name      types.StringAlias `json:"name"`
	Timestamp time.Time         `json:"timestamp"`
	CreatedAt types.DateOnly    `json:"created_at"`
}
