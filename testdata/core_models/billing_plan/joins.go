package billing_plan

import (
	"github.com/griffnb/core/lib/model"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{}...)
	options.WithIncludeFields([]string{}...)
}
