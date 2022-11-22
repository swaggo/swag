package types

import (
	"github.com/swaggo/swag/testdata/enums/consts"
)

type Class int

const (
	None Class = -1
	A    Class = consts.Base + (iota+1-1)*2/2%100 - (1&1 | 1) + (2 ^ 2) // AAA
	B                                                                   /* BBB */
	C
	D
	F = D + 1
	//G is not enum
	G = H + 10
	//H is not enum
	H = 10
	//I is not enum
	I = int(F + 2)
)

type Mask int

const (
	Mask1 Mask = 2 << iota >> 1 // Mask1
	Mask2                       /* Mask2 */
	Mask3                       // Mask3
	Mask4                       // Mask4
)

type Type string

const (
	Teacher      Type = "teacher" // teacher
	Student      Type = "student" /* student */
	Other        Type = "Other"   // Other
	Unknown           = "Unknown"
	OtherUnknown      = string(Other + Unknown)
)

type Person struct {
	Name  string
	Class Class
	Mask  Mask
	Type  Type
}
