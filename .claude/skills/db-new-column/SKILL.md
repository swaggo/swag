---
name: new-column
description: conventions for new columns
---


Be sure all database fields are snake_case
Be sure all sub structs that are in jsonb collumns are also snake_case
Be sure to not use booleans for fields, use smallint 0/1 


## Struct Tag Annotations

**Critical**: Struct tags control database migrations and constraints. Include all relevant tags:

```go
Name     *fields.StringField `column:"name" type:"text" default:"":"false"`
Email    *fields.StringField `column:"email" type:"text" default:"" unique:"true"`
Age      *fields.IntField    `column:"age" type:"integer" default:"0" null:"true"`
Status   *fields.IntConstantField[Status] `column:"status" type:"smallint" default:"1"`
Settings *fields.StructField[*Settings] `column:"settings" type:"jsonb" default:"{}"`
```

### Available Tags:

- `column:"name"` – Database column name (required)
- `type:"text|jsonb|smallint|integer|uuid|date|datetime|bigint"` – Database column type (required) note that all 'boolean' type things should be a smallint 0/1
- `default:"value/null"` – Default value for column
- `null:"true"` – Whether column allows NULL, dont add if not nullable
- `unique:"true"` – Whether column has unique constraint, dont add if not unique
- `index:"true"` – Whether to create index on column, dont add if not indexed
- `public:"view|edit"` - For public endpoints, determines whether or not the field is returned (view or edit) or on updates if its editable by the user (edit), dont add if not public facing
- IMPORTANT - for all UUID fields, they must have `default:"null" null:"true"`



## Field Types

- `StringField` – Text/string columns
- `UUIDField` – UUID columns
- `IntField` – Integer columns  / Bool fields with smallint 0/1 values
- `DecimalField` – Decimal/numeric columns
- `IntConstantField[T]` – Enum/constant fields
- `StructField[T]` – JSONB/struct columns

All fields provide `.Set(val)` and `.Get()` methods.  Struct fields have a `.GetI()` for when errors do not need to be checked