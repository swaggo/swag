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

func operationParsePackages(pkgs []*build.Package, rootDirs []string, parseDepth int) map[string]struct{} {
	within := make(map[string]struct{}, len(pkgs))
	if parseDepth <= 0 {
		for _, pkg := range pkgs {
			within[pkg.ImportPath] = struct{}{}
		}
		return within
	}

	byImportPath := make(map[string]*build.Package, len(pkgs))
	for _, pkg := range pkgs {
		byImportPath[pkg.ImportPath] = pkg
	}

	rootDirSet := make(map[string]struct{}, len(rootDirs))
	for _, dir := range rootDirs {
		if abs, err := filepath.Abs(dir); err == nil {
			rootDirSet[abs] = struct{}{}
		}
	}

	type pkgAtDepth struct {
		importPath string
		depth      int
	}

	var queue []pkgAtDepth
	for _, pkg := range pkgs {
		abs, err := filepath.Abs(pkg.Dir)
		if err != nil {
			continue
		}
		if _, ok := rootDirSet[abs]; ok {
			if _, seen := within[pkg.ImportPath]; !seen {
				within[pkg.ImportPath] = struct{}{}
				queue = append(queue, pkgAtDepth{importPath: pkg.ImportPath, depth: 0})
			}
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.depth >= parseDepth {
			continue
		}
		pkg := byImportPath[current.importPath]
		if pkg == nil {
			continue
		}
		for _, importPath := range pkg.Imports {
			if _, ok := byImportPath[importPath]; !ok {
				continue
			}
			if _, seen := within[importPath]; seen {
				continue
			}
			within[importPath] = struct{}{}
			queue = append(queue, pkgAtDepth{importPath: importPath, depth: current.depth + 1})
		}
	}

	return within
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
