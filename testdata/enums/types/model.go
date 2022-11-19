package types

type Class int

const Base = 1

const (
	None Class = -1
	A    Class = Base + (iota+1-1)*2/2 - 1 // AAA
	B                                      /* BBB */
	C
	D
	F = D + 1
	//G is not enum
	G = 10
	//H is not enum
	H = int(F + 2)
)

type Mask int

const (
	Mask1 Mask = 1 << iota // Mask1
	Mask2                  /* Mask2 */
	Mask3                  // Mask3
	Mask4                  // Mask4
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
