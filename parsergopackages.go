package swag

import (
	"go/token"
	"os"
	"path/filepath"
	"slices"

	"golang.org/x/tools/go/packages"
)

func (parser *Parser) loadPackagesAndDeps(searchDirs []string, absMainAPIFilePath string) error {
	mode := packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
		packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo
	if parser.ParseDependency > 0 {
		mode |= packages.NeedDeps
	}

	absDirs := make([]string, 0, len(searchDirs)+1)
	absDirs = append(absDirs, filepath.Dir(absMainAPIFilePath))
	for _, dir := range searchDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		// load all subpackages keep the same logic with Parser.getAllGoFileInfo
		absDirs = append(absDirs, absDir+"/...")
	}

	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Mode: mode,
		Fset: fset,
	}, absDirs...)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			return e
		}
	}

	err = parser.walkPackages(pkgs, func(pkg *packages.Package) error {
		parseFlag := ParseFlag(ParseAll)
		if !slices.Contains(pkgs, pkg) {
			parseFlag = parser.ParseDependency
		}
		for i, file := range pkg.CompiledGoFiles {
			// TODO handle vendor?
			fileInfo, err := os.Stat(file)
			if err != nil {
				return err
			}
			if parser.Skip(file, fileInfo) != nil {
				continue
			}
			if err = parser.packages.CollectAstFile(fset, pkg.PkgPath, file, pkg.Syntax[i], parseFlag); err != nil {
				return err
			}
		}
		return nil
	})

	parser.packages.AddPackages(pkgs)
	return err
}

func (parser *Parser) walkPackages(pkgs []*packages.Package, f func(p *packages.Package) error) error {
	pkgSeen := make(map[string]struct{})
	return parser.walkPackagesInternal(pkgs, f, pkgSeen)
}

func (parser *Parser) walkPackagesInternal(pkgs []*packages.Package, f func(p *packages.Package) error,
	pkgSeen map[string]struct{}) error {
	for _, pkg := range pkgs {
		if parser.skipPackageByPrefix(pkg.PkgPath) {
			continue
		}
		if _, ok := pkgSeen[pkg.PkgPath]; ok {
			continue
		}
		pkgSeen[pkg.PkgPath] = struct{}{}

		if err := f(pkg); err != nil {
			return err
		}

		if parser.ParseDependency > 0 {
			imports := make([]*packages.Package, 0, len(pkg.Imports))
			for _, dep := range pkg.Imports {
				imports = append(imports, dep)
			}
			if err := parser.walkPackagesInternal(imports, f, pkgSeen); err != nil {
				return err
			}
		}
	}
	return nil
}
