//go:generate core_gen model BillingPlan
package billing_plan

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/swaggo/swag/testdata/core_models/base"
)

// Constants for the model
const (
	TABLE        = "billing_plans"
	CHANGE_LOGS  = true
	CLIENT       = "default"
	IS_VERSIONED = false
)

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	Name            *fields.StringField              `public:"view" column:"name"              type:"text"     default:""`
	Description     *fields.StringField              `public:"view" column:"description"       type:"text"     default:""`
	InternalName    *fields.StringField              `              column:"internal_name"     type:"text"     default:""`
	FeatureSet      *fields.StructField[*FeatureSet] `public:"view" column:"feature_set"       type:"jsonb"    default:"{}"`
	Properties      *fields.StructField[*Properties] `public:"view" column:"properties"        type:"jsonb"    default:"{}"`
	StripeProductID *fields.StringField              `public:"view" column:"stripe_product_id" type:"text"     default:""   index:"true"`
	Level           *fields.IntField                 `public:"view" column:"level"             type:"smallint" default:"0"`
	IsDefault       *fields.IntField                 `public:"view" column:"is_default"        type:"smallint" default:"0"  index:"true"`
}

type JoinData struct {
	CreatedByName *fields.StringField `json:"created_by_name" type:"text"`
	UpdatedByName *fields.StringField `json:"updated_by_name" type:"text"`
}

// BillingPlan - Database model
type BillingPlan struct {
	model.BaseModel
	DBColumns
}

type BillingPlanJoined struct {
	BillingPlan
	JoinData
}

func (this *BillingPlan) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	return this.ValidateSubStructs()
}

func (this *BillingPlan) afterSave(ctx context.Context) {
	this.BaseAfterSave(ctx)
	/*
		go func() {
			err := this.UpdateCache()
			if err != nil {
				log.Error(err)
			}
		}()
	*/
}
