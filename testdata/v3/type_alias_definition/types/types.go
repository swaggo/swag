package types

import "github.com/swaggo/swag/v2/testdata/v3/type_alias_definition/othertypes"

type Struct struct {
	String       string       `json:"string"`
	NestedStruct NestedStruct `json:"nestedStruct"`
}

type NestedStruct struct {
	Int int `json:"int"`
}

type StructAlias = Struct
type OtherStructAlias = othertypes.Struct

type StructSubtype Struct
type OtherStructSubtype othertypes.Struct

type Response struct {
	Struct             `json:"struct"`
	StructAlias        `json:"structAlias"`
	OtherStructAlias   `json:"otherStructAlias"`
	StructSubtype      `json:"structSubtype"`
	OtherStructSubtype `json:"otherStructSubtype"`
}
