package structs

type FormModel struct {
	Foo string `form:"f" binding:"required" validate:"max=10"`
	// B is another field
	B string
}

type AuthHeader struct {
	// Token is the auth token
	Token string `header:"X-Auth-Token" binding:"required"`
	// AnotherHeader is another header
	AnotherHeader int `validate:"gte=0,lte=10"`
}
