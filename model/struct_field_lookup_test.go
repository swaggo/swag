package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAllSchemas_BillingPlan(t *testing.T) {
	// Test with the actual BillingPlan model from testdata
	baseModule := "github.com/swaggo/swag"
	pkgPath := "github.com/swaggo/swag/testdata/core_models/billing_plan"
	typeName := "BillingPlan"

	schemas, err := BuildAllSchemas(baseModule, pkgPath, typeName)
	require.NoError(t, err)
	require.NotNil(t, schemas)

	// Should have both billing_plan.BillingPlan and billing_plan.BillingPlanPublic
	assert.Contains(t, schemas, "billing_plan.BillingPlan")
	assert.Contains(t, schemas, "billing_plan.BillingPlanPublic")

	// Check BillingPlan schema
	billingPlan := schemas["billing_plan.BillingPlan"]
	assert.NotNil(t, billingPlan)
	assert.Equal(t, 1, len(billingPlan.Type))
	assert.Equal(t, "object", billingPlan.Type[0])

	// BillingPlan should have all fields (public and private)
	assert.Contains(t, billingPlan.Properties, "name")
	assert.Contains(t, billingPlan.Properties, "description")
	assert.Contains(t, billingPlan.Properties, "internal_name") // This is not public
	assert.Contains(t, billingPlan.Properties, "feature_set")
	assert.Contains(t, billingPlan.Properties, "properties")

	// Check BillingPlanPublic schema
	billingPlanPublic := schemas["billing_plan.BillingPlanPublic"]
	assert.NotNil(t, billingPlanPublic)
	assert.Equal(t, 1, len(billingPlanPublic.Type))
	assert.Equal(t, "object", billingPlanPublic.Type[0])

	// BillingPlanPublic should only have public fields
	assert.Contains(t, billingPlanPublic.Properties, "name")
	assert.Contains(t, billingPlanPublic.Properties, "description")
	assert.NotContains(t, billingPlanPublic.Properties, "internal_name") // This is not public
	assert.Contains(t, billingPlanPublic.Properties, "feature_set")
	assert.Contains(t, billingPlanPublic.Properties, "properties")

	// Check nested schemas were generated
	assert.Contains(t, schemas, "billing_plan.FeatureSet")
	assert.Contains(t, schemas, "billing_plan.FeatureSetPublic")
	assert.Contains(t, schemas, "billing_plan.Properties")
	assert.Contains(t, schemas, "billing_plan.PropertiesPublic")
}

func TestBuildAllSchemas_Account(t *testing.T) {
	// Test with Account which has nested structs from different packages
	baseModule := "github.com/swaggo/swag"
	pkgPath := "github.com/swaggo/swag/testdata/core_models/account"
	typeName := "Account"

	schemas, err := BuildAllSchemas(baseModule, pkgPath, typeName)
	require.NoError(t, err)
	require.NotNil(t, schemas)

	// Should have both account.Account and account.AccountPublic
	assert.Contains(t, schemas, "account.Account")
	assert.Contains(t, schemas, "account.AccountPublic")

	// Check Account schema
	account := schemas["account.Account"]
	assert.NotNil(t, account)

	// Account should have all fields
	assert.Contains(t, account.Properties, "first_name")
	assert.Contains(t, account.Properties, "last_name")
	assert.Contains(t, account.Properties, "email")
	assert.Contains(t, account.Properties, "properties")
	assert.Contains(t, account.Properties, "signup_properties")
	assert.Contains(t, account.Properties, "hashed_password") // This is not public

	// Check AccountPublic schema
	accountPublic := schemas["account.AccountPublic"]
	assert.NotNil(t, accountPublic)

	// AccountPublic should only have public fields
	assert.Contains(t, accountPublic.Properties, "first_name")
	assert.Contains(t, accountPublic.Properties, "last_name")
	assert.Contains(t, accountPublic.Properties, "email")
	assert.NotContains(t, accountPublic.Properties, "properties")        // This is not public
	assert.NotContains(t, accountPublic.Properties, "signup_properties") // This is not public
	assert.NotContains(t, accountPublic.Properties, "hashed_password")   // This is not public

	// Check nested schemas were generated
	assert.Contains(t, schemas, "account.Properties")
	assert.Contains(t, schemas, "account.PropertiesPublic")
	assert.Contains(t, schemas, "account.SignupProperties")
	assert.Contains(t, schemas, "account.SignupPropertiesPublic")
}

func TestBuildAllSchemas_WithPackageQualifiedNested(t *testing.T) {
	// Test AccountWithFeatures which has billing_plan.FeatureSet
	baseModule := "github.com/swaggo/swag"
	pkgPath := "github.com/swaggo/swag/testdata/core_models/account"
	typeName := "AccountWithFeatures"

	schemas, err := BuildAllSchemas(baseModule, pkgPath, typeName)
	require.NoError(t, err)
	require.NotNil(t, schemas)

	// Should have AccountWithFeatures schemas
	assert.Contains(t, schemas, "account.AccountWithFeatures")
	assert.Contains(t, schemas, "account.AccountWithFeaturesPublic")

	// Check that nested FeatureSet is included
	// When embedded in account package types, FeatureSet gets prefixed with account
	assert.Contains(t, schemas, "account.FeatureSet")
	assert.Contains(t, schemas, "account.FeatureSetPublic")
}

func TestBuildAllSchemas_InvalidType(t *testing.T) {
	baseModule := "github.com/swaggo/swag"
	pkgPath := "github.com/swaggo/swag/testdata/core_models/account"
	typeName := "NonExistentType"

	schemas, err := BuildAllSchemas(baseModule, pkgPath, typeName)

	// Should handle gracefully - might return empty or error
	// The important thing is it doesn't panic
	if err != nil {
		t.Logf("Expected error for non-existent type: %v", err)
	} else {
		assert.NotNil(t, schemas)
	}
}
