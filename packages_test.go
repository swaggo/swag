package swag

import (
	"go/ast"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackagesDefinitions_CollectAstFile(t *testing.T) {
	pd := PackagesDefinitions{}
	assert.NoError(t, pd.CollectAstFile("", "", nil))

	firstFile := &ast.File{
		Name: &ast.Ident{Name: "main.go"},
	}

	packageDir := "github.com/swaggo/swag/testdata/simple"
	assert.NoError(t, pd.CollectAstFile(packageDir, "testdata/simple/"+firstFile.Name.String(), firstFile))
	assert.NotEmpty(t, pd.packages[packageDir])

	absPath, _ := filepath.Abs("testdata/simple/" + firstFile.Name.String())
	astFileInfo := &AstFileInfo{
		File:        firstFile,
		Path:        absPath,
		PackagePath: packageDir,
	}
	assert.Equal(t, pd.files[firstFile], astFileInfo)

	// Override
	assert.NoError(t, pd.CollectAstFile(packageDir, "testdata/simple/"+firstFile.Name.String(), firstFile))
	assert.Equal(t, pd.files[firstFile], astFileInfo)

	// Another file
	secondFile := &ast.File{
		Name: &ast.Ident{Name: "api.go"},
	}
	assert.NoError(t, pd.CollectAstFile(packageDir, "testdata/simple/"+secondFile.Name.String(), secondFile))
}

func TestPackagesDefinitions_RangeFiles(t *testing.T) {
	pd := PackagesDefinitions{
		files: map[*ast.File]*AstFileInfo{
			{
				Name: &ast.Ident{Name: "main.go"},
			}: {
				File:        &ast.File{Name: &ast.Ident{Name: "main.go"}},
				Path:        "testdata/simple/main.go",
				PackagePath: "main",
			},
			{
				Name: &ast.Ident{Name: "api.go"},
			}: {
				File:        &ast.File{Name: &ast.Ident{Name: "api.go"}},
				Path:        "testdata/simple/api/api.go",
				PackagePath: "api",
			},
		},
	}

	i, expect := 0, []string{"testdata/simple/api/api.go", "testdata/simple/main.go"}
	_ = rangeFiles(pd.files, func(filename string, file *ast.File) error {
		assert.Equal(t, expect[i], filename)
		i++
		return nil
	})
}

func TestPackagesDefinitions_findTypeSpec(t *testing.T) {
	pd := PackagesDefinitions{}
	var nilTypeSpec *TypeSpecDef
	assert.Equal(t, nilTypeSpec, pd.findTypeSpec("model", "User"))

	userTypeSpec := TypeSpecDef{
		File:     &ast.File{},
		TypeSpec: &ast.TypeSpec{},
		PkgPath:  "model",
	}
	pd = PackagesDefinitions{
		packages: map[string]*PackageDefinitions{
			"model": {
				TypeDefinitions: map[string]*TypeSpecDef{
					"User": &userTypeSpec,
				},
			},
		},
	}
	assert.Equal(t, &userTypeSpec, pd.findTypeSpec("model", "User"))
	assert.Equal(t, nilTypeSpec, pd.findTypeSpec("others", "User"))
}
