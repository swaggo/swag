package address

import (
	"fmt"
	"strings"

	"github.com/griffnb/core/lib/tools"
)

type Normalization struct {
	QueryString      string         `json:"query_string,omitempty"`
	AWSAddressNumber string         `json:"aws_address_number,omitempty"`
	AWSStreet        string         `json:"aws_street,omitempty"`
	AWSNeighborhood  string         `json:"aws_neighborhood,omitempty"`
	AWSMunicipality  string         `json:"aws_municipality,omitempty"`
	AWSSubRegion     string         `json:"aws_subregion,omitempty"`
	AWSRegion        string         `json:"aws_region,omitempty"`
	AWSPostalCode    string         `json:"aws_postal_code,omitempty"`
	AWSPostalCode5   string         `json:"aws_postal_code_5,omitempty"`
	AWSCountry       string         `json:"aws_country,omitempty"`
	AWSLabel         string         `json:"aws_label,omitempty"`
	AWSPlaceID       string         `json:"aws_place_id,omitempty"`
	Secondary        *Normalization `json:"secondary,omitempty"`
}

func (this *Normalization) ToLookupKey() string {
	if tools.Empty(this.AWSAddressNumber) || tools.Empty(this.AWSStreet) || tools.Empty(this.AWSPostalCode5) {
		return ""
	}

	return strings.ToLower(fmt.Sprintf("%s|%s|%s",
		this.AWSAddressNumber,
		this.AWSStreet,
		this.AWSPostalCode5,
	))
}

func (this *Normalization) ToFullKey() string {
	value := this.AWSLabel
	if strings.HasSuffix(value, ", United States") {
		value = strings.Replace(value, ", United States", "", 1)
	}

	if strings.HasSuffix(value, ", USA") {
		value = strings.Replace(value, ", USA", "", 1)
	}

	if !tools.Empty(this.AWSPostalCode) && !tools.Empty(this.AWSPostalCode5) {
		value = strings.Replace(value, this.AWSPostalCode, this.AWSPostalCode5, 1)
	}

	return strings.ToLower(value)
}
