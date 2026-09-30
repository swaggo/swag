package swag

import (
	"context"
	"errors"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListPackages(t *testing.T) {

	cases := []struct {
		name      string
		args      []string
		searchDir string
		except    error
	}{
		{
			name:      "errorArgs",
			args:      []string{"-abc"},
			searchDir: "testdata/golist",
			except:    fmt.Errorf("exit status 2"),
		},
		{
			name:      "normal",
			args:      []string{"-deps"},
			searchDir: "testdata/golist",
			except:    nil,
		},
		{
			name:      "list error",
			args:      []string{"-deps"},
			searchDir: "testdata/golist_not_exist",
			except:    errors.New("searchDir not exist"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := listOnePackages(context.TODO(), c.searchDir, nil, c.args...)
			if c.except != nil {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetAllGoFileInfoFromDepsByList(t *testing.T) {
	p := New(ParseUsingGoList(true))
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	cases := []struct {
		name           string
		buildPackage   *build.Package
		ignoreInternal bool
		except         error
	}{
		{
			name: "normal",
			buildPackage: &build.Package{
				Name:       "main",
				ImportPath: "github.com/swaggo/swag/testdata/golist",
				Dir:        "testdata/golist",
				GoFiles:    []string{"main.go"},
				CgoFiles:   []string{"api/api.go"},
			},
			except: nil,
		},
		{
			name: "ignore internal",
			buildPackage: &build.Package{
				Goroot: true,
			},
			ignoreInternal: true,
			except:         nil,
		},
		{
			name: "gofiles error",
			buildPackage: &build.Package{
				Dir:     "testdata/golist_not_exist",
				GoFiles: []string{"main.go"},
			},
			except: errors.New("file not exist"),
		},
		{
			name: "cgofiles error",
			buildPackage: &build.Package{
				Dir:      "testdata/golist_not_exist",
				CgoFiles: []string{"main.go"},
			},
			except: errors.New("file not exist"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.ignoreInternal {
				p.ParseInternal = false
			}
			c.buildPackage.Dir = filepath.Join(pwd, c.buildPackage.Dir)
			err := p.getAllGoFileInfoFromDepsByList(c.buildPackage, ParseModels)
			if c.except != nil {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestOperationParsePackages(t *testing.T) {
	root := &build.Package{ImportPath: "example.com/app", Dir: "/app", Imports: []string{"example.com/direct"}}
	direct := &build.Package{ImportPath: "example.com/direct", Dir: "/app/direct", Imports: []string{"example.com/transitive"}}
	transitive := &build.Package{ImportPath: "example.com/transitive", Dir: "/app/transitive"}
	pkgs := []*build.Package{root, direct, transitive}
	rootDirs := []string{"/app"}

	depth1 := operationParsePackages(pkgs, rootDirs, 1)
	assert.Contains(t, depth1, "example.com/app")
	assert.Contains(t, depth1, "example.com/direct")
	assert.NotContains(t, depth1, "example.com/transitive")

	depth2 := operationParsePackages(pkgs, rootDirs, 2)
	assert.Contains(t, depth2, "example.com/transitive")

	unlimited := operationParsePackages(pkgs, rootDirs, 0)
	assert.Len(t, unlimited, len(pkgs))
}
