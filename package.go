package swag

import (
	"go/ast"
	"go/token"
	"strconv"
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
			Name:    valueSpec.Names[i],
			Type:    valueSpec.Type,
			Value:   valueSpec.Values[i],
			Comment: valueSpec.Comment,
			File:    astFile,
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
			x, err := strconv.ParseInt(valueExpr.Value, 10, 64)
			if err != nil {
				return nil, nil
			}
			return int(x), nil
		case token.STRING, token.CHAR:
			return valueExpr.Value[1 : len(valueExpr.Value)-1], nil
		}
	case *ast.UnaryExpr:
		x, evalType := pkg.evaluateConstValue(file, iota, valueExpr.X, globalEvaluator, recursiveStack)
		if x == nil {
			return nil, nil
		}
		switch valueExpr.Op {
		case token.SUB:
			return -x.(int), evalType
		case token.XOR:
			return ^(x.(int)), evalType
		}
	case *ast.BinaryExpr:
		x, evalTypex := pkg.evaluateConstValue(file, iota, valueExpr.X, globalEvaluator, recursiveStack)
		y, evalTypey := pkg.evaluateConstValue(file, iota, valueExpr.Y, globalEvaluator, recursiveStack)
		if x == nil || y == nil {
			return nil, nil
		}
		evalType := evalTypex
		if evalType == nil {
			evalType = evalTypey
		}
		switch valueExpr.Op {
		case token.ADD:
			if ix, ok := x.(int); ok {
				return ix + y.(int), evalType
			} else if sx, ok := x.(string); ok {
				return sx + y.(string), evalType
			}
		case token.SUB:
			return x.(int) - y.(int), evalType
		case token.MUL:
			return x.(int) * y.(int), evalType
		case token.QUO:
			return x.(int) / y.(int), evalType
		case token.REM:
			return x.(int) % y.(int), evalType
		case token.AND:
			return x.(int) & y.(int), evalType
		case token.OR:
			return x.(int) | y.(int), evalType
		case token.XOR:
			return x.(int) ^ y.(int), evalType
		case token.SHL:
			return x.(int) << y.(int), evalType
		case token.SHR:
			return x.(int) >> y.(int), evalType
		}
	case *ast.ParenExpr:
		return pkg.evaluateConstValue(file, iota, valueExpr.X, globalEvaluator, recursiveStack)
	case *ast.CallExpr:
		//data conversion
		if ident, ok := valueExpr.Fun.(*ast.Ident); ok && len(valueExpr.Args) == 1 && IsGolangPrimitiveType(ident.Name) {
			arg, _ := pkg.evaluateConstValue(file, iota, valueExpr.Args[0], globalEvaluator, recursiveStack)
			return arg, nil
		}
	}
	return nil, nil
}
