package account

import "github.com/swaggo/swag/testdata/core_models/constants"

const (
	STATUS_PENDING_INVITE             constants.Status = 1
	STATUS_PENDING_EMAIL_VERIFICATION constants.Status = 2
	STATUS_PENDING_ONBOARD            constants.Status = 3
	STATUS_ACTIVE                     constants.Status = 100
	STATUS_DISABLED                   constants.Status = 200
	STATUS_USER_DISABLED              constants.Status = 201
	STATUS_DELETED                    constants.Status = 300
	STATUS_USER_DELETED               constants.Status = 301
)

func StatusToString(status constants.Status) string {
	switch status {
	case STATUS_PENDING_INVITE:
		return "Pending Invite"
	case STATUS_PENDING_EMAIL_VERIFICATION:
		return "Pending Email Verification"
	case STATUS_PENDING_ONBOARD:
		return "Pending Onboard"
	case STATUS_ACTIVE:
		return "Active"
	case STATUS_DISABLED:
		return "Disabled"
	case STATUS_USER_DISABLED:
		return "User Disabled"
	case STATUS_DELETED:
		return "Deleted"
	case STATUS_USER_DELETED:
		return "User Deleted"
	default:
		return "Unknown Status"
	}
}
