package api

import (
	"net/http"

	"github.com/griffnb/core/lib/types"
	"github.com/swaggo/swag/testdata/core_models/account"
	"github.com/swaggo/swag/testdata/core_models/billing_plan"
)

type TestUserInput struct {
	OrganizationID types.UUID `json:"organization_id"`
}

// adminTestCreate creates a test account and optionally logs in as that account
//
//	@Summary		Create test account
//	@Description	Creates a test account with the provided details and optionally logs in as that account
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			body	body		TestUserInput	true	"Test user details"
//	@Param			login_as	query	string	false	"Set to login as the created user"
//	@Success		200	{object}	response.SuccessResponse{data=account.Account}
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/admin/testUser [post]
func CreateTestAccount(w http.ResponseWriter, r *http.Request) {

}

type APIResponse struct {
	Account *account.Account                `json:"account"`
	Billing *billing_plan.BillingPlanJoined `json:"billing_plan"`
}

// internalAPIAccount retrieves account and organization data for internal API use
//
//	@Summary		Get account with organization
//	@Description	Retrieves account and associated organization data by account ID
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Account ID"
//	@Success		200	{object}	response.SuccessResponse{data=APIResponse}
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/api/account/{id} [get]

func internalAPIAccount(_ http.ResponseWriter, req *http.Request) {
}

// authMe retrieves the current authenticated user's account details
//
//	@Public
//	@Summary		Get current user
//	@Description	Retrieves the authenticated user's account details with joined data
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.SuccessResponse{data=account.AccountJoined}
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/auth/me [get]
func Me(_ http.ResponseWriter, req *http.Request) {

}

// adminIndex lists all accounts with pagination and search
//
//	@Summary		List accounts
//	@Description	Retrieves a paginated list of accounts with optional search
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			q		query	string	false	"Search query"
//	@Param			limit	query	int		false	"Results per page"
//	@Param			offset	query	int		false	"Page offset"
//	@Success		200	{object}	response.SuccessResponse{data=[]account.AccountJoined}
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/admin/accounts [get]
func adminIndex(_ http.ResponseWriter, req *http.Request) {
}

// adminIndex lists all accounts with pagination and search
//
//	 @Public
//		@Summary		List accounts
//		@Description	Retrieves a paginated list of accounts with optional search
//		@Tags			accounts
//		@Accept			json
//		@Produce		json
//		@Param			q		query	string	false	"Search query"
//		@Param			limit	query	int		false	"Results per page"
//		@Param			offset	query	int		false	"Page offset"
//		@Success		200	{object}	response.SuccessResponse{data=[]account.AccountJoined}
//		@Failure		400	{object}	response.ErrorResponse
//		@Router			/admin/accounts [get]
func publicIndex(_ http.ResponseWriter, req *http.Request) {
}
