package account

import "github.com/swaggo/swag/testdata/core_models/constants"

type Authentication struct {
	// Raw External ID of the app
	Keys map[constants.ExternalID]*OAuthData `json:"keys"`
}

type OAuthData struct {
	AccessToken  string               `json:"access_token"`
	CompanyId    string               `json:"companyId"`
	LocationId   string               `json:"locationId"`
	ExpiresIn    int64                `json:"expires_in"`
	RefreshToken string               `json:"refresh_token"`
	Scope        string               `json:"scope"`
	TokenType    string               `json:"token_type"`
	UserId       string               `json:"userId"`
	UserType     string               `json:"userType"`
	ExpireTS     int64                `json:"expire_ts"`
	ProductID    int64                `json:"product_id"`
	ExtID        constants.ExternalID `json:"ext_id"`
}
