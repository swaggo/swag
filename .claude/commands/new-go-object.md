
## Creating an object
I want you to study the instructions at `./docs/MODELS.md`

Be sure to use #code_tools only for creation/testing
Be sure all database fields are snake_case
Be sure all sub structs that are in jsonb collumns are also snake_case
Be sure to not use booleans for fields, use smallint 0/1 
Be sure to ask if its a public or internal object before starting
Be sure to ask any other questions about fields before starting
Be sure to add the controller to the router `./internal/controllers/router.go`
Be sure to add the model to the loader `./internal/models/loader.go`

Be sure to follow best practices with the model structure, 
functions in `functions.go`
queries in `queries.go`
Be sure jsonb are not just map[string]any and are actually a struct
sub structs in their own my_sub_struct.go file
constants in their own `my_constant_name.go` file within the same package, keep the files pure by what they are doing





