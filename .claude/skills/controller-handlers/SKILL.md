---
name: controller-handlers
description: Writing controller handler functions
---

# Writing Controller Handlers

Controller handler functions follow a standardized pattern with consistent return types and error handling.

## Handler Function Signature

All handler functions use this signature:

```go
func handlerName(_ http.ResponseWriter, req *http.Request) (*ModelType, int, error)
```

**Parameters:**
- `_` - ResponseWriter (unused, handled by wrapper)
- `req` - HTTP request with context, params, and body

**Returns:**
- `*ModelType/ResponseDataType` - The response data (nil on error)
- `int` - HTTP status code (200, 400, 404, etc.)
- `error` - Error object (nil on success)

## Simple Handler Example

```go
func adminGet(_ http.ResponseWriter, req *http.Request) (*account.AccountJoined, int, error) {
    // Get the URL parameter
    id := chi.URLParam(req, "id")

    // Fetch the model using the repository pattern
    accountObj, err := account.GetJoined(req.Context(), types.UUID(id))
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*account.AccountJoined](err)
    }

    // Return success with the model data
    return response.Success(accountObj)
}
```

## Return Type Helpers

### Success Response

```go
return response.Success(data)
```

Returns: `(data, http.StatusOK, nil)`

### Admin Error Responses

For admin endpoints - returns full error details:

```go
// Bad request with error message
return response.AdminBadRequestError[*ModelType](err)
// Returns: (zeroValue, http.StatusBadRequest, err)

// Not found
return response.AdminNotFoundError[*ModelType]()
// Returns: (zeroValue, http.StatusNotFound, standardError)

// Forbidden
return response.AdminForbiddenError[*ModelType]()
// Returns: (zeroValue, http.StatusForbidden, standardError)
```

### Public Error Responses

For public endpoints - returns sanitized error messages:

```go
// Bad request (generic public error)
return response.PublicBadRequestError[*ModelType]()
// Returns: (zeroValue, http.StatusBadRequest, publicError)

// Not found
return response.PublicNotFoundError[*ModelType]()
// Returns: (zeroValue, http.StatusNotFound, publicError)

// Forbidden
return response.PublicForbiddenError[*ModelType]()
// Returns: (zeroValue, http.StatusForbidden, publicError)
```

**Important**: Public errors never expose internal error details to users.

## Accessing Request Data

### URL Parameters

```go
// Get URL parameter from route like /account/{id}
id := chi.URLParam(req, "id")
name := chi.URLParam(req, "name")
```

### Session Data

```go
// Get the current user session (available in all authenticated endpoints)
userSession := request.GetReqSession(req) // returns a session object
userObj := helpers.GetLoadedUser(req) // returns the actual user object, not a session wrap

```

### POST/PUT Request Body

```go
// For POST/PUT endpoints that are not the standard crud, use a struct for the input

data, err := request.GetJSONPostAs[*MyPostData](req)

doSomething(input.SomeDataHere)
```

### Query Parameters

```go
// Access query string parameters
queryParams := req.URL.Query()
page := queryParams.Get("page")
limit := queryParams.Get("limit")

// embeeded route params
id := chi.URLParam(req, "id")
```

## Common Handler Patterns
## Error Logging

Always log errors with the context before returning up.  The controller is the log point.

```go
if err != nil {
    log.ErrorContext(err, req.Context())
    return response.AdminBadRequestError[*ModelType](err)
}
```

This ensures errors are captured in logs with full request context.

## Handler Wrapper Usage

Handlers are wrapped in `setup.go`:

```go
// Admin endpoint - shows full errors
helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_ADMIN: response.StandardRequestWrapper(adminCreate),
})

// Public endpoint - sanitizes errors
helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authGet),
})
```



## Request Wrapper Types

### StandardRequestWrapper (Admin)

For admin endpoints:
```go
response.StandardRequestWrapper(adminHandler)
```

**Features:**
- Returns full error details
- No field filtering
- Detailed error messages for debugging

### StandardPublicRequestWrapper (Public)

For public/authenticated user endpoints:
```go
response.StandardPublicRequestWrapper(authHandler)
```

**Features:**
- Filters response fields based on `public:"view"` tags
- Validates update fields based on `public:"edit"` tags
- Returns sanitized error messages
- Prevents internal error leakage

## Security Best Practices

1. **Use appropriate wrappers**: Admin handlers with `StandardRequestWrapper`, public handlers with `StandardPublicRequestWrapper`

2. **Verify ownership**: In auth handlers, always verify the user owns the resource:
```go
if accountObj.ID_.Get() != session.AccountID {
    return response.PublicForbiddenError[*account.Account]()
}
```

3. **Principle of least privilege**: Use the lowest role required for each endpoint

4. **Don't mix admin and public logic**: Keep admin and auth handlers separate

## Related Skills

- [controller-roles](../controller-roles/SKILL.md) - Role-based access control
- [model-usage](../../model-usage/SKILL.md) - Working with models
