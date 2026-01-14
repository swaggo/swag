package address

import (
	"testing"
)

func TestAddress_HasImportantChanges(t *testing.T) {
	t.Run("No changes", func(t *testing.T) {
		address := New()
		address.Street.Set("123 Main St")
		address.City.Set("Anytown")
		address.State.Set("CA")
		address.PostalCode.Set("12345")

		address.IsPrimary.Set(1)
		address.IsShared.Set(0)

		address.SyncPreValues()

		if address.HasImportantChanges() {
			t.Errorf("Expected no changes, but got true")
		}
	})

	t.Run("Street changed", func(t *testing.T) {
		address := New()
		address.Street.Set("123 Main St")
		address.City.Set("Anytown")
		address.State.Set("CA")
		address.PostalCode.Set("12345")

		address.IsPrimary.Set(1)
		address.IsShared.Set(0)
		address.SyncPreValues()

		address.Street.Set("124 Main St")

		if !address.HasImportantChanges() {
			t.Errorf("Expected Street change to be detected")
		}
	})

	t.Run("Street space changed", func(t *testing.T) {
		address := New()
		address.Street.Set("123   Main    St")
		address.City.Set("Anytown")
		address.State.Set("CA")
		address.PostalCode.Set("12345")

		address.IsPrimary.Set(1)
		address.IsShared.Set(0)
		address.SyncPreValues()

		address.Street.Set("123 Main St")

		if address.HasImportantChanges() {
			t.Errorf("Expected Street change to not be detected")
		}
	})

	t.Run("Primary Changedd", func(t *testing.T) {
		address := New()
		address.Street.Set("123   Main    St")
		address.City.Set("Anytown")
		address.State.Set("CA")
		address.PostalCode.Set("12345")

		address.IsPrimary.Set(0)
		address.IsShared.Set(0)
		address.SyncPreValues()

		address.Street.Set("123 Main St")

		if address.HasImportantChanges() {
			t.Errorf("Expected Street change to not be detected")
		}
	})
}
