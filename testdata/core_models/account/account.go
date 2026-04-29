//go:generate core_gen model Account -skip=load -add=swaggo
package account

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/swaggo/swag/testdata/core_models/base"
	"github.com/swaggo/swag/testdata/core_models/constants"
)

// Constants for the model
const (
	TABLE        = "accounts"
	CHANGE_LOGS  = true
	CLIENT       = "default"
	IS_VERSIONED = false
)

type Structure struct {
	DBColumns
	JoinData
	ManualFields
	PlanJoins
}

type DBColumns struct {
	base.Structure
	FirstName           *fields.StringField                                 `public:"edit" column:"first_name"             type:"text"     default:""`
	LastName            *fields.StringField                                 `public:"edit" column:"last_name"              type:"text"     default:""`
	Email               *fields.StringField                                 `public:"edit" column:"email"                  type:"text"     default:""     unique:"true"`
	Phone               *fields.StringField                                 `public:"edit" column:"phone"                  type:"text"     default:""`
	ExternalID          *fields.StringField                                 `public:"view" column:"external_id"            type:"text"     default:""                   index:"true"`
	TestUserType        *fields.IntField                                    `public:"view" column:"test_user_type"         type:"smallint" default:"0"`
	OrganizationID      *fields.UUIDField                                   `public:"view" column:"organization_id"        type:"uuid"     default:"null"               index:"true" null:"true"`
	Role                *fields.IntConstantField[constants.Role]            `public:"view" column:"role"                   type:"smallint" default:"1"                  index:"true"`
	Properties          *fields.StructField[*Properties]                    `              column:"properties"             type:"jsonb"    default:"{}"`
	SignupProperties    *fields.StructField[*SignupProperties]              `              column:"signup_properties"      type:"jsonb"    default:"{}"`
	HashedPassword      *fields.StringField                                 `              column:"hashed_password"        type:"text"     default:""`
	PasswordUpdatedAtTS *fields.IntField                                    `              column:"password_updated_at_ts" type:"bigint"   default:"0"`
	EmailVerifiedAtTS   *fields.IntField                                    `              column:"email_verified_at_ts"   type:"bigint"   default:"0"`
	LastLoginTS         *fields.IntField                                    `              column:"last_login_ts"          type:"bigint"   default:"0"                  index:"true"`
	Authentication      *fields.StructField[*Authentication]                `              column:"authentication"         type:"jsonb"    default:"{}"                                          swaggerignore:"true"`
	GlobalConfigKey     *fields.IntConstantField[constants.GlobalConfigKey] `public:"view" column:"config_key"             type:"text"     default:""                   index:"true"`
}

type ManualFields struct {
	IsSuperUserSession *fields.IntField `public:"view" json:"is_super_user_session" type:"smallint"`
}

// Account - Database model
type Account struct {
	model.BaseModel
	DBColumns
	ManualFields
}

// AccountWithFeatures combines account with billing plan features
// @name UserWithFeatures
type AccountWithFeatures struct {
	Account
	PlanJoins
}

type AccountJoined struct {
	Account
	JoinData
}

type initializable interface {
	coremodel.Model
	InitializeWithChangeLogs(*model.InitializeOptions)
	Load(result map[string]any)
}

func load[T initializable](result map[string]any) T {
	obj := NewType[T]()
	obj.Load(result)

	switch o := any(obj).(type) {
	case *AccountWithFeatures:
		o.BuildFeatures()
	default:
	}
	return obj
}

func (this *Account) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)

	return nil
}

func (this *Account) afterSave(ctx context.Context) {
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
