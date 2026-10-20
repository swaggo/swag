package account

import (
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/tools"
)

func (this *AccountWithFeatures) BuildFeatures() {
	plan, err := this.BillingPlanFeatureSet.Get()
	if err != nil {
		log.Error(err)
		return
	}

	if tools.Empty(plan) {
		return
	}

	overrides, err := this.FeatureSetOverrides.Get()
	if err != nil {
		log.Error(err)
		return
	}

	plan.Merge(overrides)

	this.FeatureSet.Set(plan)
}
