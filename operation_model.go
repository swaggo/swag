package swag

type StructTagValue struct {
	ParamValue string
	Validate   string
}

type StructFieldInfo struct {
	Name     string
	Type     string
	Tag      string
	Comments []string
}
