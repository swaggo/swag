package swag

import (
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

// PackageDefinitions files and definition in a package.
type PackageDefinitions struct {
	// files in this package, map key is file's relative path starting package path
	Files map[string]*ast.File

	// definitions in this package, map key is typeName
	TypeDefinitions map[string]*TypeSpecDef

	// const variables in this package, map key is the name
	ConstTable map[string]*ConstVariable

	// const variables in order in this package
	OrderedConst []*ConstVariable

	// package name
	Name string

	// package path
	Path string
}

// ConstVariableGlobalEvaluator an interface used to evaluate enums across packages
type ConstVariableGlobalEvaluator interface {
	EvaluateConstValue(pkg *PackageDefinitions, cv *ConstVariable, recursiveStack map[string]struct{}) (interface{}, ast.Expr)
	EvaluateConstValueByName(file *ast.File, pkgPath, constVariableName string, recursiveStack map[string]struct{}) (interface{}, ast.Expr)
	FindTypeSpec(typeName string, file *ast.File) *TypeSpecDef
}

// NewPackageDefinitions new a PackageDefinitions object
func NewPackageDefinitions(name, pkgPath string) *PackageDefinitions {
	return &PackageDefinitions{
		Name:            name,
		Path:            pkgPath,
		Files:           make(map[string]*ast.File),
		TypeDefinitions: make(map[string]*TypeSpecDef),
		ConstTable:      make(map[string]*ConstVariable),
	}
}

// AddFile add a file
func (pkg *PackageDefinitions) AddFile(pkgPath string, file *ast.File) *PackageDefinitions {
	pkg.Files[pkgPath] = file
	return pkg
}

// AddTypeSpec add a type spec.
func (pkg *PackageDefinitions) AddTypeSpec(name string, typeSpec *TypeSpecDef) *PackageDefinitions {
	pkg.TypeDefinitions[name] = typeSpec
	return pkg
}

// AddConst add a const variable.
func (pkg *PackageDefinitions) AddConst(astFile *ast.File, valueSpec *ast.ValueSpec) *PackageDefinitions {
	for i := 0; i < len(valueSpec.Names) && i < len(valueSpec.Values); i++ {
		variable := &ConstVariable{
			Name:  valueSpec.Names[i],
			Type:  valueSpec.Type,
			Value: valueSpec.Values[i],
			File:  astFile,
		}
		//take the nearest line as comment from comment list or doc list. comment list first.
		if valueSpec.Comment != nil && len(valueSpec.Comment.List) > 0 {
			variable.Comment = valueSpec.Comment.List[0].Text
		} else if valueSpec.Doc != nil && len(valueSpec.Doc.List) > 0 {
			variable.Comment = valueSpec.Doc.List[len(valueSpec.Doc.List)-1].Text
		}
		pkg.ConstTable[valueSpec.Names[i].Name] = variable
		pkg.OrderedConst = append(pkg.OrderedConst, variable)
	}
	return pkg
}

func (pkg *PackageDefinitions) evaluateConstValue(file *ast.File, iota int, expr ast.Expr, globalEvaluator ConstVariableGlobalEvaluator, recursiveStack map[string]struct{}) (interface{}, ast.Expr) {
	switch valueExpr := expr.(type) {
	case *ast.Ident:
		if valueExpr.Name == "iota" {
			return iota, nil
		}
		if pkg.ConstTable != nil {
			if cv, ok := pkg.ConstTable[valueExpr.Name]; ok {
				return globalEvaluator.EvaluateConstValue(pkg, cv, recursiveStack)
			}
		}
	case *ast.SelectorExpr:
		pkgIdent, ok := valueExpr.X.(*ast.Ident)
		if !ok {
			return nil, nil
		}
		return globalEvaluator.EvaluateConstValueByName(file, pkgIdent.Name, valueExpr.Sel.Name, recursiveStack)
	case *ast.BasicLit:
		switch valueExpr.Kind {
		case token.INT:
			// handle underscored number, such as 1_000_000
			if strings.ContainsRune(valueExpr.Value, '_') {
				valueExpr.Value = strings.Replace(valueExpr.Value, "_", "", -1)
			}
			if len(valueExpr.Value) >= 2 && valueExpr.Value[0] == '0' {
				var start, base = 2, 8
				switch valueExpr.Value[1] {
				case 'x', 'X':
					//hex
					base = 16
				case 'b', 'B':
					//binary
					base = 2
				default:
					//octet
					start = 1
				}
				if x, err := strconv.ParseInt(valueExpr.Value[start:], base, 64); err == nil {
					return int(x), nil
				} else if x, err := strconv.ParseUint(valueExpr.Value[start:], base, 64); err == nil {
					return x, nil
				} else {
					panic(err)
				}
			}

			//a basic literal integer is int type in default, or must have an explicit converting type in front
			if x, err := strconv.ParseInt(valueExpr.Value, 10, 64); err == nil {
				return int(x), nil
			} else if x, err := strconv.ParseUint(valueExpr.Value, 10, 64); err == nil {
				return x, nil
			} else {
				panic(err)
			}
		case token.STRING:
			if valueExpr.Value[0] == '`' {
				return valueExpr.Value[1 : len(valueExpr.Value)-1], nil
			}
			return EvaluateEscapedString(valueExpr.Value[1 : len(valueExpr.Value)-1]), nil
		case token.CHAR:
			return EvaluateEscapedChar(valueExpr.Value[1 : len(valueExpr.Value)-1]), nil
		}
	case *ast.UnaryExpr:
		x, evalType := pkg.evaluateConstValue(file, iota, valueExpr.X, globalEvaluator, recursiveStack)
		if x == nil {
			return x, evalType
		}
		return EvaluateUnary(x, valueExpr.Op, evalType)
	case *ast.BinaryExpr:
		x, evalTypex := pkg.evaluateConstValue(file, iota, valueExpr.X, globalEvaluator, recursiveStack)
		y, evalTypey := pkg.evaluateConstValue(file, iota, valueExpr.Y, globalEvaluator, recursiveStack)
		if x == nil || y == nil {
			return nil, nil
		}
		return EvaluateBinary(x, y, valueExpr.Op, evalTypex, evalTypey)
	case *ast.ParenExpr:
		return pkg.evaluateConstValue(file, iota, valueExpr.X, globalEvaluator, recursiveStack)
	case *ast.CallExpr:
		//data conversion
		if len(valueExpr.Args) != 1 {
			return nil, nil
		}
		arg := valueExpr.Args[0]
		if ident, ok := valueExpr.Fun.(*ast.Ident); ok {
			name := ident.Name
			if name == "uintptr" {
				name = "uint"
			}
			value, _ := pkg.evaluateConstValue(file, iota, arg, globalEvaluator, recursiveStack)
			if IsGolangPrimitiveType(name) {
				value = EvaluateDataConversion(value, name)
				return value, nil
			} else if name == "len" {
				if value != nil {
					return reflect.ValueOf(value).Len(), nil
				}
				return pkg.evaluateArrayExprLength(file, iota, arg, globalEvaluator, recursiveStack), nil
			}
			typeDef := globalEvaluator.FindTypeSpec(name, file)
			if typeDef == nil {
				return nil, nil
			}
			return value, valueExpr.Fun
		} else if selector, ok := valueExpr.Fun.(*ast.SelectorExpr); ok {
			typeDef := globalEvaluator.FindTypeSpec(fullTypeName(selector.X.(*ast.Ident).Name, selector.Sel.Name), file)
			if typeDef == nil {
				return nil, nil
			}
			return arg, typeDef.TypeSpec.Type
		}
	}
	return nil, nil
}

func (pkg *PackageDefinitions) evaluateArrayExprLength(file *ast.File, iota int, expr ast.Expr, globalEvaluator ConstVariableGlobalEvaluator, recursiveStack map[string]struct{}) interface{} {
	switch subType := expr.(type) {
	case *ast.Ident:
		if subType.Obj != nil && subType.Obj.Decl != nil {
			if typeSpec, ok := subType.Obj.Decl.(*ast.TypeSpec); ok {
				return pkg.evaluateArrayExprLength(file, iota, typeSpec.Type, globalEvaluator, recursiveStack)
			}
		}
	case *ast.CompositeLit:
		return pkg.evaluateArrayExprLength(file, iota, subType.Type, globalEvaluator, recursiveStack)
	case *ast.IndexExpr:
		eleType := pkg.getArrayType(file, subType, globalEvaluator)
		if eleType == nil {
			return nil
		}
		return pkg.evaluateArrayExprLength(file, iota, eleType, globalEvaluator, recursiveStack)
	case *ast.ArrayType:
		length, _ := pkg.evaluateConstValue(file, iota, subType.Len, globalEvaluator, recursiveStack)
		return length
	case *ast.SelectorExpr:
		sType := pkg.getArrayType(file, subType, globalEvaluator)
		if sType == nil {
			return nil
		}
		return pkg.evaluateArrayExprLength(file, iota, sType, globalEvaluator, recursiveStack)
	}

	return nil
}

func (pkg *PackageDefinitions) getArrayType(file *ast.File, expr ast.Expr, globalEvaluator ConstVariableGlobalEvaluator) ast.Expr {
	switch xType := expr.(type) {
	case *ast.StructType:
		return expr
	case *ast.SelectorExpr:
		if xxType, ok := xType.X.(*ast.Ident); ok {
			typeSpec := globalEvaluator.FindTypeSpec(fullTypeName(xxType.Name, xType.Sel.Name), file)
			if typeSpec != nil {
				return typeSpec.TypeSpec.Type
			}
		}
		xxType := pkg.getArrayType(file, xType.X, globalEvaluator)
		if structType, ok := xxType.(*ast.StructType); ok {
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					if name.Name == xType.Sel.Name {
						return pkg.getArrayType(file, field.Type, globalEvaluator)
					}
				}
			}
		}
	case *ast.CompositeLit:
		return pkg.getArrayType(file, xType.Type, globalEvaluator)
	case *ast.IndexExpr:
		xxTYpe := pkg.getArrayType(file, xType.X, globalEvaluator)
		if arrayType, ok := xxTYpe.(*ast.ArrayType); ok {
			return pkg.getArrayType(file, arrayType.Elt, globalEvaluator)
		}
	case *ast.Ident:
		if xType.Obj != nil && xType.Obj.Decl != nil {
			if typeSpec, ok := xType.Obj.Decl.(*ast.TypeSpec); ok {
				return pkg.getArrayType(file, typeSpec.Type, globalEvaluator)
			}
		}
		typeSpec := globalEvaluator.FindTypeSpec(xType.Name, file)
		if typeSpec != nil {
			return typeSpec.TypeSpec.Type
		}
	}
	return expr
}
