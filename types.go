package swag

import (
	"github.com/go-openapi/spec"
	"go/ast"
)

type Schema struct {
	PkgPath      string //package import path used to rename Name of a definition int case of conflict
	Name         string //Name in definitions
	*spec.Schema        //
}

type TypeSpecDef struct {
	//path of package starting from under ${GOPATH}/src or from module path in go.mod
	PkgPath string

	//ast file where TypeSpec is
	File *ast.File

	//the TypeSpec of this type definition
	TypeSpec *ast.TypeSpec
}

func (t *TypeSpecDef) Name() string {
	return t.TypeSpec.Name.Name
}

func (t *TypeSpecDef) FullName() string {
	return fullTypeName(t.File.Name.Name, t.TypeSpec.Name.Name)
}

type AstFileInfo struct {
	File        *ast.File
	Path        string
	PackagePath string
}

//PackageDefinitions files and definition in a package
type PackageDefinitions struct {
	//package name
	Name string

	//files in this package, map key is file's relative path starting package path
	Files map[string]*ast.File

	//definitions in this package, map key is typeName
	TypeDefinitions map[string]*TypeSpecDef
}
