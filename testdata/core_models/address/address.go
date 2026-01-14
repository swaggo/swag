//go:generate core_generate model Address -version=v2 -add=Joined -skip=Save,SaveWithContext
package address

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/model/fields"
)

const TABLE string = "addresses"

const (
	CHANGE_LOGS       = true
	CLIENT            = "default"
	IS_VERSIONED bool = false
)

type Structure struct {
	DBColumns
	JoinData
	ManualFields
}

type DBColumns struct {
	AccountID           *fields.UUIDField                   `public:"view" column:"account_id"            type:"uuid"     index:"true" null:"true" default:"null"`
	LockedParentID      *fields.UUIDField                   `public:"view" column:"locked_parent_id"      type:"uuid"     index:"true" null:"true" default:"null"`
	IsShared            *fields.IntField                    `public:"edit" column:"is_shared"             type:"smallint" index:"true"             default:"0"`
	Street              *fields.StringField                 `public:"edit" column:"street"                type:"text"                              default:""`
	Street2             *fields.StringField                 `public:"edit" column:"street_2"              type:"text"                              default:""`
	City                *fields.StringField                 `public:"edit" column:"city"                  type:"text"                              default:""`
	State               *fields.StringField                 `public:"edit" column:"state"                 type:"text"                              default:""`
	PostalCode          *fields.StringField                 `public:"edit" column:"postal_code"           type:"text"                              default:""`
	Verification        *fields.StructField[*Verification]  `public:"edit" column:"verification"          type:"jsonb"                             default:"{}"`
	RawVerification     *fields.StructField[map[string]any] `              column:"raw_verification"      type:"jsonb"                             default:"{}"`
	AutoConfirmed       *fields.IntField                    `public:"edit" column:"auto_confirmed"        type:"smallint"                          default:"0"`
	ManualConfirmed     *fields.IntField                    `public:"edit" column:"manual_confirmed"      type:"smallint"                          default:"0"`
	IsPrimary           *fields.IntField                    `public:"view" column:"is_primary"            type:"smallint" index:"true"             default:"0"`
	IsLocked            *fields.IntField                    `              column:"is_locked"             type:"smallint"                          default:"0"`
	Normalization       *fields.StructField[*Normalization] `              column:"normalization"         type:"jsonb"    index:"true"             default:"{}"`
	NormalizedLookupKey *fields.StringField                 `              column:"normalized_lookup_key" type:"text"     index:"true"             default:""`
	NormalizedFullKey   *fields.StringField                 `              column:"normalized_full_key"   type:"text"     index:"true"             default:""`

	Properties *fields.StructField[*Properties] `column:"properties" type:"jsonb" default:"{}"`

	// Address changes were last copied from
	AppliedChangesFromID *fields.UUIDField `public:"edit" column:"applied_changes_from_id" type:"uuid" index:"true" null:"true" default:"null"`
	// Direct related address
	RelatedAddressID *fields.UUIDField `public:"edit" column:"related_address_id"      type:"uuid" index:"true" null:"true" default:"null"`

	// Origin is the first id in the lineage
	OriginID *fields.UUIDField `public:"edit" column:"origin_id"                 type:"uuid" index:"true" null:"true" default:"null"`
	// RelatedOriginAddress is the first address in the lineage of a related address
	OriginRelatedAddressID *fields.UUIDField `public:"edit" column:"origin_related_address_id" type:"uuid" index:"true" null:"true" default:"null"`
}

type JoinData struct {
	AccountFamilyID           *fields.UUIDField   `public:"view" json:"account_family_id"            type:"uuid"`
	RelatedAddressAccountName *fields.StringField `public:"view" json:"related_address_account_name" type:"text"`
	RelatedAddressAccountID   *fields.UUIDField   `public:"view" json:"related_address_account_id"   type:"uuid"`
}

type ManualFields struct {
	SentCount *fields.IntField `json:"sent_count"             public:"view" type:"int"`
	// Used by the rule engine, not meant to be exposed
	SharedWithMe *fields.IntField `json:"shared_with_me"                       type:"smallint"`
	// Used by PrepareDSR
	LastRecommendationID *fields.UUIDField `json:"last_recommendation_id" public:"view" type:"uuid"`
}

type Address struct {
	model.BaseModel
	DBColumns
	ManualFields
}

type AddressJoined struct {
	Address
	JoinData
}

func (this *Address) beforeSave(ctx context.Context, allowLockBypass bool) error {
	this.BaseBeforeSave(ctx)

	return this.ValidateSubStructs()
}

func (this *Address) afterSave(ctx context.Context) {
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

func (this *Address) Save(savingUser coremodel.Model) error {
	return this.SaveWithContext(context.Background(), savingUser)
}

func (this *Address) SaveWithContext(ctx context.Context, savingUser coremodel.Model) error {
	err := this.beforeSave(ctx, false)
	if err != nil {
		return err
	}
	_, err = this.BaseSave(ctx, savingUser)
	if err != nil {
		return err
	}
	this.afterSave(ctx)
	return nil
}

func (this *Address) SaveWithLockBypass(ctx context.Context, savingUser coremodel.Model) error {
	err := this.beforeSave(ctx, true)
	if err != nil {
		return err
	}
	_, err = this.BaseSave(ctx, savingUser)
	if err != nil {
		return err
	}
	this.afterSave(ctx)
	return nil
}
