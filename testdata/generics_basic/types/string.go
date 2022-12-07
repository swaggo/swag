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

type Hello struct {
	MyStringField1    Field[*string]                `json:"myStringField1"`
	MyStringField2    Field[string]                 `json:"myStringField2"`
	MyArrayField      DoubleField[*string, string]  `json:"myNewField"`
	MyArrayDepthField TrippleField[*string, string] `json:"myNewArrayField"`
}
