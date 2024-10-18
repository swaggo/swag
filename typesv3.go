package swag

import "github.com/sv-tools/openapi/spec"

// SchemaV3 parsed schema.
type SchemaV3 struct {
	*spec.Schema        //
	PkgPath      string // package import path used to rename Name of a definition int case of conflict
	Name         string // Name in definitions
}
