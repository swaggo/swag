---
name: model-conventions
description: Standards, conventions, and struct tag reference for models
---

# Model Conventions and Standards

This skill covers all standards, conventions, and struct tag annotations for the model system.

## Database Field Naming

**All database fields must use snake_case:**

```go
// Correct
UserID      *fields.UUIDField   `column:"user_id" ...`
FirstName   *fields.StringField `column:"first_name" ...`
CreatedAt   *fields.IntField    `column:"created_at" ...`

// Incorrect
UserID      *fields.UUIDField   `column:"userId" ...`    // ❌ camelCase
FirstName   *fields.StringField `column:"FirstName" ...` // ❌ PascalCase
```

## JSONB Sub-Structs

All sub-structs stored in JSONB columns must also use snake_case for JSON tags:

```go
type Settings struct {
    ThemeColor    string `json:"theme_color"`    // ✅ Correct
    NotifyEmail   bool   `json:"notify_email"`   // ✅ Correct
    LastLoginDate string `json:"lastLoginDate"`  // ❌ Incorrect
}
```

## Boolean Values

**Never use boolean types. Always use `smallint` with 0/1 values:**

```go
// Correct
IsActive *fields.IntField `column:"is_active" type:"smallint" default:"0"`

// Incorrect
IsActive *fields.BoolField `column:"is_active" type:"boolean" default:"false"` // ❌
```

Usage:
```go
user.IsActive.Set(1)  // Active
user.IsActive.Set(0)  // Inactive
```

## Struct Tag Annotations

**Critical:** Struct tags control database migrations and constraints. Include all relevant tags.

### Example with All Tags

```go
type UserV1 struct {
    base.Structure
    Name     *fields.StringField                `column:"name"     type:"text"     default:""`
    Email    *fields.StringField                `column:"email"    type:"text"     default:"" unique:"true" index:"true"`
    Age      *fields.IntField                   `column:"age"      type:"integer"  default:"0" null:"true"`
    Status   *fields.IntConstantField[Status]   `column:"status"   type:"smallint" default:"1"`
    Settings *fields.StructField[*Settings]     `column:"settings" type:"jsonb"    default:"{}"`
    ParentID *fields.UUIDField                  `column:"parent_id" type:"uuid"    default:"null" null:"true"`
}
```

### Available Tags

#### Required Tags

- `column:"name"` - Database column name (snake_case)
- `type:"..."` - Database column type (see types below)

#### Optional Tags (only add when needed)

- `default:"value/null"` - Default value for column
- `null:"true"` - Allow NULL values (omit if NOT NULL)
- `unique:"true"` - Unique constraint (omit if not unique)
- `index:"true"` - Create index on column (omit if not indexed)
- `public:"view|edit"` - For public endpoints (omit if internal only)
  - `view` - Field returned in responses
  - `edit` - Field editable by users in updates

### Database Types

Available values for `type:` tag:

| Type | Use For | Example |
|------|---------|---------|
| `text` | String/text columns | `type:"text"` |
| `jsonb` | JSON/struct columns | `type:"jsonb"` |
| `smallint` | Small integers, booleans (0/1), enums | `type:"smallint"` |
| `integer` | Standard integers | `type:"integer"` |
| `bigint` | Large integers | `type:"bigint"` |
| `uuid` | UUID columns | `type:"uuid"` |
| `date` | Date only | `type:"date"` |
| `datetime` | Date and time | `type:"datetime"` |
| `numeric` | Decimal numbers | `type:"numeric"` |

### Important UUID Rule

**All UUID fields must have `default:"null" null:"true"`:**

```go
// Correct
UserID   *fields.UUIDField `column:"user_id"   type:"uuid" default:"null" null:"true"`
ParentID *fields.UUIDField `column:"parent_id" type:"uuid" default:"null" null:"true"`

// Incorrect
UserID   *fields.UUIDField `column:"user_id" type:"uuid"` // ❌ Missing null handling
```

## Field Types

| Field Type | Go Type | Use For |
|------------|---------|---------|
| `StringField` | `string` | Text/string columns |
| `UUIDField` | `types.UUID` | UUID columns |
| `IntField` | `int` | Integer columns, boolean (0/1) fields |
| `DecimalField` | `decimal.Decimal` | Decimal/numeric columns |
| `IntConstantField[T]` | `T` (int-based) | Enum/constant fields |
| `StructField[T]` | `T` | JSONB/struct columns |

### Field Methods

All fields provide:
- `.Set(val)` - Set field value
- `.Get()` - Get field value

Struct fields additionally provide:
- `.GetI()` - Get value ignoring errors (when error checking not needed)

## File Organization

**Standard file structure for models:**

- **queries.go** - All specific database queries
- **functions.go** - All model functions (not methods)
- **<model>.go** - Generated model struct and methods
- **migrations/<model>.go** - Database migrations

## Method Receivers

**All methods must use pointer receivers with `this` as the receiver variable:**

```go
// Correct
func (this *User) FullName() string {
    return this.FirstName.Get() + " " + this.LastName.Get()
}

// Incorrect
func (u *User) FullName() string {     // ❌ Wrong receiver name
    return u.FirstName.Get() + " " + u.LastName.Get()
}

func (this User) FullName() string {   // ❌ Value receiver instead of pointer
    return this.FirstName.Get() + " " + this.LastName.Get()
}
```

## Thread Safety

All model operations are thread-safe by default. Models use internal mutexes to protect concurrent access.

## Code Generation

**Always use the code generator tool - never hand-write model structs:**

- Use `#code_tools make_object` for internal models
- Use `#code_tools make_public_object` for public-facing models

## Related Skills

- [model-usage](../model-usage/SKILL.md) - Using models and fields
- [model-queries](../model-queries/SKILL.md) - Building database queries
