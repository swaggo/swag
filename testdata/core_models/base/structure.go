package base

import (
	"github.com/griffnb/core/lib/model/fields"
	"github.com/swaggo/swag/testdata/core_models/constants"
)

type Structure struct {
	ID_          *fields.UUIDField                          `public:"view" column:"id"             type:"uuid"     pk:"true" default:"gen_random_uuid()"`
	URN          *fields.StringField                        `public:"view" column:"urn"            type:"text"               default:""                  index:"true" unique:"true"`
	CreatedByURN *fields.StringField                        `public:"view" column:"created_by_urn" type:"text"               default:"null"              index:"true"               null:"true"`
	UpdatedByURN *fields.StringField                        `public:"view" column:"updated_by_urn" type:"text"               default:"null"              index:"true"               null:"true"`
	Status       *fields.IntConstantField[constants.Status] `public:"view" column:"status"         type:"smallint"           default:"0"                 index:"true"`
	Deleted      *fields.IntField                           `public:"view" column:"deleted"        type:"smallint"           default:"0"                 index:"true"`
	Disabled     *fields.IntField                           `public:"view" column:"disabled"       type:"smallint"           default:"0"                 index:"true"`
	UpdatedAt    *fields.TimeField                          `public:"view" column:"updated_at"     type:"tswtz"              default:"CURRENT_TIMESTAMP" index:"true"`
	CreatedAt    *fields.TimeField                          `public:"view" column:"created_at"     type:"tswtz"              default:"CURRENT_TIMESTAMP" index:"true"`
}
