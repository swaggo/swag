package swag

import (
	"go/ast"
	"strings"
)

type TypeSpecDef struct {
	PkgPath  string
	File     *ast.File
	TypeSpec *ast.TypeSpec
}

func (t *TypeSpecDef) Name() string {
	return t.TypeSpec.Name.Name
}

func (t *TypeSpecDef) FullName() string {
	return fullTypeName(t.File.Name.Name, t.TypeSpec.Name.Name)
}

type PackageDefinitions struct {
	//package name
	Name string

	Files map[string]*ast.File

	//definitions in this package
	TypeDefinitions map[string]*TypeSpecDef
}

//PackagesDefinitions map[package import path]*PackageDefinitions
type PackagesDefinitions map[string]*PackageDefinitions

func (pkgs *PackagesDefinitions) FindTypeSpec(pkgPath string, typeName string) *TypeSpecDef {
	if pd, ok := (*pkgs)[pkgPath]; ok {
		if typeSpec, ok := pd.TypeDefinitions[typeName]; ok {
			return typeSpec
		}
	}
	return nil
}

func (pkgs *PackagesDefinitions) FindPackagePathFromImports(pkg string, file *ast.File) string {
	if file == nil {
		return ""
	}

	if strings.ContainsRune(pkg, '.') {
		pkg = strings.Split(pkg, ".")[0]
	}

	hasAnonymousPkg := false
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

func (pkgs *PackagesDefinitions) FindTypePackage(typeName string, file *ast.File, pkgPath string) (string, string) {
	if IsGolangPrimitiveType(typeName) {
		return "", typeName
	}
	if strings.ContainsRune(typeName, '.') {
		parts := strings.Split(typeName, ".")
		pkgPath = pkgs.FindPackagePathFromImports(parts[0], file)
		typeName = parts[1]
	}
	return pkgPath, typeName
}
