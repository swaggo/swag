package swag

import "go/ast"

// PackageDefinitions files and definition in a package.
type PackageDefinitions struct {
	// files in this package, map key is file's relative path starting package path
	Files map[string]*ast.File

	// definitions in this package, map key is typeName
	TypeDefinitions map[string]*TypeSpecDef

	// const variables in this package, map key is the name
	ConstTable map[string]*ConstVariable

	// const variables in order in this package
	OrderedConst []*ConstVariable

	// package name
	Name string
}

func NewPackageDefinitions(name string) *PackageDefinitions {
	return &PackageDefinitions{
		Name:            name,
		Files:           make(map[string]*ast.File),
		TypeDefinitions: make(map[string]*TypeSpecDef),
		ConstTable:      make(map[string]*ConstVariable),
	}
}

// AddFile add a file
func (pkg *PackageDefinitions) AddFile(pkgPath string, file *ast.File) *PackageDefinitions {
	pkg.Files[pkgPath] = file
	return pkg
}

// AddTypeSpec add a type spec.
func (pkg *PackageDefinitions) AddTypeSpec(name string, typeSpec *TypeSpecDef) *PackageDefinitions {
	pkg.TypeDefinitions[name] = typeSpec
	return pkg
}

// AddConst add a const variable.
func (pkg *PackageDefinitions) AddConst(valueSpec *ast.ValueSpec) *PackageDefinitions {
	for i := 0; i < len(valueSpec.Names) && i < len(valueSpec.Values); i++ {
		variable := &ConstVariable{
			Name:    valueSpec.Names[i],
			Type:    valueSpec.Type,
			Value:   valueSpec.Values[i],
			Comment: valueSpec.Comment,
		}
		pkg.ConstTable[valueSpec.Names[i].Name] = variable
		pkg.OrderedConst = append(pkg.OrderedConst, variable)
	}
	return pkg
}
