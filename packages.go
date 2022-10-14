package swag

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
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
func (pkgDefs *PackagesDefinitions) CollectAstFile(packageDir, path string, astFile *ast.File) error {
	if pkgDefs.files == nil {
		pkgDefs.files = make(map[*ast.File]*AstFileInfo)
	}

	if pkgDefs.packages == nil {
		pkgDefs.packages = make(map[string]*PackageDefinitions)
	}

	// return without storing the file if we lack a packageDir
	if packageDir == "" {
		return nil
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	dependency, ok := pkgDefs.packages[packageDir]
	if ok {
		// return without storing the file if it already exists
		_, exists := dependency.Files[path]
		if exists {
			return nil
		}

		dependency.Files[path] = astFile
	} else {
		pkgDefs.packages[packageDir] = &PackageDefinitions{
			Name:            astFile.Name.Name,
			Files:           map[string]*ast.File{path: astFile},
			TypeDefinitions: make(map[string]*TypeSpecDef),
		}
	}

	pkgDefs.files[astFile] = &AstFileInfo{
		File:        astFile,
		Path:        path,
		PackagePath: packageDir,
	}

	return nil
}

// RangeFiles for range the collection of ast.File in alphabetic order.
func rangeFiles(files map[*ast.File]*AstFileInfo, handle func(filename string, file *ast.File) error) error {
	sortedFiles := make([]*AstFileInfo, 0, len(files))
	for _, info := range files {
		// ignore package path prefix with 'vendor' or $GOROOT,
		// because the router info of api will not be included these files.
		if strings.HasPrefix(info.PackagePath, "vendor") || strings.HasPrefix(info.Path, runtime.GOROOT()) {
			continue
		}
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
func (pkgDefs *PackagesDefinitions) ParseTypes() (map[*TypeSpecDef]*Schema, error) {
	parsedSchemas := make(map[*TypeSpecDef]*Schema)
	for astFile, info := range pkgDefs.files {
		pkgDefs.parseTypesFromFile(astFile, info.PackagePath, parsedSchemas)
		pkgDefs.parseFunctionScopedTypesFromFile(astFile, info.PackagePath, parsedSchemas)
	}
	return parsedSchemas, nil
}

func (pkgDefs *PackagesDefinitions) parseTypesFromFile(astFile *ast.File, packagePath string, parsedSchemas map[*TypeSpecDef]*Schema) {
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

					if pkgDefs.uniqueDefinitions == nil {
						pkgDefs.uniqueDefinitions = make(map[string]*TypeSpecDef)
					}

					fullName := typeSpecFullName(typeSpecDef)

					anotherTypeDef, ok := pkgDefs.uniqueDefinitions[fullName]
					if ok {
						if typeSpecDef.PkgPath == anotherTypeDef.PkgPath {
							continue
						} else {
							delete(pkgDefs.uniqueDefinitions, fullName)
						}
					} else {
						pkgDefs.uniqueDefinitions[fullName] = typeSpecDef
					}

					if pkgDefs.packages[typeSpecDef.PkgPath] == nil {
						pkgDefs.packages[typeSpecDef.PkgPath] = &PackageDefinitions{
							Name:            astFile.Name.Name,
							TypeDefinitions: map[string]*TypeSpecDef{typeSpecDef.Name(): typeSpecDef},
						}
					} else if _, ok = pkgDefs.packages[typeSpecDef.PkgPath].TypeDefinitions[typeSpecDef.Name()]; !ok {
						pkgDefs.packages[typeSpecDef.PkgPath].TypeDefinitions[typeSpecDef.Name()] = typeSpecDef
					}
				}
			}
		}
	}
}

func (pkgDefs *PackagesDefinitions) parseFunctionScopedTypesFromFile(astFile *ast.File, packagePath string, parsedSchemas map[*TypeSpecDef]*Schema) {
	for _, astDeclaration := range astFile.Decls {
		funcDeclaration, ok := astDeclaration.(*ast.FuncDecl)
		if ok && funcDeclaration.Body != nil {
			for _, stmt := range funcDeclaration.Body.List {
				if declStmt, ok := (stmt).(*ast.DeclStmt); ok {
					if genDecl, ok := (declStmt.Decl).(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
						for _, astSpec := range genDecl.Specs {
							if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
								typeSpecDef := &TypeSpecDef{
									PkgPath:    packagePath,
									File:       astFile,
									TypeSpec:   typeSpec,
									ParentSpec: astDeclaration,
								}

								if idt, ok := typeSpec.Type.(*ast.Ident); ok && IsGolangPrimitiveType(idt.Name) && parsedSchemas != nil {
									parsedSchemas[typeSpecDef] = &Schema{
										PkgPath: typeSpecDef.PkgPath,
										Name:    astFile.Name.Name,
										Schema:  PrimitiveSchema(TransToValidSchemeType(idt.Name)),
									}
								}

								if pkgDefs.uniqueDefinitions == nil {
									pkgDefs.uniqueDefinitions = make(map[string]*TypeSpecDef)
								}

								fullName := typeSpecFullName(typeSpecDef)

								anotherTypeDef, ok := pkgDefs.uniqueDefinitions[fullName]
								if ok {
									if typeSpecDef.PkgPath == anotherTypeDef.PkgPath {
										continue
									} else {
										delete(pkgDefs.uniqueDefinitions, fullName)
									}
								} else {
									pkgDefs.uniqueDefinitions[fullName] = typeSpecDef
								}

								if pkgDefs.packages[typeSpecDef.PkgPath] == nil {
									pkgDefs.packages[typeSpecDef.PkgPath] = &PackageDefinitions{
										Name:            astFile.Name.Name,
										TypeDefinitions: map[string]*TypeSpecDef{fullName: typeSpecDef},
									}
								} else if _, ok = pkgDefs.packages[typeSpecDef.PkgPath].TypeDefinitions[fullName]; !ok {
									pkgDefs.packages[typeSpecDef.PkgPath].TypeDefinitions[fullName] = typeSpecDef
								}
							}
						}

					}
				}
			}
		}
	}
}

func (pkgDefs *PackagesDefinitions) findTypeSpec(pkgPath string, typeName string) *TypeSpecDef {
	if pkgDefs.packages == nil {
		return nil
	}

	pd, found := pkgDefs.packages[pkgPath]
	if found {
		typeSpec, ok := pd.TypeDefinitions[typeName]
		if ok {
			return typeSpec
		}
	}

	return nil
}

func (pkgDefs *PackagesDefinitions) loadExternalPackage(importPath string) error {
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
			pkgDefs.parseTypesFromFile(astFile, pkgPath, nil)
		}
	}

	return nil
}

// findPackagePathFromImports finds out the package path of a package via ranging imports of an ast.File
// @pkg the name of the target package
// @file current ast.File in which to search imports
// @fuzzy search for the package path that the last part matches the @pkg if true
// @return the package path of a package of @pkg.
func (pkgDefs *PackagesDefinitions) findPackagePathFromImports(pkg string, file *ast.File, fuzzy bool) string {
	if file == nil {
		return ""
	}

	if strings.ContainsRune(pkg, '.') {
		pkg = strings.Split(pkg, ".")[0]
	}

	hasAnonymousPkg := false

	matchLastPathPart := func(pkgPath string) bool {
		paths := strings.Split(pkgPath, "/")

		return paths[len(paths)-1] == pkg
	}

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

		if pkgDefs.packages != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			if fuzzy {
				if matchLastPathPart(path) {
					return path
				}

				continue
			}

			pd, ok := pkgDefs.packages[path]
			if ok && pd.Name == pkg {
				return path
			}
		}
	}

	// match unnamed package
	if hasAnonymousPkg && pkgDefs.packages != nil {
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
				} else if pd, ok := pkgDefs.packages[path]; ok && pd.Name == pkg {
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
func (pkgDefs *PackagesDefinitions) FindTypeSpec(typeName string, file *ast.File, parseDependency bool) *TypeSpecDef {
	if IsGolangPrimitiveType(typeName) {
		return nil
	}

	if file == nil { // for test
		return pkgDefs.uniqueDefinitions[typeName]
	}

	parts := strings.Split(strings.Split(typeName, "[")[0], ".")
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
			typeDef, ok := pkgDefs.uniqueDefinitions[typeName]
			if ok {
				return typeDef
			}
		}

		pkgPath := pkgDefs.findPackagePathFromImports(parts[0], file, false)
		if len(pkgPath) == 0 {
			// check if the current package
			if parts[0] == file.Name.Name {
				pkgPath = pkgDefs.files[file].PackagePath
			} else if parseDependency {
				// take it as an external package, needs to be loaded
				if pkgPath = pkgDefs.findPackagePathFromImports(parts[0], file, true); len(pkgPath) > 0 {
					if err := pkgDefs.loadExternalPackage(pkgPath); err != nil {
						return nil
					}
				}
			}
		}

		if def := pkgDefs.findGenericTypeSpec(typeName, file, parseDependency); def != nil {
			return def
		}

		return pkgDefs.findTypeSpec(pkgPath, parts[1])
	}

	if def := pkgDefs.findGenericTypeSpec(fullTypeName(file.Name.Name, typeName), file, parseDependency); def != nil {
		return def
	}

	typeDef, ok := pkgDefs.uniqueDefinitions[fullTypeName(file.Name.Name, typeName)]
	if ok {
		return typeDef
	}

	typeDef = pkgDefs.findTypeSpec(pkgDefs.files[file].PackagePath, typeName)
	if typeDef != nil {
		return typeDef
	}

	for _, imp := range file.Imports {
		if imp.Name != nil && imp.Name.Name == "." {
			typeDef := pkgDefs.findTypeSpec(strings.Trim(imp.Path.Value, `"`), typeName)
			if typeDef != nil {
				return typeDef
			}
		}
	}

	return nil
}

func (pkgDefs *PackagesDefinitions) findGenericTypeSpec(typeName string, file *ast.File, parseDependency bool) *TypeSpecDef {
	if strings.Contains(typeName, "[") {
		// genericName differs from typeName in that it does not contain any type parameters
		genericName := strings.SplitN(typeName, "[", 2)[0]
		for tName, tSpec := range pkgDefs.uniqueDefinitions {
			if !strings.Contains(tName, "[") {
				continue
			}

			if strings.Contains(tName, genericName) {
				if parametrized := pkgDefs.parametrizeGenericType(file, tSpec, typeName, parseDependency); parametrized != nil {
					return parametrized
				}
			}
		}
	}
	return nil
}
