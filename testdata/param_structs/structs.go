package structs

type FormModel struct {
	Foo string `form:"f" binding:"required" validate:"max=10"`
	// B is another field
	B bool
}

type AuthHeader struct {
	// Token is the auth token
	Token string `header:"X-Auth-Token" binding:"required"`
	// AnotherHeader is another header
	AnotherHeader int `validate:"gte=0,lte=10"`
}

type PathModel struct {
	// ID is the id
	Identifier int    `uri:"id" binding:"required"`
	Name       string `validate:"max=10"`
}
