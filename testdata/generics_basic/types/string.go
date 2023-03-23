package types

type Field[T any] struct {
	Value T
}

type DoubleField[T1 any, T2 any] struct {
	Value1 T1
	Value2 T2
}

type TrippleField[T1 any, T2 any] struct {
	Value1 T1
	Value2 T2
}

type ArrayField[T any] []T
type MapField[K comparable, V any] map[K]V
type MapFieldValue struct {
	S string
	F float64
}
type MapFieldNestedStruct[K comparable] map[K]MapFieldValue

type Hello struct {
	MyStringField1       Field[*string]                `json:"myStringField1"`
	MyStringField2       Field[string]                 `json:"myStringField2"`
	MyArrayField         DoubleField[*string, string]  `json:"myNewField"`
	MyArrayDepthField    TrippleField[*string, string] `json:"myNewArrayField"`
	ArrayField           ArrayField[string]            `json:"arrayField"`
	MapField             MapField[string, float64]     `json:"mapField"`
	OriginArrayField     []string                      `json:"originArrayField"`
	OriginMapField       map[string]float64            `json:"originMapField"`
	MapFieldNestedStruct MapFieldNestedStruct[string]  `json:"mapFieldNestedStruct"`
}
