package billing_plan

type Properties struct {
	PricingText         string `public:"view" json:"pricing_text"`
	DefaultDiscountCode string `              json:"default_discount_code"` // Used if we want to use a standard plan but say its a discount
}
