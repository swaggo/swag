---
name: controller-roles
description: Role-based access control with RoleHandler
---

# Role-Based Access Control

The `helpers.RoleHandler` function provides role-based access control by mapping roles to specific handler functions.

## Basic Usage

```go
helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminGet),
    constants.ROLE_ADMIN: response.StandardRequestWrapper(adminCreate),
})
```

## Role Hierarchy

Roles are defined as integer constants in descending order of privilege:

| Role | Value | Description |
|------|-------|-------------|
| `ROLE_ADMIN` | 100 | Full system administrator access |
| `ROLE_READ_ADMIN` | 90 | Read-only administrator access |
| `ROLE_ANY_AUTHORIZED` | 0 | Any authenticated user |
| `ROLE_UNAUTHORIZED` | -1 | Unauthenticated requests |

## How RoleHandler Works

1. **Extracts session** from request headers/cookies
2. **Looks up user's role** from the database
3. **Finds highest-privilege handler** the user can access
4. **Falls back** to lower privilege handlers if exact role match isn't found
5. **Returns 401 Unauthorized** if no suitable handler is found

### Fallback Behavior

If a user's role doesn't exactly match a handler, the system checks lower-privilege handlers:

```go
helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminGet),
    constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(authGet),
})
```

**Examples:**
- User with `ROLE_ADMIN` (100) → Uses `ROLE_READ_ADMIN` handler (fallback)
- User with `ROLE_READ_ADMIN` (90) → Uses `ROLE_READ_ADMIN` handler (exact match)
- User with `ROLE_ANY_AUTHORIZED` (0) → Uses `ROLE_ANY_AUTHORIZED` handler (exact match)
- Unauthenticated user → Returns 401 Unauthorized

## Session Context

The `RoleHandler` automatically injects the session into the request context, making it available via:

```go
userSession := request.GetReqSession(req)
```

**Session Fields:**
```go
type Session struct {
    User       coremodel.Model // thin wrapper over session data if you only need the users ID, i.e. sessionObj.User.ID(), or used to save data so we can track who saved it.
	LoadedUser any // fully loaded user from the database, dont access directly, use the helper.GetLoadedUser(req)
}
```

## Common Role Patterns

### Admin-Only Endpoints

Full admin access required:

```go
r.Post("/", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_ADMIN: response.StandardRequestWrapper(adminCreate),
}))

r.Put("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_ADMIN: response.StandardRequestWrapper(adminUpdate),
}))

r.Delete("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_ADMIN: response.StandardRequestWrapper(adminDelete),
}))
```

### Read-Only Admin Access

Both full admins and read-only admins can access:

```go
r.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminIndex),
}))

r.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminGet),
}))

r.Get("/count", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminCount),
}))
```

### Authenticated User Endpoints

Any authenticated user can access:

```go
r.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authIndex),
}))

r.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authGet),
}))
```

### Mixed Role Handlers

Different handlers for different roles on the same route:

```go
r.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_ADMIN: response.StandardRequestWrapper(adminGetFull),
    constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authGetLimited),
}))
```

**Example:**
- Admin users → Get full details via `adminGetFull`
- Regular users → Get limited details via `authGetLimited`



## Related Skills
- [controller-handlers](../controller-handlers/SKILL.md) - Writing handler functions
- [controller-generation](../controller-generation/SKILL.md) - Code generation
