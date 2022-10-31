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
	pkgDefs.removeAllNotUniqueTypes()
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

					fullName := typeSpecDef.TypeName()

					anotherTypeDef, ok := pkgDefs.uniqueDefinitions[fullName]
					if ok {
						if anotherTypeDef == nil {
							typeSpecDef.NotUnique = true
							pkgDefs.uniqueDefinitions[typeSpecDef.TypeName()] = typeSpecDef
						} else if typeSpecDef.PkgPath != anotherTypeDef.PkgPath {
							anotherTypeDef.NotUnique = true
							typeSpecDef.NotUnique = true
							pkgDefs.uniqueDefinitions[fullName] = nil
							pkgDefs.uniqueDefinitions[anotherTypeDef.TypeName()] = anotherTypeDef
							pkgDefs.uniqueDefinitions[typeSpecDef.TypeName()] = typeSpecDef
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

								fullName := typeSpecDef.TypeName()

								anotherTypeDef, ok := pkgDefs.uniqueDefinitions[fullName]
								if ok {
									if anotherTypeDef == nil {
										typeSpecDef.NotUnique = true
										pkgDefs.uniqueDefinitions[typeSpecDef.TypeName()] = typeSpecDef
									} else if typeSpecDef.PkgPath != anotherTypeDef.PkgPath {
										anotherTypeDef.NotUnique = true
										typeSpecDef.NotUnique = true
										pkgDefs.uniqueDefinitions[fullName] = nil
										pkgDefs.uniqueDefinitions[anotherTypeDef.TypeName()] = anotherTypeDef
										pkgDefs.uniqueDefinitions[typeSpecDef.TypeName()] = typeSpecDef
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

func (pkgDefs *PackagesDefinitions) removeAllNotUniqueTypes() {
	for key, ud := range pkgDefs.uniqueDefinitions {
		if ud == nil {
			delete(pkgDefs.uniqueDefinitions, key)
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
// @return the package paths of a package of @pkg.
func (pkgDefs *PackagesDefinitions) findPackagePathFromImports(pkg string, file *ast.File) (matchedPkgPaths, externalPkgPaths []string) {
	if file == nil {
		return
	}

	if strings.ContainsRune(pkg, '.') {
		pkg = strings.Split(pkg, ".")[0]
	}

	matchLastPathPart := func(pkgPath string) bool {
		paths := strings.Split(pkgPath, "/")
		return paths[len(paths)-1] == pkg
	}

	// prior to match named package
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if imp.Name != nil {
			if imp.Name.Name == pkg {
				// if name match, break loop and return
				_, ok := pkgDefs.packages[path]
				if ok {
					matchedPkgPaths = []string{path}
					externalPkgPaths = nil
				} else {
					externalPkgPaths = []string{path}
					matchedPkgPaths = nil
				}
				break
			} else if imp.Name.Name == "_" && len(pkg) > 0 {
				//for unused types
				pd, ok := pkgDefs.packages[path]
				if ok {
					if pd.Name == pkg {
						matchedPkgPaths = append(matchedPkgPaths, path)
					}
				} else if matchLastPathPart(path) {
					externalPkgPaths = append(externalPkgPaths, path)
				}
			} else if imp.Name.Name == "." && len(pkg) == 0 {
				_, ok := pkgDefs.packages[path]
				if ok {
					matchedPkgPaths = append(matchedPkgPaths, path)
				} else if len(pkg) == 0 || matchLastPathPart(path) {
					externalPkgPaths = append(externalPkgPaths, path)
				}
			}
		} else if pkgDefs.packages != nil && len(pkg) > 0 {
			pd, ok := pkgDefs.packages[path]
			if ok {
				if pd.Name == pkg {
					matchedPkgPaths = append(matchedPkgPaths, path)
				}
			} else if matchLastPathPart(path) {
				externalPkgPaths = append(externalPkgPaths, path)
			}
		}
	}

	if len(pkg) == 0 || file.Name.Name == pkg {
		matchedPkgPaths = append(matchedPkgPaths, pkgDefs.files[file].PackagePath)
	}

	return
}

func (pkgDefs *PackagesDefinitions) findTypeSpecFromPackagePaths(matchedPkgPaths, externalPkgPaths []string, name string, parseDependency bool) (typeDef *TypeSpecDef) {
	for _, pkgPath := range matchedPkgPaths {
		typeDef = pkgDefs.findTypeSpec(pkgPath, name)
		if typeDef != nil {
			return typeDef
		}
	}

	if parseDependency {
		for _, pkgPath := range externalPkgPaths {
			if err := pkgDefs.loadExternalPackage(pkgPath); err == nil {
				typeDef = pkgDefs.findTypeSpec(pkgPath, name)
				if typeDef != nil {
					return typeDef
				}
			}
		}
	}

	return typeDef
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
		typeDef, ok := pkgDefs.uniqueDefinitions[typeName]
		if ok {
			return typeDef
		}

		pkgPaths, externalPkgPaths := pkgDefs.findPackagePathFromImports(parts[0], file)
		typeDef = pkgDefs.findTypeSpecFromPackagePaths(pkgPaths, externalPkgPaths, parts[1], parseDependency)
		return pkgDefs.parametrizeGenericType(file, typeDef, typeName, parseDependency)
	}

	typeDef, ok := pkgDefs.uniqueDefinitions[fullTypeName(file.Name.Name, typeName)]
	if ok {
		return typeDef
	}

	//in case that comment //@name renamed the type with a name without a dot
	typeDef, ok = pkgDefs.uniqueDefinitions[typeName]
	if ok {
		return typeDef
	}

	name := parts[0]
	typeDef, ok = pkgDefs.uniqueDefinitions[fullTypeName(file.Name.Name, name)]
	if !ok {
		pkgPaths, externalPkgPaths := pkgDefs.findPackagePathFromImports("", file)
		typeDef = pkgDefs.findTypeSpecFromPackagePaths(pkgPaths, externalPkgPaths, name, parseDependency)
	}
	return pkgDefs.parametrizeGenericType(file, typeDef, typeName, parseDependency)
}
