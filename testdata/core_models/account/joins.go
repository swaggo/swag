package account

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/swaggo/swag/testdata/core_models/billing_plan"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		"LEFT JOIN organizations ON organizations.id = accounts.organization_id",
	}...)
	options.WithIncludeFields([]string{
		"concat(accounts.first_name, ' ', accounts.last_name) as name",
		"organizations.name as organization_name",
	}...)
}

type PlanJoins struct {
	// Only on with AddPlans to join
	BillingPlanLevel      *fields.IntField                                       `public:"view" json:"billing_plan_level"       type:"smallint"`
	BillingPlanPrice      *fields.DecimalField                                   `public:"view" json:"billing_plan_price"       type:"numeric"`
	BillingPlanID         *fields.UUIDField                                      `public:"view" json:"billing_plan_id"          type:"uuid"`
	BillingPlanName       *fields.StringField                                    `public:"view" json:"billing_plan_name"        type:"text"`
	BillingPlanFeatureSet *fields.StructField[*billing_plan.FeatureSet]          `              json:"billing_plan_feature_set" type:"jsonb"`
	FeatureSetOverrides   *fields.StructField[*billing_plan.MergeableFeatureSet] `              json:"feature_set_overrides"    type:"jsonb"    default:"{}"`
	FeatureSet            *fields.StructField[*billing_plan.FeatureSet]          `public:"view" json:"feature_set"              type:"jsonb"`
}

func AddPlans(options *model.Options) *model.Options {
	options.WithPrependJoins([]string{
		"LEFT JOIN billing_plan_prices ON billing_plan_prices.id = organizations.billing_plan_price_id",
		"LEFT JOIN billing_plans ON billing_plans.id = billing_plan_prices.billing_plan_id",
	}...)
	options.WithIncludeFields([]string{
		"billing_plans.name AS billing_plan_name",
		"billing_plans.id AS billing_plan_id",
		"billing_plan_prices.price AS billing_plan_price_price",
		"billing_plans.feature_set AS billing_plan_feature_set",
		"billing_plans.properties AS billing_plan_properties",
		"billing_plans.level AS billing_plan_level",
		"organizations.feature_set_overrides as feature_set_overrides",
	}...)
	return options
}
