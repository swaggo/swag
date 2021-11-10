package swag

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/loader"
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
func (pkg *PackagesDefinitions) CollectAstFile(packageDir, path string, astFile *ast.File) error {
	if pkg.files == nil {
		pkg.files = make(map[*ast.File]*AstFileInfo)
	}

	if pkg.packages == nil {
		pkg.packages = make(map[string]*PackageDefinitions)
	}

	// return without storing the file if we lack a packageDir
	if len(packageDir) == 0 {
		return nil
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	pd, ok := pkg.packages[packageDir]
	if ok {
		// return without storing the file if it already exists
		_, exists := pd.Files[path]
		if exists {
			return nil
		}
		pd.Files[path] = astFile
	} else {
		pkg.packages[packageDir] = &PackageDefinitions{
			Name:            astFile.Name.Name,
			Files:           map[string]*ast.File{path: astFile},
			TypeDefinitions: make(map[string]*TypeSpecDef),
		}
	}

	pkg.files[astFile] = &AstFileInfo{
		File:        astFile,
		Path:        path,
		PackagePath: packageDir,
	}

	return nil
}

// RangeFiles for range the collection of ast.File in alphabetic order.
func (pkg *PackagesDefinitions) RangeFiles(handle func(filename string, file *ast.File) error) error {
	sortedFiles := make([]*AstFileInfo, 0, len(pkg.files))
	for _, info := range pkg.files {
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
func (pkg *PackagesDefinitions) ParseTypes() (map[*TypeSpecDef]*Schema, error) {
	parsedSchemas := make(map[*TypeSpecDef]*Schema)
	for astFile, info := range pkg.files {
		pkg.parseTypesFromFile(astFile, info.PackagePath, parsedSchemas)
	}
	return parsedSchemas, nil
}

func (pkg *PackagesDefinitions) parseTypesFromFile(astFile *ast.File, packagePath string, parsedSchemas map[*TypeSpecDef]*Schema) {
	for _, astDeclaration := range astFile.Decls {
		if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
			for _, astSpec := range generalDeclaration.Specs {
				if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
					typeSpecDef := &TypeSpecDef{
						PkgPath:  packagePath,
						File:     astFile,
						TypeSpec: typeSpec,
					}

					if idt, ok := typeSpec.Type.(*ast.Ident); ok && IsGolangPrimitiveType(idt.Name) && parsedSchemas != nil {
						parsedSchemas[typeSpecDef] = &Schema{
							PkgPath: typeSpecDef.PkgPath,
							Name:    astFile.Name.Name,
							Schema:  PrimitiveSchema(TransToValidSchemeType(idt.Name)),
						}
					}

					if pkg.uniqueDefinitions == nil {
						pkg.uniqueDefinitions = make(map[string]*TypeSpecDef)
					}

					fullName := typeSpecDef.FullName()
					anotherTypeDef, ok := pkg.uniqueDefinitions[fullName]
					if ok {
						if typeSpecDef.PkgPath == anotherTypeDef.PkgPath {
							continue
						} else {
							delete(pkg.uniqueDefinitions, fullName)
						}
					} else {
						pkg.uniqueDefinitions[fullName] = typeSpecDef
					}

					if pkg.packages[typeSpecDef.PkgPath] == nil {
						pkg.packages[typeSpecDef.PkgPath] = &PackageDefinitions{
							Name:            astFile.Name.Name,
							TypeDefinitions: map[string]*TypeSpecDef{typeSpecDef.Name(): typeSpecDef},
						}
					} else if _, ok = pkg.packages[typeSpecDef.PkgPath].TypeDefinitions[typeSpecDef.Name()]; !ok {
						pkg.packages[typeSpecDef.PkgPath].TypeDefinitions[typeSpecDef.Name()] = typeSpecDef
					}
				}
			}
		}
	}
}

func (pkg *PackagesDefinitions) findTypeSpec(pkgPath string, typeName string) *TypeSpecDef {
	if pkg.packages == nil {
		return nil
	}
	pd, found := pkg.packages[pkgPath]
	if found {
		typeSpec, ok := pd.TypeDefinitions[typeName]
		if ok {
			return typeSpec
		}
	}

	return nil
}

func (pkg *PackagesDefinitions) loadExternalPackage(importPath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	conf := loader.Config{
		ParserMode: goparser.ParseComments,
		Cwd:        cwd,
	}

	conf.Import(importPath)

	loaderProgram, err := conf.Load()
	if err != nil {
		return err
	}

	for _, info := range loaderProgram.AllPackages {
		pkgPath := strings.TrimPrefix(info.Pkg.Path(), "vendor/")
		for _, astFile := range info.Files {
			pkg.parseTypesFromFile(astFile, pkgPath, nil)
		}
	}

	return nil
}

// findPackagePathFromImports finds out the package path of a package via ranging imports of an ast.File
// @pkg the name of the target package
// @file current ast.File in which to search imports
// @fuzzy search for the package path that the last part matches the @pkg if true
// @return the package path of a package of @pkg.
func (pkg *PackagesDefinitions) findPackagePathFromImports(p string, file *ast.File, fuzzy bool) string {
	if file == nil {
		return ""
	}

	if strings.ContainsRune(p, '.') {
		p = strings.Split(p, ".")[0]
	}

	hasAnonymousPkg := false

	matchLastPathPart := func(pkgPath string) bool {
		paths := strings.Split(pkgPath, "/")
		return paths[len(paths)-1] == p
	}

	// prior to match named package
	for _, imp := range file.Imports {
		if imp.Name != nil {
			if imp.Name.Name == p {
				return strings.Trim(imp.Path.Value, `"`)
			}
			if imp.Name.Name == "_" {
				hasAnonymousPkg = true
			}

			continue
		}
		if pkg.packages != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			if fuzzy {
				if matchLastPathPart(path) {
					return path
				}
			} else if pd, ok := pkg.packages[path]; ok && pd.Name == p {
				return path
			}
		}
	}

	// match unnamed package
	if hasAnonymousPkg && pkg.packages != nil {
		for _, imp := range file.Imports {
			if imp.Name == nil {
				continue
			}
			if imp.Name.Name == "_" {
				path := strings.Trim(imp.Path.Value, `"`)
				if fuzzy {
					if matchLastPathPart(path) {
						return path
					}
				} else if pd, ok := pkg.packages[path]; ok && pd.Name == p {
					return path
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
func (pkg *PackagesDefinitions) FindTypeSpec(typeName string, file *ast.File, parseDependency bool) *TypeSpecDef {
	if IsGolangPrimitiveType(typeName) {
		return nil
	}
	if file == nil { // for test
		return pkg.uniqueDefinitions[typeName]
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
			typeDef, ok := pkg.uniqueDefinitions[typeName]
			if ok {
				return typeDef
			}
		}
		pkgPath := pkg.findPackagePathFromImports(parts[0], file, false)
		if len(pkgPath) == 0 {
			// check if the current package
			if parts[0] == file.Name.Name {
				pkgPath = pkg.files[file].PackagePath
			} else if parseDependency {
				// take it as an external package, needs to be loaded
				if pkgPath = pkg.findPackagePathFromImports(parts[0], file, true); len(pkgPath) > 0 {
					if err := pkg.loadExternalPackage(pkgPath); err != nil {
						return nil
					}
				}
			}
		}

		return pkg.findTypeSpec(pkgPath, parts[1])
	}

	typeDef, ok := pkg.uniqueDefinitions[fullTypeName(file.Name.Name, typeName)]
	if ok {
		return typeDef
	}

	typeDef = pkg.findTypeSpec(pkg.files[file].PackagePath, typeName)
	if typeDef != nil {
		return typeDef
	}

	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == "." {
			typeDef := pkg.findTypeSpec(strings.Trim(imp.Path.Value, `"`), typeName)
			if typeDef != nil {
				return typeDef
			}
		}
	}

	return nil
}
