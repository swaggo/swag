package address

import (
	"github.com/griffnb/core/lib/model"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		"LEFT JOIN accounts ON accounts.id = addresses.account_id",
		"LEFT JOIN addresses related_addresses ON related_addresses.id = addresses.related_address_id",
		"LEFT JOIN accounts related_accounts ON related_accounts.id = related_addresses.account_id",
	}...)
	options.WithIncludeFields([]string{
		"accounts.family_id as account_family_id",
		"related_addresses.account_id as related_address_account_id",
		"CONCAT(related_accounts.first_name, ' ', related_accounts.last_name) as related_address_account_name",
	}...)
}
