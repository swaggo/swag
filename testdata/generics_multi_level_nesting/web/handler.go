package web

import (
	"encoding/json"
)

type TestResponse struct {
	Field1 Entity[int64]   `json:"field_1"`
	Field2 Entity[float64] `json:"field_2"`
}

type Entity[T int64 | float64] struct {
	LineWithFixType     EmptyArray[DataPoint[float64]] `json:"line_with_fix_type"`
	LineWithGenericType EmptyArray[DataPoint[T]]       `json:"line_with_generic_type"`
	MultipleLines       MultipleLines[T]               `json:"multiple_lines"`
}

type DataPoint[T int64 | float64] struct {
	Value     T     `json:"value"`
	Timestamp int64 `json:"timestamp"`
}

// EmptyArray will show [] instead of nil.
type EmptyArray[T any] []T

func (arr EmptyArray[T]) MarshalJSON() ([]byte, error) {
	if arr != nil {
		return json.Marshal([]T(arr))
	}

	return json.Marshal(make([]T, 0))
}

type MultipleLines[T int64 | float64] []NamedLineData[T]

type NamedLineData[T int64 | float64] struct {
	Name string                   `json:"name"`
	Data EmptyArray[DataPoint[T]] `json:"data"`
}
