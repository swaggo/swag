package types

type Class int

const (
	A Class = iota + 1 // AAA
	B                  /* BBB */
	C
	D
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
	Teacher Type = "teacher" // teacher
	Student Type = "student" /* student */
	Other   Type = "Other"   // Other
	Unknown      = "Unknown"
	OtherUnknown
)

type Person struct {
	Name  string
	Class Class
	Mask  Mask
	Type  Type
}
