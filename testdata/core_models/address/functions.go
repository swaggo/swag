package address

import (
	"fmt"
	"strings"

	"github.com/griffnb/core/lib/model/coremodel"
)

// CreateRelatedAddress creates a new address with the same data as the current address, Account is removed
// Used for onboarding directly from a shared address
func (this *Address) CreateRelatedAddress() (*Address, error) {
	data := this.GetDataCopy()

	delete(data, "id")
	delete(data, "created_at")
	delete(data, "updated_at")
	delete(data, "created_by_urn")
	delete(data, "updated_by_urn")
	delete(data, "urn")
	delete(data, "locked_parent_id")
	delete(data, "account_id")
	delete(data, "is_primary")
	delete(data, "is_locked")
	delete(data, "origin_address_id")

	newAddress := New()
	newAddress.MergeData(data)

	return newAddress, nil
}

// String  street street2 city, state, postal code
func (this *Address) String() string {
	if this.Street2.IsEmpty() {
		return fmt.Sprintf("%s, %s, %s %s", this.Street.Get(), this.City.Get(), strings.ToUpper(this.State.Get()), this.PostalCode.Get())
	}

	return fmt.Sprintf(
		"%s %s, %s, %s %s",
		this.Street.Get(),
		this.Street2.Get(),
		this.City.Get(),
		strings.ToUpper(this.State.Get()),
		this.PostalCode.Get(),
	)
}

func (this *Address) Lock(savingUser coremodel.Model) error {
	if this.IsLocked.Bool() {
		return nil
	}
	this.IsLocked.Set(1)
	return this.Save(savingUser)
}

func (this *Address) HasImportantChanges() bool {
	if this.Street.HasSignificantChanged() {
		return true
	}

	if this.Street2.HasSignificantChanged() {
		return true
	}
	if this.City.HasSignificantChanged() {
		return true
	}
	if this.State.HasSignificantChanged() {
		return true
	}

	if this.PostalCode.HasSignificantChanged() {
		return true
	}

	return false
}

func (this *Address) StreetNumber() string {
	streetParts := strings.Split(this.Street.Get(), " ")
	if len(streetParts) > 0 {
		return streetParts[0]
	}
	return ""
}

func (this *Address) StreetName() string {
	streetParts := strings.Split(this.Street.Get(), " ")
	if len(streetParts) > 0 {
		return strings.Join(streetParts[1:], " ")
	}
	return ""
}
