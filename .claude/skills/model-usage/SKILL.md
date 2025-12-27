---
name: model-usage
description: Working with model instances and fields
---

# Using Models

Models provide thread-safe field access through typed methods. All operations use `.Set()` and `.Get()` patterns.

## Creating Model Instances

### Basic Constructor

```go
// Create new model instance
user := user.New()
```

### Typed Constructor

For specific types (like joined models):

```go
// Create a new model instance for a specific type
joinedUser := user.NewType[*user.JoinedUser]()
```

## Setting Field Values

All fields use `.Set(value)` method:

```go
user.Name.Set("Alice")
user.Email.Set("alice@example.com")
user.Age.Set(30)
user.Status.Set(constants.StatusActive)
```

## Getting Field Values

All fields use `.Get()` method:

```go
name := user.Name.Get()      // "Alice"
email := user.Email.Get()    // "alice@example.com"
age := user.Age.Get()        // 30
status := user.Status.Get()  // constants.StatusActive
```

## Working with Struct Fields

Struct fields (JSONB columns) have additional methods:

### Setting Struct Values

```go
type Bookmarks struct {
    FavoriteID types.UUID `json:"favorite_id"`
    RecentIDs  []types.UUID `json:"recent_ids"`
}

user.Bookmarks.Set(&Bookmarks{
    FavoriteID: uuid1,
    RecentIDs:  []types.UUID{uuid2, uuid3},
})
```

### Getting Struct Values

```go
// .Get() returns value and error
bookmarks, err := user.Bookmarks.Get()
if err != nil {
    return err
}

// .GetI() ignores errors (use when error checking not needed)
bookmarks := user.Bookmarks.GetI()
```

## Field Types Reference

| Field Type | Go Type | Database Type | Use Case |
|------------|---------|---------------|----------|
| `StringField` | `string` | `text` | Text/string columns |
| `UUIDField` | `types.UUID` | `uuid` | UUID columns |
| `IntField` | `int` | `integer/smallint` | Integer/boolean (0/1) columns |
| `DecimalField` | `decimal.Decimal` | `numeric` | Decimal/numeric columns |
| `IntConstantField[T]` | `T` (int-based) | `smallint` | Enum/constant fields |
| `StructField[T]` | `T` | `jsonb` | JSONB/struct columns |

## Thread Safety

All field operations are thread-safe by default. Models use internal mutexes to protect concurrent access.

```go
// Safe to use from multiple goroutines
go user.Name.Set("Alice")
go user.Email.Set("alice@example.com")
```

## Common Patterns

### Updating Multiple Fields

```go
user := user.New()
user.Name.Set("Alice")
user.Email.Set("alice@example.com")
user.Age.Set(30)
user.Status.Set(constants.StatusActive)

err := user.Save(ctx)
if err != nil {
    return err
}
```

### Reading and Modifying

```go
user, err := user.GetByID(ctx, userID)
if err != nil {
    return err
}

// Modify fields
currentAge := user.Age.Get()
user.Age.Set(currentAge + 1)

err = user.Save(ctx)
if err != nil {
    return err
}
```

### Working with Optional Fields

Fields with `null:"true"` can be set to nil:

```go
// Set to nil
user.MiddleName.SetNull()

// Check if nil
user.MiddleName.IsNull()

```

## File Organization Standards

- **queries.go** - All specific queries
- **functions.go** - All model functions (not methods)
- **Methods** - Use pointer receivers with `this` as the receiver variable

## Related Skills

- [model-queries](../model-queries/SKILL.md) - Building database queries
- [model-conventions](../model-conventions/SKILL.md) - Standards and field type reference
