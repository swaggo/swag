package swag

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

// PackagesDefinitions map[package import path]*PackageDefinitions.
type PackagesDefinitions struct {
	files             map[*ast.File]*AstFileInfo
	packages          map[string]*PackageDefinitions
	uniqueDefinitions map[string]*TypeSpecDef
}

// NewPackagesDefinitions create object PackagesDefinitions.
func NewPackagesDefinitions() *PackagesDefinitions {
	return &PackagesDefinitions{
		files:             make(map[*ast.File]*AstFileInfo),
		packages:          make(map[string]*PackageDefinitions),
		uniqueDefinitions: make(map[string]*TypeSpecDef),
	}
}

// CollectAstFile collect ast.file.
func (pkgs *PackagesDefinitions) CollectAstFile(packageDir, path string, astFile *ast.File) error {
	if pkgs.files == nil {
		pkgs.files = make(map[*ast.File]*AstFileInfo)
	}

	if pkgs.packages == nil {
		pkgs.packages = make(map[string]*PackageDefinitions)
	}

	// return without storing the file if we lack a packageDir
	if len(packageDir) == 0 {
		return nil
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	pd, ok := pkgs.packages[packageDir]
	if ok {
		// return without storing the file if it already exists
		_, exists := pd.Files[path]
		if exists {
			return nil
		}
		pd.Files[path] = astFile
	} else {
		pkgs.packages[packageDir] = &PackageDefinitions{
			Name:            astFile.Name.Name,
			Files:           map[string]*ast.File{path: astFile},
			TypeDefinitions: make(map[string]*TypeSpecDef),
		}
	}

	pkgs.files[astFile] = &AstFileInfo{
		File:        astFile,
		Path:        path,
		PackagePath: packageDir,
	}

	return nil
}

// RangeFiles for range the collection of ast.File in alphabetic order.
func (pkgs *PackagesDefinitions) RangeFiles(handle func(filename string, file *ast.File) error) error {
	sortedFiles := make([]*AstFileInfo, 0, len(pkgs.files))
	for _, info := range pkgs.files {
		sortedFiles = append(sortedFiles, info)
	}

	sort.Slice(sortedFiles, func(i, j int) bool {
		return strings.Compare(sortedFiles[i].Path, sortedFiles[j].Path) < 0
	})

	for _, info := range sortedFiles {
		err := handle(info.Path, info.File)
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseTypes parse types
// @Return parsed definitions.
func (pkgs *PackagesDefinitions) ParseTypes() (map[*TypeSpecDef]*Schema, error) {
	parsedSchemas := make(map[*TypeSpecDef]*Schema)
	for astFile, info := range pkgs.files {
		for _, astDeclaration := range astFile.Decls {
			generalDeclaration, ok := astDeclaration.(*ast.GenDecl)
			if ok && generalDeclaration.Tok == token.TYPE {
				for _, astSpec := range generalDeclaration.Specs {
					typeSpec, ok := astSpec.(*ast.TypeSpec)
					if ok {
						typeSpecDef := &TypeSpecDef{
							PkgPath:  info.PackagePath,
							File:     astFile,
							TypeSpec: typeSpec,
						}

						idt, ok := typeSpec.Type.(*ast.Ident)
						if ok && IsGolangPrimitiveType(idt.Name) {
							parsedSchemas[typeSpecDef] = &Schema{
								PkgPath: typeSpecDef.PkgPath,
								Name:    astFile.Name.Name,
								Schema:  PrimitiveSchema(TransToValidSchemeType(idt.Name)),
							}
						}

						if pkgs.uniqueDefinitions == nil {
							pkgs.uniqueDefinitions = make(map[string]*TypeSpecDef)
						}

						fullName := typeSpecDef.FullName()
						anotherTypeDef, ok := pkgs.uniqueDefinitions[fullName]
						if ok {
							if typeSpecDef.PkgPath == anotherTypeDef.PkgPath {
								continue
							} else {
								delete(pkgs.uniqueDefinitions, fullName)
							}
						} else {
							pkgs.uniqueDefinitions[fullName] = typeSpecDef
						}

						pkgs.packages[typeSpecDef.PkgPath].TypeDefinitions[typeSpecDef.Name()] = typeSpecDef
					}
				}
			}
		}
	}

	return parsedSchemas, nil
}

func (pkgs *PackagesDefinitions) findTypeSpec(pkgPath string, typeName string) *TypeSpecDef {
	if pkgs.packages == nil {
		return nil
	}
	pd, found := pkgs.packages[pkgPath]
	if found {
		typeSpec, ok := pd.TypeDefinitions[typeName]
		if ok {
			return typeSpec
		}
	}

	return nil
}

// findPackagePathFromImports finds out the package path of a package via ranging imports of a ast.File
// @pkg the name of the target package
// @file current ast.File in which to search imports
// @return the package path of a package of @pkg.
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
			}
			if imp.Name.Name == "_" {
				hasAnonymousPkg = true
			}

			continue
		}
		if pkgs.packages != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			pd, ok := pkgs.packages[path]
			if ok {
				if pd.Name == pkg {
					return path
				}
			}
		}
	}

	// match unnamed package
	if hasAnonymousPkg && pkgs.packages != nil {
		for _, imp := range file.Imports {
			if imp.Name == nil {
				continue
			}
			if imp.Name.Name == "_" {
				path := strings.Trim(imp.Path.Value, `"`)
				pd, ok := pkgs.packages[path]
				if ok {
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
// @pkgPath the package path of @file.
func (pkgs *PackagesDefinitions) FindTypeSpec(typeName string, file *ast.File) *TypeSpecDef {
	if IsGolangPrimitiveType(typeName) {
		return nil
	}
	if file == nil { // for test
		return pkgs.uniqueDefinitions[typeName]
	}

	parts := strings.Split(typeName, ".")
	if len(parts) > 1 {
		isAliasPkgName := func(file *ast.File, pkgName string) bool {
			if file != nil && file.Imports != nil {
				for _, pkg := range file.Imports {
					if pkg.Name != nil && pkg.Name.Name == pkgName {
						return true
					}
				}
			}

			return false
		}

		if !isAliasPkgName(file, parts[0]) {
			typeDef, ok := pkgs.uniqueDefinitions[typeName]
			if ok {
				return typeDef
			}
		}

		pkgPath := pkgs.findPackagePathFromImports(parts[0], file)
		if len(pkgPath) == 0 && parts[0] == file.Name.Name {
			pkgPath = pkgs.files[file].PackagePath
		}

		return pkgs.findTypeSpec(pkgPath, parts[1])
	}

	typeDef, ok := pkgs.uniqueDefinitions[fullTypeName(file.Name.Name, typeName)]
	if ok {
		return typeDef
	}

	typeDef = pkgs.findTypeSpec(pkgs.files[file].PackagePath, typeName)
	if typeDef != nil {
		return typeDef
	}

	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == "." {
			typeDef := pkgs.findTypeSpec(strings.Trim(imp.Path.Value, `"`), typeName)
			if typeDef != nil {
				return typeDef
			}
		}
	}

	return nil
}
