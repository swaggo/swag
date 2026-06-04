package swag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/build"
	"os/exec"
	"path/filepath"
	"slices"
)

func listPackages(ctx context.Context, dirs []string, env []string, args ...string) ([]*build.Package, error) {
	pkgMap := make(map[string]*build.Package)
	for i, dir := range dirs {
		pkgs, err := listOnePackages(ctx, dir, env, args...)
		if err != nil {
			if i == 0 {
				return nil, fmt.Errorf("pkg %s cannot find all dependencies, %s", dir, err)
			}
			continue // ignore search dir load error?
		}
		for _, pkg := range pkgs {
			pkgMap[pkg.Dir] = pkg
		}
	}
	pkgs := make([]*build.Package, 0, len(pkgMap))
	for _, pkg := range pkgMap {
		pkgs = append(pkgs, pkg)
	}
	slices.SortFunc(pkgs, func(a, b *build.Package) int {
		if a.Dir < b.Dir {
			return -1
		} else if a.Dir > b.Dir {
			return 1
		}
		return 0
	})
	return pkgs, nil
}

func listOnePackages(ctx context.Context, dir string, env []string, args ...string) (pkgs []*build.Package, finalErr error) {
	cmd := exec.CommandContext(ctx, "go", append([]string{"list", "-json", "-e"}, args...)...)
	cmd.Env = env
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	defer func() {
		if (finalErr != nil) && (stderrBuf.Len() > 0) {
			finalErr = fmt.Errorf("%v\n%s", finalErr, stderrBuf.Bytes())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(stdout)
	for dec.More() {
		var pkg build.Package
		err = dec.Decode(&pkg)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, &pkg)
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}
	return pkgs, nil
}

func filterPackagesByParseDepth(pkgs []*build.Package, rootDirs []string, parseDepth int) []*build.Package {
	if parseDepth == 0 {
		return pkgs
	}

	pkgsByImportPath := make(map[string]*build.Package, len(pkgs))
	pkgsByDir := make(map[string]*build.Package, len(pkgs))
	for _, pkg := range pkgs {
		if pkg.ImportPath != "" {
			pkgsByImportPath[pkg.ImportPath] = pkg
		}

		absDir, err := filepath.Abs(pkg.Dir)
		if err == nil {
			pkgsByDir[absDir] = pkg
		}
	}

	type packageAtDepth struct {
		pkg   *build.Package
		depth int
	}

	queue := make([]packageAtDepth, 0)
	for _, rootDir := range rootDirs {
		absRootDir, err := filepath.Abs(rootDir)
		if err != nil {
			continue
		}

		rootPkg := pkgsByDir[absRootDir]
		if rootPkg == nil {
			continue
		}

		for _, importPath := range rootPkg.Imports {
			if pkg := pkgsByImportPath[importPath]; pkg != nil {
				queue = append(queue, packageAtDepth{pkg: pkg, depth: 1})
			}
		}
	}

	selected := make(map[string]struct{})
	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]

		key := packageListKey(next.pkg)
		if _, ok := selected[key]; ok {
			continue
		}
		selected[key] = struct{}{}

		if next.depth >= parseDepth {
			continue
		}

		for _, importPath := range next.pkg.Imports {
			if pkg := pkgsByImportPath[importPath]; pkg != nil {
				queue = append(queue, packageAtDepth{pkg: pkg, depth: next.depth + 1})
			}
		}
	}

	filteredPkgs := make([]*build.Package, 0, len(selected))
	for _, pkg := range pkgs {
		if _, ok := selected[packageListKey(pkg)]; ok {
			filteredPkgs = append(filteredPkgs, pkg)
		}
	}

	return filteredPkgs
}

func packageListKey(pkg *build.Package) string {
	if pkg.ImportPath != "" {
		return pkg.ImportPath
	}

	absDir, err := filepath.Abs(pkg.Dir)
	if err == nil {
		return absDir
	}

	return pkg.Dir
}

func (parser *Parser) getAllGoFileInfoFromDepsByList(pkg *build.Package, parseFlag ParseFlag) error {
	ignoreInternal := pkg.Goroot && !parser.ParseInternal
	if ignoreInternal { // ignored internal
		return nil
	}

	if parser.skipPackageByPrefix(pkg.ImportPath) {
		return nil // ignored by user-defined package path prefixes
	}

	srcDir := pkg.Dir
	var err error
	for i := range pkg.GoFiles {
		err = parser.parseFile(pkg.ImportPath, filepath.Join(srcDir, pkg.GoFiles[i]), nil, parseFlag)
		if err != nil {
			return err
		}
	}

	// parse .go source files that import "C"
	for i := range pkg.CgoFiles {
		err = parser.parseFile(pkg.ImportPath, filepath.Join(srcDir, pkg.CgoFiles[i]), nil, parseFlag)
		if err != nil {
			return err
		}
	}

	return nil
}
