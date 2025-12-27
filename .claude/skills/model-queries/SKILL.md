---
name: model-queries
description: Building database queries with the Options API
---

# Building Database Queries

All database queries use the `Options` struct with generated column helpers for type-safe query building.

## Basic Query Pattern

```go
func GetJoined(ctx context.Context, id types.UUID) (*AdminJoined, error) {
    mocker, ok := model.GetMocker[*Mocker](ctx,PACKAGE)
	if ok {
		return mocker.GetByExternalID(ctx, externalID)
	}
    options := model.NewOptions().
        WithCondition("%s = :id:", Columns.ID_.Column()).
        WithParam(":id:", id)

    return first[*AdminJoined](ctx, options)
}
```

## Query Methods

### WithCondition(format, values...)

Add AND condition to query. Use `:key:` syntax for parameter placeholders:

```go
options := model.NewOptions().
    WithCondition("%s = :user_id:", Columns.UserID.Column()).
    WithCondition("%s > :min_age:", Columns.Age.Column())
```

### WithParam(key, value)

Add query parameter. Key must match placeholder in condition (`:key:`):

```go
options.WithParam(":user_id:", userID).
        WithParam(":min_age:", 18)
```

**Note:** Handles slices automatically when using `IN(:myval:)`:

```go
options.WithCondition("%s IN(:ids:)", Columns.ID_.Column()).
        WithParam(":ids:", []types.UUID{id1, id2, id3})
```

### WithLimit(limit)

Set maximum number of results:

```go
options.WithLimit(10)
```

### WithOrder(order)

Set result ordering:

```go
options.WithOrder("created_at DESC")
```

### WithJoins(joins...)

Add table joins:

```go
options.WithJoins(
    model.Join{
        Table: user.TABLE,
        On:    "user.id = admin.user_id",
    },
)
```

## Column Helpers

Every model has a generated `Columns` struct with helpers for each field:

```go
// Use .Column() to get the column name with table prefix
Columns.ID_.Column()      // Returns "table_name.id"
Columns.Name.Column()     // Returns "table_name.name"
Columns.Email.Column()    // Returns "table_name.email"
```

## File Organization

**All specific queries must go in `queries.go`** within the model package.

**AVOID ADHOC QUERIES**
- they should always be in a function
- the mocker struct inside of queries.go should be updated
- all functions must check mocker context
```go
    mocker, ok := model.GetMocker[*Mocker](ctx,PACKAGE)
	if ok {
		return mocker.GetByExternalID(ctx, externalID)
	}
```

## Example Queries

### Simple Lookup

```go
func GetByEmail(ctx context.Context, email string) (*User, error) {
    mocker, ok := model.GetMocker[*Mocker](ctx,PACKAGE)
	if ok {
		return mocker.GetByExternalID(ctx, externalID)
	}
    options := model.NewOptions().
        WithCondition("%s = :email:", Columns.Email.Column()).
        WithParam(":email:", email).
        WithLimit(1)

    return first[*User](ctx, options)
}
```

### Complex Query with Multiple Conditions

```go
func GetActiveUsers(ctx context.Context, minAge int, roles []int) ([]*User, error) {
    mocker, ok := model.GetMocker[*Mocker](ctx,PACKAGE)
	if ok {
		return mocker.GetByExternalID(ctx, externalID)
	}
    options := model.NewOptions().
        WithCondition("%s = :status:", Columns.Status.Column()).
        WithCondition("%s >= :min_age:", Columns.Age.Column()).
        WithCondition("%s IN(:roles:)", Columns.Role.Column()).
        WithParam(":status:", 1).
        WithParam(":min_age:", minAge).
        WithParam(":roles:", roles).
        WithOrder("created_at DESC")

    return list[*User](ctx, options)
}
```

### Query with Joins
All joins should go into the `joins.go` file

```go
// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		"LEFT JOIN manufacturers ON manufacturers.id = models.manufacturer_id",
		"LEFT JOIN categories ON categories.id = models.category_id",
		`LEFT JOIN (
				SELECT count(1) as asset_count , assets.model_id
				FROM assets
				WHERE assets.disabled = 0
				GROUP BY assets.model_id
		) as asset_counts ON asset_counts.model_id = models.id`,
	}...)
	options.WithIncludeFields([]string{
		"manufacturers.name AS manufacturer_name",
		"categories.name AS category_name",
		"COALESCE(asset_counts.asset_count, 0) AS asset_count",
	}...)
}

```

## Related Skills

- [model-usage](../model-usage/SKILL.md) - How to use models and access fields
- [model-conventions](../model-conventions/SKILL.md) - Standards and conventions
