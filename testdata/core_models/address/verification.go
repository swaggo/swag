package address

type Verification struct {
	Street                         *FieldVerification `public:"view" json:"street"`
	Street2                        *FieldVerification `public:"view" json:"street_2"`
	City                           *FieldVerification `public:"view" json:"city"`
	State                          *FieldVerification `public:"view" json:"state"`
	PostalCode                     *FieldVerification `public:"view" json:"postal_code"`
	GoogleFormattedAddress         string             `public:"view" json:"google_formatted_address"`
	GoogleStreetNumber             *FieldVerification `public:"view" json:"google_street_number"`
	GoogleRoute                    *FieldVerification `public:"view" json:"google_route"`
	GoogleSubpremise               *FieldVerification `public:"view" json:"google_subpremise"`
	GoogleNeighborhood             *FieldVerification `public:"view" json:"google_neighborhood"`
	GoogleLocality                 *FieldVerification `public:"view" json:"google_locality"`
	GoogleSublocality              *FieldVerification `public:"view" json:"google_sublocality"`
	GoogleAdministrativeAreaLevel1 *FieldVerification `public:"view" json:"google_administrative_area_level_1"`
	GoogleAdministrativeAreaLevel2 *FieldVerification `public:"view" json:"google_administrative_area_level_2"`
	GooglePostalCode               *FieldVerification `public:"view" json:"google_postal_code"`
	GooglePostalCodeSuffix         *FieldVerification `public:"view" json:"google_postal_code_suffix"`
	GoogleCountry                  *FieldVerification `public:"view" json:"google_country"`
	IsMatch                        int64              `public:"view" json:"is_match"`
	IsUnconfirmed                  int64              `public:"view" json:"is_unconfirmed"`
	ExternalID                     string             `public:"view" json:"external_id"` // PlaceID / AWS ID
}

type FieldVerification struct {
	Value             string `public:"view" json:"value"`
	GoogleValue       string `public:"view" json:"google_value"`
	ConfirmationLevel string `public:"view" json:"confirmation_level"`
	IsDifferent       int64  `public:"view" json:"is_different"`
	IsUnconfirmed     int64  `public:"view" json:"is_unconfirmed"`
}
