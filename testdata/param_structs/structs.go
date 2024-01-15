package structs

type AuthHeader struct {
	// Token is the auth token
	Token string `header:"X-Auth-Token" binding:"required"`
	// AnotherHeader is another header
	AnotherHeader int `validate:"gte=0,lte=10"`
}
