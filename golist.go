package swag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/build"
	"os/exec"
	"path/filepath"
)

func listPackages(ctx context.Context, dir string, env []string, args ...string) (pkgs []*build.Package, finalErr error) {
	goArgs := append([]string{"list", "-json", "-e"}, args...)
	cmd := exec.CommandContext(ctx, "go", goArgs...)
	cmd.Env = env
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	defer func() {
		if stderrBuf.Len() > 0 {
			finalErr = fmt.Errorf("%v\n%s", finalErr, stderrBuf.Bytes())
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	dec := json.NewDecoder(stdout)
	for dec.More() {
		var pkg build.Package
		if err := dec.Decode(&pkg); err != nil {
			return nil, err
		}
		pkgs = append(pkgs, &pkg)
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return pkgs, nil
}

func (parser *Parser) getAllGoFileInfoFromDepsByList(pkg *build.Package) error {
	ignoreInternal := pkg.Goroot && !parser.ParseInternal
	if ignoreInternal { // ignored internal
		return nil
	}

	// Skip cgo
	if pkg.Name == "C" {
		return nil
	}

	srcDir := pkg.Dir
	for i := range pkg.GoFiles {
		path := filepath.Join(srcDir, pkg.GoFiles[i])
		if err := parser.parseFile(pkg.ImportPath, path, nil); err != nil {
			return err
		}
	}

	for i := range pkg.CFiles {
		path := filepath.Join(srcDir, pkg.CFiles[i])
		if err := parser.parseFile(pkg.ImportPath, path, nil); err != nil {
			return err
		}
	}

	return nil
}
