package model

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"
	"sync"

	"github.com/swaggo/swag/console"
	"golang.org/x/tools/go/packages"
)

// Global package cache for enum lookup
var (
	enumPackageCache = make(map[string]*packages.Package)
	enumCacheMutex   sync.RWMutex
)

// ParserEnumLookup implements TypeEnumLookup using CoreStructParser
type ParserEnumLookup struct {
	Parser       *CoreStructParser
	BaseModule   string
	PkgPath      string
	packageCache map[string]*packages.Package // Local cache for loaded packages
	cacheMutex   sync.RWMutex                 // Protect local cache
}

// GetEnumsForType looks up enum values for a given type name
// typeName should be fully qualified like "constants.Role" or just "Role"
// or a full package path like "github.com/swaggo/swag/testdata/core_models/constants.Role"
func (p *ParserEnumLookup) GetEnumsForType(typeName string, file *ast.File) ([]EnumValue, error) {
	if p.Parser == nil {
		return nil, fmt.Errorf("parser is nil")
	}

	// Parse the type name to extract package path and type name
	// Handle both "constants.Role" and "github.com/.../constants.Role"
	var targetPkgPath, baseTypeName string
	lastDot := strings.LastIndex(typeName, ".")
	if lastDot == -1 {
		// No dot, just a type name - use the current package path
		baseTypeName = typeName
		targetPkgPath = p.PkgPath
	} else {
		baseTypeName = typeName[lastDot+1:]
		// Everything before the last dot is the package path
		targetPkgPath = typeName[:lastDot]
	}

	console.Printf("$Bold{$Red{Looking up enums for type: %s in package: %s (path: %s)}}\n", baseTypeName, "", targetPkgPath)

	// Initialize local cache if needed
	p.cacheMutex.Lock()
	if p.packageCache == nil {
		p.packageCache = make(map[string]*packages.Package)
	}
	p.cacheMutex.Unlock()

	// Check global cache first
	enumCacheMutex.RLock()
	pkg, pkgCached := enumPackageCache[targetPkgPath]
	enumCacheMutex.RUnlock()

	if !pkgCached {
		// Check local cache
		p.cacheMutex.RLock()
		pkg, pkgCached = p.packageCache[targetPkgPath]
		p.cacheMutex.RUnlock()
	}

	if !pkgCached {
		// Load the package
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
				packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		}

		pkgs, err := packages.Load(cfg, targetPkgPath)
		if err != nil {
			return nil, err
		}

		if len(pkgs) == 0 {
			return nil, fmt.Errorf("no packages found for %s", targetPkgPath)
		}

		pkg = pkgs[0]

		// Store in both caches
		enumCacheMutex.Lock()
		enumPackageCache[targetPkgPath] = pkg
		enumCacheMutex.Unlock()

		p.cacheMutex.Lock()
		p.packageCache[targetPkgPath] = pkg
		p.cacheMutex.Unlock()
	}

	// Look for the type definition and collect const values
	var enums []EnumValue
	var typeFound bool

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			// First, find the type definition
			if genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if ok && typeSpec.Name.Name == baseTypeName {
						typeFound = true
						break
					}
				}
			}

			// Collect constants of this type
			if genDecl.Tok == token.CONST && typeFound {
				for _, spec := range genDecl.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}

					// Check if this const is of the target type
					if valueSpec.Type != nil {
						if ident, ok := valueSpec.Type.(*ast.Ident); ok && ident.Name == baseTypeName {
							// Evaluate the const value
							for i, name := range valueSpec.Names {
								if i < len(valueSpec.Values) {
									// Try to get the value from TypesInfo
									if pkg.TypesInfo != nil {
										if constObj, ok := pkg.TypesInfo.Defs[name].(*types.Const); ok {
											value := constObj.Val()
											comment := ""
											if valueSpec.Comment != nil && len(valueSpec.Comment.List) > 0 {
												comment = strings.TrimSpace(strings.TrimPrefix(valueSpec.Comment.List[0].Text, "//"))
											} else if valueSpec.Doc != nil && len(valueSpec.Doc.List) > 0 {
												comment = strings.TrimSpace(strings.TrimPrefix(valueSpec.Doc.List[len(valueSpec.Doc.List)-1].Text, "//"))
											}

											// Convert constant value to the appropriate Go type
											var enumValue interface{}
											switch value.Kind() {
											case constant.Int:
												// Convert to int64 then to int
												if v, ok := constant.Int64Val(value); ok {
													enumValue = int(v)
												}
											case constant.String:
												// ExactString includes quotes, so use StringVal
												enumValue = constant.StringVal(value)
											default:
												// Fallback to string representation
												enumValue = value.ExactString()
											}

											enums = append(enums, EnumValue{
												Key:     name.Name,
												Value:   enumValue,
												Comment: comment,
											})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if !typeFound {
		return nil, fmt.Errorf("type %s not found", baseTypeName)
	}

	return enums, nil
}
