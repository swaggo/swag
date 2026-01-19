package model

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"
	"sync"

	"github.com/go-openapi/spec"
	"github.com/swaggo/swag/console"
	"golang.org/x/tools/go/packages"
)

// Global package cache shared across all parsers
var (
	globalPackageCache = make(map[string]*packages.Package)
	globalCacheMutex   sync.RWMutex
	debugMode          = false // Set to true to enable debug logging
)

type CoreStructParser struct {
	basePackage   *packages.Package
	packageMap    map[string]*packages.Package
	visited       map[string]bool
	packageCache  map[string]*packages.Package // Cache loaded packages
	typeCache     map[string]*StructBuilder    // Cache processed types
	cacheMutex    sync.RWMutex                 // Protect caches
	packageLoader sync.Once                    // Load packages only once
}

// debugLog prints debug messages only when debugMode is enabled
func debugLog(format string, args ...interface{}) {
	if debugMode {
		fmt.Printf(format, args...)
	}
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
	// Check type cache first
	cacheKey := importPath + ":" + typeName
	c.cacheMutex.RLock()
	if cached, exists := c.typeCache[cacheKey]; exists {
		c.cacheMutex.RUnlock()
		debugLog("Using cached type: %s\n", cacheKey)
		return cached
	}
	c.cacheMutex.RUnlock()

	builder := &StructBuilder{}

	// Initialize caches if needed
	c.cacheMutex.Lock()
	if c.packageCache == nil {
		c.packageCache = make(map[string]*packages.Package)
	}
	if c.typeCache == nil {
		c.typeCache = make(map[string]*StructBuilder)
	}
	c.cacheMutex.Unlock()

	// Check global cache first
	globalCacheMutex.RLock()
	pkg, pkgCached := globalPackageCache[importPath]
	globalCacheMutex.RUnlock()

	var packageMap map[string]*packages.Package

	if !pkgCached {
		debugLog("Loading package: %s\n", importPath)
		cfg := &packages.Config{
			Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName | packages.NeedImports | packages.NeedDeps,
			Fset: token.NewFileSet(),
		}
		// Load the main package with all its dependencies
		pkgs, err := packages.Load(cfg, importPath)
		if err != nil || len(pkgs) == 0 {
			log.Fatalf("failed to load package %s: %v", importPath, err)
		}
		packageMap = make(map[string]*packages.Package)

		// Recursively add all packages including imports and dependencies
		var addPackage func(*packages.Package)
		addPackage = func(p *packages.Package) {
			if p == nil || packageMap[p.PkgPath] != nil {
				return
			}
			packageMap[p.PkgPath] = p

			// Add all imports
			for _, imp := range p.Imports {
				addPackage(imp)
			}
		}

		for _, p := range pkgs {
			addPackage(p)
		}

		// Cache all loaded packages in both local and global cache
		globalCacheMutex.Lock()
		c.cacheMutex.Lock()
		for path, p := range packageMap {
			globalPackageCache[path] = p
			c.packageCache[path] = p
		}
		pkg = packageMap[importPath]
		c.cacheMutex.Unlock()
		globalCacheMutex.Unlock()
		debugLog("Cached %d packages from %s\n", len(packageMap), importPath)
	} else {
		debugLog("Using globally cached package: %s\n", importPath)
		// Use cached packages from global cache
		globalCacheMutex.RLock()
		packageMap = make(map[string]*packages.Package)
		for k, v := range globalPackageCache {
			packageMap[k] = v
		}
		globalCacheMutex.RUnlock()
	}

	// Set the packageMap on the parser so checkNamed can use it
	c.packageMap = packageMap

	if pkg == nil || pkg.PkgPath != importPath {
		debugLog("Package not found or mismatch: %v\n", importPath)
		return builder
	}

	debugLog("Processing package: %+v %s\n", pkg, typeName)

	visited := make(map[string]bool)
	c.visited = visited
	fields := c.ExtractFieldsRecursive(pkg, typeName, packageMap, visited)

	// Process all fields
	for _, f := range fields {
		debugLog("Field: %s, Type: %s, Tag: %s\n", f.Name, f.Type, f.Tag)

		// Check if it's a special StructField type that needs expansion
		if f.Type != nil && strings.Contains(f.Type.String(), "fields.StructField") {
			c.processStructField(f, packageMap, builder)
		} else {
			builder.Fields = append(builder.Fields, f)
		}
	}

	// Cache the result before returning
	c.cacheMutex.Lock()
	c.typeCache[cacheKey] = builder
	c.cacheMutex.Unlock()

	return builder
}

// processStructField handles the expansion of StructField[T] types
func (c *CoreStructParser) processStructField(f *StructField, packageMap map[string]*packages.Package, builder *StructBuilder) {
	parts := strings.Split(f.Type.String(), ".StructField[")
	if len(parts) != 2 {
		builder.Fields = append(builder.Fields, f)
		return
	}

	subTypeName := strings.TrimSuffix(parts[1], "]")

	// Handle array and pointer prefixes - keep them in originalTypeName but strip for type lookup
	arrayPrefix := ""
	if strings.HasPrefix(subTypeName, "[]") {
		arrayPrefix = "[]"
		subTypeName = strings.TrimPrefix(subTypeName, "[]")
	}
	if strings.HasPrefix(subTypeName, "*") {
		arrayPrefix = arrayPrefix + "*"
		subTypeName = strings.TrimPrefix(subTypeName, "*")
	}

	// Store the original full type name with package path
	originalTypeName := subTypeName
	var subTypePackage string

	debugLog("----Sub Type Name: %s (arrayPrefix: %s)\n", subTypeName, arrayPrefix)

	// Parse package and type name
	if strings.Contains(subTypeName, "/") {
		// Full package path like "github.com/griffnb/project/internal/models/billing_plan.FeatureSet"
		pathParts := strings.Split(subTypeName, "/")
		lastPart := pathParts[len(pathParts)-1]

		dotParts := strings.Split(lastPart, ".")
		if len(dotParts) < 2 {
			f.TypeString = subTypeName
			builder.Fields = append(builder.Fields, f)
			return
		}

		packageName := dotParts[0]
		typeName := dotParts[len(dotParts)-1]
		originalTypeName = arrayPrefix + fmt.Sprintf("%s.%s", packageName, typeName)
		fullPackagePath := strings.Join(pathParts[:len(pathParts)-1], "/") + "/" + packageName

		subTypePackage = fullPackagePath
		subTypeName = typeName
	} else if strings.Contains(subTypeName, ".") {
		// Already in package.Type format
		subParts := strings.Split(subTypeName, ".")
		if len(subParts) < 2 {
			f.TypeString = subTypeName
			builder.Fields = append(builder.Fields, f)
			return
		}
		packageName := subParts[len(subParts)-2]
		typeName := subParts[len(subParts)-1]
		originalTypeName = arrayPrefix + fmt.Sprintf("%s.%s", packageName, typeName)

		subTypePackage = strings.Join(subParts[:len(subParts)-1], ".")
		subTypeName = typeName
	} else {
		f.TypeString = arrayPrefix + subTypeName
		builder.Fields = append(builder.Fields, f)
		return
	}

	debugLog("-----Final Sub type Package %s\n Final Sub Type Name: %s\n", subTypePackage, subTypeName)

	// Find the target package
	targetPkg := packageMap[subTypePackage]
	if targetPkg == nil {
		debugLog("WARNING: Package not found in map for %s\n", subTypePackage)
		targetPkg = c.basePackage
	} else {
		debugLog("-----Found target package: %s\n", targetPkg.PkgPath)
	}

	if targetPkg == nil {
		debugLog("------No package available\n")
		f.TypeString = originalTypeName
		builder.Fields = append(builder.Fields, f)
		return
	}

	// Extract subfields
	debugLog("\n\n-------Sub Package Struct-----: \n%s\n", subTypeName)
	subFields := c.ExtractFieldsRecursive(targetPkg, subTypeName, packageMap, make(map[string]bool))
	debugLog("--------Extracted %d subfields for %s\n", len(subFields), subTypeName)

	for _, subField := range subFields {
		debugLog("Sub Field: %s, Type: %s, Tag: %s\n", subField.Name, subField.Type, subField.Tag)
	}

	f.TypeString = originalTypeName
	f.Fields = subFields
	debugLog("-------Set field %s with TypeString=%s and %d Fields\n", f.Name, f.TypeString, len(f.Fields))

	builder.Fields = append(builder.Fields, f)
	fmt.Println("-------- End Sub Package Struct --------")
}

func (c *CoreStructParser) ExtractFieldsRecursive(
	pkg *packages.Package,
	typeName string,
	packageMap map[string]*packages.Package,
	visited map[string]bool,
) []*StructField {
	// Create a unique cache key with package path
	cacheKey := pkg.PkgPath + ":" + typeName
	if visited[cacheKey] {
		return nil
	}
	visited[cacheKey] = true

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
				debugLog("----Matched StructType & Processing: %s (has %d fields)\n", ts.Name.Name, len(st.Fields.List))
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

					debugLog(
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
							debugLog("Skipping empty embedded field: %s\n", fieldName)
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

						debugLog("----Added Struct Field: %s of type %s with %d subfields\n", fieldName, typeName, len(subFields))
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

// shouldTreatAsSwaggerPrimitive checks if a named type should be treated as a primitive in Swagger
// even though it might be a struct in Go (like time.Time or decimal.Decimal)
func shouldTreatAsSwaggerPrimitive(named *types.Named) bool {
	if named.Obj().Pkg() == nil {
		return false
	}

	pkgPath := named.Obj().Pkg().Path()
	typeName := named.Obj().Name()

	// Types that are structs in Go but should be primitives in Swagger
	primitiveTypes := map[string][]string{
		"time":                          {"Time"},
		"github.com/shopspring/decimal": {"Decimal"},
		"gopkg.in/guregu/null.v4":       {"String", "Int", "Float", "Bool", "Time"},
		"database/sql":                  {"NullString", "NullInt64", "NullFloat64", "NullBool", "NullTime"},
	}

	if typeNames, ok := primitiveTypes[pkgPath]; ok {
		for _, name := range typeNames {
			if typeName == name {
				return true
			}
		}
	}

	return false
}

func (c *CoreStructParser) checkNamed(fieldType types.Type) ([]*StructField, *types.Named, bool) {
	named, ok := fieldType.(*types.Named)
	if ok {
		if strings.Contains(named.Obj().Pkg().Path(), "/lib/model/fields") {
			return nil, nil, false
		}
		// Skip types that should be treated as primitives in Swagger
		if shouldTreatAsSwaggerPrimitive(named) {
			return nil, nil, false
		}
		if _, ok := named.Underlying().(*types.Struct); ok {
			debugLog("Found sub type Package %s Name %s\n", named.Obj().Pkg().Path(), named.Obj().Name())
			nextPackage, ok := c.packageMap[named.Obj().Pkg().Path()]
			if !ok {
				debugLog("Package not found for %s\n", named.Obj().Pkg().Path())
				return nil, nil, true
			}
			debugLog("Next Package: %s\n", nextPackage.PkgPath)
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
func buildSchemasRecursive(
	builder *StructBuilder,
	schemaName string,
	public bool,
	allSchemas map[string]*spec.Schema,
	processed map[string]bool,
	parser *CoreStructParser,
	baseModule, pkgPath, packageName string,
) error {
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
	// Create a parser-based enum lookup that can access the packages
	enumLookup := &ParserEnumLookup{Parser: parser, BaseModule: baseModule, PkgPath: pkgPath}
	schema, nestedTypes, err := builder.BuildSpecSchema(baseTypeName, public, enumLookup)
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

		err = buildSchemasRecursive(
			nestedBuilder,
			baseNestedType+"Public",
			true,
			allSchemas,
			processed,
			parser,
			baseModule,
			nestedPkgPath,
			nestedPackageName,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
