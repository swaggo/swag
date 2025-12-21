package constants

type Role int

const (
	ROLE_UNAUTHORIZED   Role = -1 // Unauthorized
	ROLE_ANY_AUTHORIZED Role = 0  // Any Authorized
	ROLE_USER           Role = 1  // User

	ROLE_ORG_ADMIN Role = 40 // Org Admin
	ROLE_ORG_OWNER Role = 50 // Org Owner

	ROLE_READ_ADMIN      = 90  // Read-Only Admin
	ROLE_ADMIN      Role = 100 // System Admin

)

var DescOrderedAccountRoles = []Role{
	ROLE_ORG_OWNER,
	ROLE_ORG_ADMIN,
	ROLE_USER,
	ROLE_ANY_AUTHORIZED,
	ROLE_UNAUTHORIZED,
}

var DescOrderedAdminRoles = []Role{
	ROLE_ADMIN,
	ROLE_READ_ADMIN,
	ROLE_ANY_AUTHORIZED,
	ROLE_UNAUTHORIZED,
}
