package data

import (
	typesapplication "github.com/nguyennm96/swag/v2/testdata/alias_import/types"
)

type ApplicationResponse struct {
	typesapplication.TypeToEmbed

	Application      typesapplication.Application   `json:"application"`
	ApplicationArray []typesapplication.Application `json:"application_array"`
	ApplicationTime  typesapplication.DateOnly      `json:"application_time"`
}
