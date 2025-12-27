---
name: controller-generation
description: Code generation for CRUD endpoints
---

# Controller Code Generation

The system uses `core_gen` to automatically create standard CRUD operations for controllers.

## Code Generation Command

Add this directive to your `setup.go` file:

```go
//go:generate core_gen controller Account -modelPackage=account
```

**Parameters:**
- `controller` - Command type
- `Account` - Model name (PascalCase)
- `-modelPackage=account` - Package name containing the model
- `-options=admin` if its an admin only controller
- `-skip=xxxYYY,aaaBBB` skip generating this function because we need to customize it

## Generated Files

Running `go generate` creates two files:

**`x_gen_admin.go`** - Admin CRUD handlers:
- `adminIndex` - List all resources
- `adminGet` - Get single resource
- `adminCreate` - Create new resource
- `adminUpdate` - Update existing resource
- `adminCount` - Get total count

(optionally)
**`x_gen_auth.go`** - Public CRUD handlers:
- `authIndex` - List resources (filtered to user's data)
- `authGet` - Get single resource (with ownership check)
- `authCreate` - Create new resource
- `authUpdate` - Update resource (with ownership check)

## Generated Endpoints

| Method | Admin Route | Public Route | Function | Description |
|--------|-------------|--------------|----------|-------------|
| GET | `/admin/account` | `/account` | `adminIndex`, `authIndex` | List resources |
| GET | `/admin/account/{id}` | `/account/{id}` | `adminGet`, `authGet` | Get single resource |
| POST | `/admin/account` | `/account` | `adminCreate`, `authCreate` | Create new resource |
| PUT | `/admin/account/{id}` | `/account/{id}` | `adminUpdate`, `authUpdate` | Update resource |
| GET | `/admin/account/count` | - | `adminCount` | Get total count |
| GET | `/admin/account/_ts` | - | TypeScript | TS type generation |

## Skipping Endpoints

You can disable specific endpoints using the `-skip` parameter:

```go
//go:generate core_gen controller AiTool -modelPackage=ai_tool -skip=authCreate,authUpdate
```

This will generate all endpoints except `authCreate` and `authUpdate`.

**Available Skip Options:**
- `adminIndex` - Skip admin list endpoint
- `adminGet` - Skip admin get endpoint
- `adminCreate` - Skip admin create endpoint
- `adminUpdate` - Skip admin update endpoint
- `adminCount` - Skip admin count endpoint
- `authIndex` - Skip public list endpoint
- `authGet` - Skip public get endpoint
- `authCreate` - Skip public create endpoint
- `authUpdate` - Skip public update endpoint

### Common Skip Patterns

**Read-only public endpoint:**
```go
//go:generate core_gen controller Config -modelPackage=config -skip=authCreate,authUpdate
```

**Admin-only resource:**
```go
//go:generate core_gen controller SystemLog -modelPackage=system_log -options=admin
```


## Customizing Generated Code

**DO NOT edit generated files directly.** They will be overwritten on next generation.

Instead, create custom handlers in separate files, name them accordingly
auth.go
open.go
admin.go

if theres lots of functions, group them together by relation then use a prefix, auth_password.go auth_emails.go

**custom_handlers.go:**
```go
package account

func customSearch(_ http.ResponseWriter, req *http.Request) ([]*account.Account, int, error) {
    // Custom search logic here
    // ...
}
```

Then wire up in `setup.go`:

```go
r.Get("/search", helpers.RoleHandler(helpers.RoleHandlerMap{
    constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(customSearch),
}))
```

## Regenerating Code

Re-run generation after:
- Updating skip parameters

```bash
go generate path/to/file
```

## Related Skills

- [controller-handlers](../controller-handlers/SKILL.md) - Writing custom handlers
- [controller-roles](../controller-roles/SKILL.md) - Role-based access control
