package data

import (
	typesapplication "github.com/swaggo/swag/testdata/global_override/types"
)

type ApplicationResponse struct {
	typesapplication.TypeToEmbed

	Application      typesapplication.Application   `json:"application"`
	Application2     typesapplication.Application2  `json:"application2"`
	ApplicationArray []typesapplication.Application `json:"application_array"`
	ApplicationTime  typesapplication.DateOnly      `json:"application_time"`
	ShouldSkip       typesapplication.ShouldSkip    `json:"should_skip"`
}
