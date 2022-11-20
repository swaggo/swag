package swag

const (
	enumVarNamesExtension = "x-enum-varnames"
	enumCommentsExtension = "x-enum-comments"
)

// EnumValue a model to record an enum const variable
type EnumValue struct {
	key     string
	Value   interface{}
	Comment string
}
