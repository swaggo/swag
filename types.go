package swag

import (
	"go/ast"

	"github.com/go-openapi/spec"
)

// Schema parsed schema.
type Schema struct {
	*spec.Schema        //
	PkgPath      string // package import path used to rename Name of a definition int case of conflict
	Name         string // Name in definitions
}

// TypeSpecDef the whole information of a typeSpec.
type TypeSpecDef struct {
	// ast file where TypeSpec is
	File *ast.File

	// the TypeSpec of this type definition
	TypeSpec *ast.TypeSpec

	// path of package starting from under ${GOPATH}/src or from module path in go.mod
	PkgPath    string
	ParentSpec ast.Decl
}

// Name the name of the typeSpec.
func (t *TypeSpecDef) Name() string {
	if t.TypeSpec != nil {
		return t.TypeSpec.Name.Name
	}

	return ""
}

// FullName full name of the typeSpec.
func (t *TypeSpecDef) FullName() string {
	var fullName string
	if parentFun, ok := (t.ParentSpec).(*ast.FuncDecl); ok && parentFun != nil {
		fullName = fullTypeNameFunctionScoped(t.File.Name.Name, parentFun.Name.Name, t.TypeSpec.Name.Name)
	} else {
		fullName = fullTypeName(t.File.Name.Name, t.TypeSpec.Name.Name)
	}
	return fullName
}

// FullPath of the typeSpec.
func (t *TypeSpecDef) FullPath() string {
	return t.PkgPath + "." + t.Name()
}

// AstFileInfo information of an ast.File.
type AstFileInfo struct {
	// File ast.File
	File *ast.File

	// Path the path of the ast.File
	Path string

	// PackagePath package import path of the ast.File
	PackagePath string
}

// PackageDefinitions files and definition in a package.
type PackageDefinitions struct {
	// files in this package, map key is file's relative path starting package path
	Files map[string]*ast.File

	// definitions in this package, map key is typeName
	TypeDefinitions map[string]*TypeSpecDef

	// package name
	Name string
}
