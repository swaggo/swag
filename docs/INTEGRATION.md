# Plan: Integrate Custom Struct Parser with Public/Private Schema Variants

Your `LookupStructFields` correctly resolves embedded generic fields using `go/packages` type information. Swag's AST-based `parseStruct` cannot handle `fields.StructField[T]` generics. The integration must **eagerly generate both public and non-public schema variants** for each type, with operations selecting the appropriate version based on an `@public` annotation. Generic fields follow pattern `StructField[T]` where `T` is the actual field type (e.g., `User`, `[]User`, `map[string]User`).

## Steps

### 1. Add ToSpecSchema with Public Context in model/struct_field.go

Create `ToSpecSchema(public bool) (propName string, schema *spec.Schema, required bool, nestedTypes []string, err error)` that:
- Extracts property name from json tag (first part before comma)
- Filters field if `public=true` and `!IsPublic()`
- Detects `fields.StructField[T]` pattern and extracts type parameter `T` using bracket parsing
- Generates schema for `T` with struct type names suffixed `Public` when `public=true`
- Determines required from absence of `omitempty` in json tag
- Returns list of struct type names encountered for recursive definition generation

### 2. Create BuildSpecSchema in model/struct_builder.go

Add `BuildSpecSchema(typeName string, public bool) (schema *spec.Schema, nestedStructs []string, err error)` that:
- Iterates `Fields` calling each `field.ToSpecSchema(public)`
- Aggregates properties and required arrays into single object schema
- Collects all unique nested struct type names from all fields
- Returns main schema plus flat list of struct dependencies for recursive processing

### 3. Add Eager Dual Schema Generation in model/struct_field_lookup.go

Create `BuildAllSchemas(baseModule, pkgPath, typeName string) (allSchemas map[string]*spec.Schema, err error)` that:
- Calls `LookupStructFields`
- Generates `builder.BuildSpecSchema(typeName, false)` and `builder.BuildSpecSchema(typeName+"Public", true)`
- Recursively calls `BuildAllSchemas` for each nested struct discovered
- Stores all schemas (base and Public variants) in single map
- Ensures complete transitive closure of all referenced types and their variants

### 4. Detect and Integrate in ParseDefinition in parser.go

- Add helper `requiresCustomParser(typeSpecDef) bool` that checks if struct has fields importing `fields` package or using `StructField` types
- When true call `BuildAllSchemas` instead of `parseStruct`
- Store all returned schemas in `parser.swagger.Definitions`
- Cache all variants in `parsedSchemas` mapping from original `TypeSpecDef` to base schema (track Public variant separately)
- Return appropriate schema based on context

### 5. Add @public Parsing and Schema Selection in operation.go

- Add `IsPublic bool` field to `Operation`
- Add `publicAttr = "@public"` constant
- Parse in operation comment loop checking `strings.ToLower(commentLine) == publicAttr`
- In `ParseResponseComment` and parameter parsing when building type references append `Public` suffix to all struct type names if `operation.IsPublic`
- Leave primitive types unchanged

## Further Considerations

### 1. Generic Type Parameter Parsing

For `StructField[map[string][]User]`, need robust bracket counting to extract full type including nested brackets. Should handle pointer prefixes like `StructField[*User]` by preserving the `*` in extracted type.

### 2. Schema Reference Strategy

When `ToSpecSchema` encounters nested struct in public mode (e.g., field type `User`), should it return `$ref: "#/definitions/UserPublic"` or just type name `UserPublic` for later reference resolution? Follow existing swag pattern from `RefSchema()` helper.

### 3. Primitive Array/Map Handling

For `StructField[[]string]` or `StructField[map[string]int]`, extracted type is primitive collection. Schema should be array/object with primitive items, no definition references created. Type checking in `ToSpecSchema` needed to distinguish struct vs primitive extraction results.





**IMPORTANT** 
After implementing each step above, add in proper unit testing, you can use 

`/Users/griffnb/projects/swag/testdata/core_models` for a test package and models to test against that the abovev code works properly.

No task is complete without a functioning unit test!