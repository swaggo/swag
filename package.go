package swag

import (
	"go/ast"
	"strings"
)

// TypeSpecDef typeSpec with its ast.File
type TypeSpecDef struct {
	//path of package starting from under ${GOPATH}/src or from module path in go.mod
	PkgPath string

	//ast file where TypeSpec is
	File *ast.File

	//the TypeSpec of this type definition
	TypeSpec *ast.TypeSpec
}

// Name name of type
func (t *TypeSpecDef) Name() string {
	return t.TypeSpec.Name.Name
}

// FullName name with prefixed package name of type
func (t *TypeSpecDef) FullName() string {
	return fullTypeName(t.File.Name.Name, t.TypeSpec.Name.Name)
}

//PackageDefinitions sorted by packages
type PackageDefinitions struct {
	//package name
	Name string

	//files in this package, map key is file's relative path starting package path
	Files map[string]*ast.File

	//definitions in this package, map key is typeName
	TypeDefinitions map[string]*TypeSpecDef
}

//PackagesDefinitions map[package import path]*PackageDefinitions
type PackagesDefinitions map[string]*PackageDefinitions

func (pkgs *PackagesDefinitions) findTypeSpec(pkgPath string, typeName string) *TypeSpecDef {
	if pd, ok := (*pkgs)[pkgPath]; ok {
		if typeSpec, ok := pd.TypeDefinitions[typeName]; ok {
			return typeSpec
		}
	}
	return nil
}

// findPackagePathFromImports finds out the package path of a package via ranging imports of a ast.File
// @pkg the name of the target package
// @file current ast.File in which to search imports
// @return the package path of a package of @pkg
func (pkgs *PackagesDefinitions) findPackagePathFromImports(pkg string, file *ast.File) string {
	if file == nil {
		return ""
	}

	if strings.ContainsRune(pkg, '.') {
		pkg = strings.Split(pkg, ".")[0]
	}

	hasAnonymousPkg := false

	// prior to match named package
	for _, imp := range file.Imports {
		if imp.Name != nil {
			if imp.Name.Name == pkg {
				return strings.Trim(imp.Path.Value, `"`)
			} else if imp.Name.Name == "_" {
				hasAnonymousPkg = true
			}
		} else {
			path := strings.Trim(imp.Path.Value, `"`)
			if pd, ok := (*pkgs)[path]; ok {
				if pd.Name == pkg {
					return path
				}
			}
		}
	}

	//match unnamed package
	if hasAnonymousPkg {
		for _, imp := range file.Imports {
			if imp.Name == nil {
				continue
			}
			if imp.Name.Name == "_" {
				path := strings.Trim(imp.Path.Value, `"`)
				if pd, ok := (*pkgs)[path]; ok {
					if pd.Name == pkg {
						return path
					}
				}
			}
		}
	}
	return ""
}

// FindTypeSpec finds out TypeSpecDef of a type by typeName
// @typeName the name of the target type, if it starts with a package name, find its own package path from imports on top of @file
// @file the ast.file in which @typeName is used
// @pkgPath the package path of @file
func (pkgs *PackagesDefinitions) FindTypeSpec(typeName string, file *ast.File, pkgPath string) *TypeSpecDef {
	if IsGolangPrimitiveType(typeName) {
		return nil
	}

	if strings.ContainsRune(typeName, '.') {
		parts := strings.Split(typeName, ".")
		typeName = parts[1]
		newPkgPath := pkgs.findPackagePathFromImports(parts[0], file)
		if len(newPkgPath) == 0 && parts[0] == file.Name.Name {
			newPkgPath = pkgPath
		}
		return pkgs.findTypeSpec(newPkgPath, typeName)
	} else if typeDef := pkgs.findTypeSpec(pkgPath, typeName); typeDef != nil {
		return typeDef
	}

	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == "." {
			pkgPath = strings.Trim(imp.Path.Value, `"`)
			if typeDef := pkgs.findTypeSpec(pkgPath, typeName); typeDef != nil {
				return typeDef
			}
		}
	}

	return nil
}
