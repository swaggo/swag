package swag

const (
	enumVarNamesExtension     = "x-enum-varnames"
	enumCommentsExtension     = "x-enum-comments"
	enumDescriptionsExtension = "x-enum-descriptions"
)

// EnumValue a model to record an enum consts variable
type EnumValue struct {
	key     string
	Value   interface{}
	Comment string
}
