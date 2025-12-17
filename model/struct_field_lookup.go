package model

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/swaggo/swag/console"
	"golang.org/x/tools/go/packages"
)

type CoreStructParser struct {
	basePackage *packages.Package
	packageMap  map[string]*packages.Package
	visited     map[string]bool
}

// toPascalCase converts package_name or package-name to PascalCase (PackageName)
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// Split by underscore or hyphen
	var parts []string
	for _, part := range strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	}) {
		if len(part) > 0 {
			// Capitalize first letter of each part
			parts = append(parts, strings.ToUpper(part[:1])+part[1:])
		}
	}

	if len(parts) == 0 {
		// No delimiters, just capitalize first letter
		return strings.ToUpper(s[:1]) + s[1:]
	}

	return strings.Join(parts, "")
}

func (c *CoreStructParser) LookupStructFields(baseModule, importPath, typeName string) *StructBuilder {
	builder := &StructBuilder{}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName | packages.NeedImports | packages.NeedDeps,
		Fset: token.NewFileSet(),
	}
	// Load the main package with all its dependencies
	pkgs, err := packages.Load(cfg, importPath)
	if err != nil || len(pkgs) == 0 {
		log.Fatalf("failed to load package %s: %v", importPath, err)
	}
	packageMap := make(map[string]*packages.Package)

	// Recursively add all packages including imports and dependencies
	var addPackage func(*packages.Package)
	addPackage = func(pkg *packages.Package) {
		if pkg == nil || packageMap[pkg.PkgPath] != nil {
			return
		}
		packageMap[pkg.PkgPath] = pkg

		// Add all imports
		for _, imp := range pkg.Imports {
			addPackage(imp)
		}
	}

	for _, pkg := range pkgs {
		addPackage(pkg)
	}

	// Set the packageMap on the parser so checkNamed can use it
	c.packageMap = packageMap

	for _, pkg := range pkgs {
		if pkg.PkgPath != importPath {
			continue
		}

		log.Printf("Processing package: %+v\n", pkg)

		visited := make(map[string]bool)
		c.visited = visited
		//fmt.Printf("\n\n-------Package: %s------- \n", pkg.PkgPath)
		//for _, f := range pkg.Syntax {
		//	fmt.Println("Parsed file:", pkg.Fset.Position(f.Pos()).Filename)
		//}
		fields := c.ExtractFieldsRecursive(pkg, typeName, packageMap, visited)

		for _, f := range fields {
			fmt.Printf("Field: %s, Type: %s, Tag: %s\n", f.Name, f.Type, f.Tag)

			if f.Type != nil && strings.Contains(f.Type.String(), "fields.StructField") {

				parts := strings.Split(f.Type.String(), ".StructField[")
				if len(parts) != 2 {
					continue
				}

				subTypeName := strings.TrimSuffix(parts[1], "]")
				// Remove leading * if it's a pointer
				subTypeName = strings.TrimPrefix(subTypeName, "*")

				// Store the original full type name with package path
				var originalTypeName = subTypeName

				var subTypePackage string
				fmt.Printf("----Sub Type Name: %s\n", subTypeName)
				if strings.Contains(subTypeName, "/") {
					// Has a full package path like "github.com/griffnb/assettradingdesk-go/internal/models/billing_plan.FeatureSet"
					// Split by "/" to get path segments
					pathParts := strings.Split(subTypeName, "/")
					lastPart := pathParts[len(pathParts)-1] // "billing_plan.FeatureSet"

					// Split the last part by "." to separate package and type
					dotParts := strings.Split(lastPart, ".")
					if len(dotParts) < 2 {
						continue
					}

					packageName := dotParts[0]            // "billing_plan"
					typeName := dotParts[len(dotParts)-1] // "FeatureSet"

					// Always use package.Type format for consistency with schema storage
					originalTypeName = fmt.Sprintf("%s.%s", packageName, typeName)

					// Check if it's from the same module
					fullPackagePath := strings.Join(pathParts[:len(pathParts)-1], "/") + "/" + packageName

					subTypePackage = fullPackagePath
					subTypeName = typeName
				} else if strings.Contains(subTypeName, ".") {
					// Already in package.Type format like "billing_plan.FeatureSet"
					subParts := strings.Split(subTypeName, ".")
					if len(subParts) < 2 {
						continue
					}
					packageName := subParts[len(subParts)-2]
					typeName := subParts[len(subParts)-1]

					// Use package.Type format
					originalTypeName = fmt.Sprintf("%s.%s", packageName, typeName)

					subTypePackage = strings.Join(subParts[:len(subParts)-1], ".")
					subTypeName = typeName
				} else {
					f.TypeString = subTypeName
					builder.Fields = append(builder.Fields, f)
					continue
				}
				fmt.Printf("-----Final Sub type Package %s\n Final Sub Type Name: %s\n", subTypePackage, subTypeName)

				// If the field is a StructField, we can extract its fields
				fmt.Printf("\n\n-------Sub Package Struct-----: \n%s\n", subTypeName)

				// Try to find the package that contains this type
				var targetPkg *packages.Package
				if subTypePackage != "" {
					targetPkg = packageMap[subTypePackage]
					if targetPkg == nil {
						fmt.Printf("WARNING: Package not found in map for %s\n", subTypePackage)
						fmt.Printf("Available packages: %v\n", func() []string {
							keys := make([]string, 0, len(packageMap))
							for k := range packageMap {
								keys = append(keys, k)
							}
							return keys
						}())
					} else {
						fmt.Printf("-----Found target package: %s\n", targetPkg.PkgPath)
					}
				}
				if targetPkg == nil {
					targetPkg = pkg
					fmt.Printf("------Using current package as fallback: %s\n", pkg.PkgPath)
				}

				subFields := c.ExtractFieldsRecursive(targetPkg, subTypeName, packageMap, make(map[string]bool))
				fmt.Printf("--------Extracted %d subfields for %s\n", len(subFields), subTypeName)
				for _, subField := range subFields {
					fmt.Printf("Sub Field: %s, Type: %s, Tag: %s\n", subField.Name, subField.Type, subField.Tag)
				}

				// Use the original type name with package path
				f.TypeString = originalTypeName
				f.Fields = subFields

				fmt.Printf("-------Set field %s with TypeString=%s and %d Fields\n", f.Name, f.TypeString, len(f.Fields))

				builder.Fields = append(builder.Fields, f)

				fmt.Println("-------- End Sub Package Struct --------")

			} else {
				builder.Fields = append(builder.Fields, f)
			}

		}

	}

	return builder
}

func (c *CoreStructParser) ExtractFieldsRecursive(pkg *packages.Package, typeName string, packageMap map[string]*packages.Package, visited map[string]bool) []*StructField {
	if visited[typeName] {
		return nil
	}
	visited[typeName] = true

	var fields []*StructField

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range genDecl.Specs {
				ts, ok := spec.(*ast.TypeSpec)

				if !ok || ts.Name.Name != typeName {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				fmt.Printf("----Matched StructType & Processing: %s (has %d fields)\n", ts.Name.Name, len(st.Fields.List))
				for i, field := range st.Fields.List {
					var fieldName string
					if len(field.Names) > 0 {
						fieldName = field.Names[0].Name
					} else {
						switch expr := field.Type.(type) {
						case *ast.Ident:
							fieldName = expr.Name
						case *ast.SelectorExpr:
							fieldName = expr.Sel.Name
						default:
							fieldName = "unknown"
						}
					}

					tag := ""
					if field.Tag != nil {
						tag = strings.Trim(field.Tag.Value, "`")
					}

					var fieldType types.Type
					var obj types.Object
					if len(field.Names) > 0 {
						if obj, ok = pkg.TypesInfo.Defs[field.Names[0]]; ok {
							fieldType = obj.Type()
						}
					} else {
						if typ := pkg.TypesInfo.Types[field.Type]; typ.Type != nil {
							fieldType = typ.Type
						}
					}

					fmt.Printf(
						"----[Field %d/%d] Validating Field Name: %s, Type: %s (%T), Tag: %s\n",
						i+1,
						len(st.Fields.List),
						fieldName,
						fieldType,
						fieldType,
						tag,
					)

					// Embedded Fields
					if subFields, _, ok := c.checkNamed(fieldType); ok {

						if len(subFields) == 0 {
							fmt.Printf("Skipping empty embedded field: %s\n", fieldName)
							continue
						}
						fields = append(fields, subFields...)
						continue
					}

					if subFields, typeName, ok := c.checkStruct(fieldType); ok {
						fields = append(fields, &StructField{
							Name:       fieldName,
							Type:       fieldType,
							Tag:        tag,
							TypeString: typeName,
							Fields:     subFields,
						})

						fmt.Printf("----Added Struct Field: %s of type %s with %d subfields\n", fieldName, typeName, len(subFields))

						continue
					}

					if subFields, typeName, ok := c.checkSlice(fieldType); ok {
						fields = append(fields, &StructField{
							Name:       fieldName,
							Type:       fieldType,
							Tag:        tag,
							TypeString: typeName,
							Fields:     subFields,
						})
						continue
					}

					if subFields, typeName, ok := c.checkMap(fieldType); ok {
						fields = append(fields, &StructField{
							Name:       fieldName,
							Type:       fieldType,
							Tag:        tag,
							TypeString: typeName,
							Fields:     subFields,
						})
						continue
					}

					fields = append(fields, &StructField{
						Name:       fieldName,
						Type:       fieldType,
						Tag:        tag,
						TypeString: fieldType.String(),
					})
				}
			}
		}
	}

	return fields
}

func (c *CoreStructParser) checkNamed(fieldType types.Type) ([]*StructField, *types.Named, bool) {
	named, ok := fieldType.(*types.Named)
	if ok {
		if named.Obj().Pkg().Path() == "github.com/griffnb/core/lib/model/fields" {
			return nil, nil, false
		}
		if _, ok := named.Underlying().(*types.Struct); ok {
			fmt.Printf("Found sub type Package %s Name %s\n", named.Obj().Pkg().Path(), named.Obj().Name())
			nextPackage, ok := c.packageMap[named.Obj().Pkg().Path()]
			if !ok {
				fmt.Printf("Package not found for %s\n", named.Obj().Pkg().Path())
				return nil, nil, true
			}
			fmt.Printf("Next Package: %s\n", nextPackage.PkgPath)
			subFields := c.ExtractFieldsRecursive(nextPackage, named.Obj().Name(), c.packageMap, c.visited)
			return subFields, named, true
		}
	}

	return nil, nil, false
}

func (c *CoreStructParser) checkStruct(fieldType types.Type) ([]*StructField, string, bool) {
	pointer, isPointer := fieldType.(*types.Pointer)
	if isPointer {
		fields, namedType, ok := c.checkNamed(pointer.Elem())
		if ok && namedType != nil {
			return fields, fmt.Sprintf("*%s", namedType.Obj().Name()), true
		}
	} else {
		fields, namedType, ok := c.checkNamed(fieldType)
		if ok {
			return fields, namedType.Obj().Name(), true
		}
	}

	return nil, "", false
}

func (c *CoreStructParser) checkSlice(fieldType types.Type) ([]*StructField, string, bool) {
	slice, isSlice := fieldType.(*types.Slice)
	if isSlice {
		fields, structType, ok := c.checkStruct(slice.Elem())
		if ok {
			return fields, fmt.Sprintf("[]%s", structType), true
		}
	}

	return nil, "", false
}

func (c *CoreStructParser) checkMap(fieldType types.Type) ([]*StructField, string, bool) {
	mapType, isMap := fieldType.(*types.Map)
	if isMap {
		var mapPart string
		if strings.Contains(fieldType.String(), "*github.com") {
			mapPart = strings.Split(fieldType.String(), "*github.com")[0]
		} else {
			mapPart = strings.Split(fieldType.String(), "github.com/")[0]
		}

		fields, sliceType, isSlice := c.checkSlice(mapType.Elem())
		if isSlice {
			return fields, fmt.Sprintf("%s%s", mapPart, sliceType), true
		}

		fields, structType, isStruct := c.checkStruct(mapType.Elem())
		if isStruct {
			return fields, fmt.Sprintf("%s%s", mapPart, structType), true
		}
	}

	return nil, "", false
}

// BuildAllSchemas generates both public and non-public schema variants for a type
// Returns a map of schema names to schemas (includes both base and Public variants)
func BuildAllSchemas(baseModule, pkgPath, typeName string) (map[string]*spec.Schema, error) {
	parser := &CoreStructParser{}

	// Extract package name from pkgPath (last segment)
	packageName := pkgPath
	if idx := strings.LastIndex(pkgPath, "/"); idx >= 0 {
		packageName = pkgPath[idx+1:]
	}

	// Lookup struct fields using existing LookupStructFields
	builder := parser.LookupStructFields(baseModule, pkgPath, typeName)
	if builder == nil {
		return nil, fmt.Errorf("failed to lookup struct fields for %s", typeName)
	}

	allSchemas := make(map[string]*spec.Schema)
	processed := make(map[string]bool) // Track processed types to avoid infinite recursion

	// Generate schemas for the main type with package prefix
	fullTypeName := packageName + "." + typeName
	err := buildSchemasRecursive(builder, typeName, false, allSchemas, processed, parser, baseModule, pkgPath, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to build schemas for %s: %w", fullTypeName, err)
	}

	err = buildSchemasRecursive(builder, typeName+"Public", true, allSchemas, processed, parser, baseModule, pkgPath, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to build public schemas for %s: %w", fullTypeName, err)
	}

	return allSchemas, nil
}

// buildSchemasRecursive recursively builds schemas for a type and all its nested types
func buildSchemasRecursive(builder *StructBuilder, schemaName string, public bool, allSchemas map[string]*spec.Schema, processed map[string]bool, parser *CoreStructParser, baseModule, pkgPath, packageName string) error {
	// Avoid infinite recursion
	if processed[schemaName] {
		return nil
	}
	processed[schemaName] = true

	// Extract base type name (remove Public suffix if present)
	baseTypeName := schemaName
	if public && strings.HasSuffix(schemaName, "Public") {
		baseTypeName = strings.TrimSuffix(schemaName, "Public")
	}

	// Build schema for current type
	schema, nestedTypes, err := builder.BuildSpecSchema(baseTypeName, public)
	if err != nil {
		return fmt.Errorf("failed to build schema for %s: %w", schemaName, err)
	}

	// Store the schema with package prefix
	fullSchemaName := packageName + "." + schemaName

	// Set title to create clean class names in code generators
	// Strategy:
	// 1. If package is a prefix of type (e.g., account.Account, account.AccountJoined)
	//    → use just the type name (Account, AccountJoined)
	// 2. Otherwise (e.g., account.Properties, billing_plan.FeatureSet)
	//    → combine as PascalCase (AccountProperties, BillingPlanFeatureSet)

	typeName := schemaName // Use schemaName to preserve "Public" suffix if present

	// Remove underscores/hyphens from package for comparison
	// e.g., billing_plan → billingplan
	packageNoSeparators := strings.ReplaceAll(strings.ReplaceAll(packageName, "_", ""), "-", "")

	if strings.HasPrefix(strings.ToLower(typeName), strings.ToLower(packageNoSeparators)) {
		// Package is a prefix of type name (case-insensitive, ignoring separators)
		// e.g., account.Account → Account, account.AccountJoined → AccountJoined
		//       billing_plan.BillingPlanJoined → BillingPlanJoined
		schema.Title = typeName
	} else {
		// Package and type don't align - combine them
		// e.g., account.Properties → AccountProperties
		// Convert package_name to PascalCase: billing_plan → BillingPlan
		packagePascal := toPascalCase(packageName)
		schema.Title = packagePascal + typeName
	}

	allSchemas[fullSchemaName] = schema

	// Recursively process nested types
	for _, nestedTypeName := range nestedTypes {
		// Parse package name and type name from nested type
		// e.g., "account.Properties" -> package="account", type="Properties"
		// e.g., "billing_plan.FeatureSet" -> package="billing_plan", type="FeatureSet"
		var nestedPackageName, baseNestedType string
		if strings.Contains(nestedTypeName, ".") {
			parts := strings.Split(nestedTypeName, ".")
			nestedPackageName = parts[0]
			baseNestedType = parts[len(parts)-1]
		} else {
			// No package prefix, use current package
			nestedPackageName = packageName
			baseNestedType = nestedTypeName
		}

		// Determine the full package path for the nested type
		// If it's from the same package, use the current pkgPath
		// Otherwise, construct the path by replacing the last segment
		nestedPkgPath := pkgPath
		if nestedPackageName != packageName {
			// Different package - need to construct the full path
			// e.g., if pkgPath is "github.com/swaggo/swag/testdata/core_models/account"
			// and nestedPackageName is "billing_plan"
			// then nestedPkgPath should be "github.com/swaggo/swag/testdata/core_models/billing_plan"
			if idx := strings.LastIndex(pkgPath, "/"); idx >= 0 {
				nestedPkgPath = pkgPath[:idx+1] + nestedPackageName
			} else {
				nestedPkgPath = nestedPackageName
			}
		}

		// Need to lookup the nested type's fields using the correct package path
		nestedBuilder := parser.LookupStructFields(baseModule, nestedPkgPath, baseNestedType)
		if nestedBuilder == nil {
			console.Printf("$Yellow{Warning: Could not lookup nested type %s in package %s}\n", baseNestedType, nestedPkgPath)
			continue
		}

		// Generate both public and non-public variants for nested types
		// Use the nested package name when storing schemas
		err = buildSchemasRecursive(nestedBuilder, baseNestedType, false, allSchemas, processed, parser, baseModule, nestedPkgPath, nestedPackageName)
		if err != nil {
			return err
		}

		err = buildSchemasRecursive(nestedBuilder, baseNestedType+"Public", true, allSchemas, processed, parser, baseModule, nestedPkgPath, nestedPackageName)
		if err != nil {
			return err
		}
	}

	return nil
}
