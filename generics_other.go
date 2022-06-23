//go:build !go1.18
// +build !go1.18

package swag

func typeSpecFullName(typeSpecDef *TypeSpecDef) string {
	return typeSpecDef.FullName()
}

func (pkgDefs *PackagesDefinitions) parametrizeStruct(original *TypeSpecDef, fullGenericForm string) *TypeSpecDef {
	return original
}
