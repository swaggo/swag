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
func (pkgDef *PackagesDefinitions) CollectAstFile(packageDir, path string, astFile *ast.File) error {
	if pkgDef.files == nil {
		pkgDef.files = make(map[*ast.File]*AstFileInfo)
	}

	if pkgDef.packages == nil {
		pkgDef.packages = make(map[string]*PackageDefinitions)
	}

	// return without storing the file if we lack a packageDir
	if len(packageDir) == 0 {
		return nil
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	pd, ok := pkgDef.packages[packageDir]
	if ok {
		// return without storing the file if it already exists
		_, exists := pd.Files[path]
		if exists {
			return nil
		}
		pd.Files[path] = astFile
	} else {
		pkgDef.packages[packageDir] = &PackageDefinitions{
			Name:            astFile.Name.Name,
			Files:           map[string]*ast.File{path: astFile},
			TypeDefinitions: make(map[string]*TypeSpecDef),
		}
	}

	pkgDef.files[astFile] = &AstFileInfo{
		File:        astFile,
		Path:        path,
		PackagePath: packageDir,
	}

	return nil
}

// RangeFiles for range the collection of ast.File in alphabetic order.
func (pkgDef *PackagesDefinitions) RangeFiles(handle func(filename string, file *ast.File) error) error {
	sortedFiles := make([]*AstFileInfo, 0, len(pkgDef.files))
	for _, info := range pkgDef.files {
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
func (pkgDef *PackagesDefinitions) ParseTypes() (map[*TypeSpecDef]*Schema, error) {
	parsedSchemas := make(map[*TypeSpecDef]*Schema)
	for astFile, info := range pkgDef.files {
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

						if pkgDef.uniqueDefinitions == nil {
							pkgDef.uniqueDefinitions = make(map[string]*TypeSpecDef)
						}

						fullName := typeSpecDef.FullName()
						anotherTypeDef, ok := pkgDef.uniqueDefinitions[fullName]
						if ok {
							if typeSpecDef.PkgPath == anotherTypeDef.PkgPath {
								continue
							} else {
								delete(pkgDef.uniqueDefinitions, fullName)
							}
						} else {
							pkgDef.uniqueDefinitions[fullName] = typeSpecDef
						}

						pkgDef.packages[typeSpecDef.PkgPath].TypeDefinitions[typeSpecDef.Name()] = typeSpecDef
					}
				}
			}
		}
	}

	return parsedSchemas, nil
}

func (pkgDef *PackagesDefinitions) findTypeSpec(pkgPath string, typeName string) *TypeSpecDef {
	if pkgDef.packages == nil {
		return nil
	}
	pd, found := pkgDef.packages[pkgPath]
	if found {
		typeSpec, ok := pd.TypeDefinitions[typeName]
		if ok {
			return typeSpec
		}
	}

	return nil
}

// findPackagePathFromImports finds out the package path of a package via ranging imports of an ast.File
// @pkg the name of the target package
// @file current ast.File in which to search imports
// @return the package path of a package of @pkg.
func (pkgDef *PackagesDefinitions) findPackagePathFromImports(pkg string, file *ast.File) string {
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
		if pkgDef.packages != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			pd, ok := pkgDef.packages[path]
			if ok {
				if pd.Name == pkg {
					return path
				}
			}
		}
	}

	// match unnamed package
	if hasAnonymousPkg && pkgDef.packages != nil {
		for _, imp := range file.Imports {
			if imp.Name == nil {
				continue
			}
			if imp.Name.Name == "_" {
				path := strings.Trim(imp.Path.Value, `"`)
				pd, ok := pkgDef.packages[path]
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
// @typeName the name of the target type, if it starts with a package name, find its own package path from imports on top of file
// @file the ast.file in which @typeName is used
// @pkgPath the package path of @file.
func (pkgDef *PackagesDefinitions) FindTypeSpec(typeName string, file *ast.File) *TypeSpecDef {
	if IsGolangPrimitiveType(typeName) {
		return nil
	}
	if file == nil { // for test
		return pkgDef.uniqueDefinitions[typeName]
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
			typeDef, ok := pkgDef.uniqueDefinitions[typeName]
			if ok {
				return typeDef
			}
		}

		pkgPath := pkgDef.findPackagePathFromImports(parts[0], file)
		if len(pkgPath) == 0 && parts[0] == file.Name.Name {
			pkgPath = pkgDef.files[file].PackagePath
		}

		return pkgDef.findTypeSpec(pkgPath, parts[1])
	}

	typeDef, ok := pkgDef.uniqueDefinitions[fullTypeName(file.Name.Name, typeName)]
	if ok {
		return typeDef
	}

	typeDef = pkgDef.findTypeSpec(pkgDef.files[file].PackagePath, typeName)
	if typeDef != nil {
		return typeDef
	}

	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == "." {
			typeDef := pkgDef.findTypeSpec(strings.Trim(imp.Path.Value, `"`), typeName)
			if typeDef != nil {
				return typeDef
			}
		}
	}

	return nil
}
