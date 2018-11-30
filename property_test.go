package swag

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPropertyNameSelectorExpr(t *testing.T) {
	input := &ast.Field{
		Type: &ast.SelectorExpr{
			X: &ast.Ident{
				NamePos: 1136,
				Name:    "time",
				Obj:     (*ast.Object)(nil),
			},
			Sel: &ast.Ident{
				NamePos: 1141,
				Name:    "Time",
				Obj:     (*ast.Object)(nil),
			},
		},
	}
	expected := propertyName{
		"string",
		"string",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameIdentObjectId(t *testing.T) {
	input := &ast.Field{
		Type: &ast.SelectorExpr{
			X: &ast.Ident{
				NamePos: 1136,
				Name:    "hoge",
				Obj:     (*ast.Object)(nil),
			},
			Sel: &ast.Ident{
				NamePos: 1141,
				Name:    "ObjectId",
				Obj:     (*ast.Object)(nil),
			},
		},
	}
	expected := propertyName{
		"string",
		"string",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameIdentUUID(t *testing.T) {
	input := &ast.Field{
		Type: &ast.SelectorExpr{
			X: &ast.Ident{
				NamePos: 1136,
				Name:    "hoge",
				Obj:     (*ast.Object)(nil),
			},
			Sel: &ast.Ident{
				NamePos: 1141,
				Name:    "uuid",
				Obj:     (*ast.Object)(nil),
			},
		},
	}
	expected := propertyName{
		"string",
		"string",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameIdentDecimal(t *testing.T) {
	input := &ast.Field{
		Type: &ast.SelectorExpr{
			X: &ast.Ident{
				NamePos: 1136,
				Name:    "hoge",
				Obj:     (*ast.Object)(nil),
			},
			Sel: &ast.Ident{
				NamePos: 1141,
				Name:    "Decimal",
				Obj:     (*ast.Object)(nil),
			},
		},
	}
	expected := propertyName{
		"number",
		"string",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameIdentTime(t *testing.T) {
	input := &ast.Field{
		Type: &ast.SelectorExpr{
			X: &ast.Ident{
				NamePos: 1136,
				Name:    "hoge",
				Obj:     (*ast.Object)(nil),
			},
			Sel: &ast.Ident{
				NamePos: 1141,
				Name:    "Time",
				Obj:     (*ast.Object)(nil),
			},
		},
	}
	expected := propertyName{
		"string",
		"string",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, nil))
}

func TestGetPropertyNameStarExprIdent(t *testing.T) {
	input := &ast.Field{
		Type: &ast.StarExpr{
			Star: 1026,
			X: &ast.Ident{
				NamePos: 1027,
				Name:    "string",
				Obj:     (*ast.Object)(nil),
			},
		},
	}
	expected := propertyName{
		"string",
		"string",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameStarExprMap(t *testing.T) {
	input := &ast.Field{
		Type: &ast.StarExpr{
			Star: 1026,
			X: &ast.MapType{
				Map: 1027,
				Key: &ast.Ident{
					NamePos: 1034,
					Name:    "string",
					Obj:     (*ast.Object)(nil),
				},
				Value: &ast.Ident{
					NamePos: 1041,
					Name:    "string",
					Obj:     (*ast.Object)(nil),
				},
			},
		},
	}
	expected := propertyName{
		"object",
		"object",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameArrayStarExpr(t *testing.T) {
	input := &ast.Field{
		Type: &ast.ArrayType{
			Lbrack: 465,
			Len:    nil,
			Elt: &ast.StarExpr{
				X: &ast.Ident{
					NamePos: 467,
					Name:    "string",
					Obj:     (*ast.Object)(nil),
				},
			},
		},
	}
	expected := propertyName{
		"array",
		"string",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameArrayStarExprSelector(t *testing.T) {
	input := &ast.Field{
		Type: &ast.ArrayType{
			Lbrack: 1111,
			Len:    nil,
			Elt: &ast.StarExpr{
				X: &ast.SelectorExpr{
					X: &ast.Ident{
						NamePos: 1136,
						Name:    "hoge",
						Obj:     (*ast.Object)(nil),
					},
					Sel: &ast.Ident{
						NamePos: 1141,
						Name:    "ObjectId",
						Obj:     (*ast.Object)(nil),
					},
				},
			},
		},
	}
	expected := propertyName{
		"array",
		"string",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameMap(t *testing.T) {
	input := &ast.Field{
		Type: &ast.MapType{
			Key: &ast.Ident{
				Name: "string",
			},
			Value: &ast.Ident{
				Name: "string",
			},
		},
	}
	expected := propertyName{
		"object",
		"object",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameStruct(t *testing.T) {
	input := &ast.Field{
		Type: &ast.StructType{},
	}
	expected := propertyName{
		"object",
		"object",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestGetPropertyNameInterface(t *testing.T) {
	input := &ast.Field{
		Type: &ast.InterfaceType{},
	}
	expected := propertyName{
		"object",
		"object",
		"",
	}
	assert.Equal(t, expected, getPropertyName(input, New()))
}

func TestParseTag(t *testing.T) {
	searchDir := "testdata/tags"
	mainAPIFile := "main.go"
	p := New()
	p.PropNamingStrategy = PascalCase
	p.ParseAPI(searchDir, mainAPIFile)

	if len(p.swagger.Tags) != 2 {
		t.Log("Number of tags did not match")
		t.Fail()
	}

	dogs := p.swagger.Tags[0]
	if dogs.TagProps.Name != "dogs" || dogs.TagProps.Description != "Dogs are cool" {
		t.Log("Failed to parse dogs name or description")
	}

	cats := p.swagger.Tags[1]
	if cats.TagProps.Name != "cats" || dogs.TagProps.Description != "Cats are the devil" {
		t.Log("Failed to parse cats name or description")
	}
}
