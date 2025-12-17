package model

import (
	"fmt"
	"strings"
)

type StructBuilder struct {
	Fields []*StructField `json:"fields"` // For nested structs
}

func (this *StructBuilder) BuildStructs(name string, public bool, aliasName string, childStructs map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", name))
	fmt.Printf("\n\nBuilding struct %s (public=%v) with %d fields\n", name, public, len(this.Fields))
	for _, field := range this.Fields {
		if public && !field.IsPublic() {
			continue
		}
		sb.WriteString(field.BuildStructDef(public))

		fmt.Printf("Field %s: IsStruct=%v, IsPublic=%v, TypeString=%s, FieldsCount=%d\n",
			field.Name, field.IsStruct(), field.IsPublic(), field.TypeString, len(field.Fields))

		if field.IsStruct() && public && field.IsPublic() {
			// Strip package prefix from TypeString for the key
			typeName := field.TypeString
			if strings.Contains(typeName, ".") {
				parts := strings.Split(typeName, ".")
				typeName = parts[len(parts)-1]
			}
			fmt.Printf("Creating child struct for %s -> %sPublic\n", field.TypeString, typeName)
			childStructs[typeName+"Public"] = field.BuildStruct(childStructs, public, typeName)
		}
	}
	sb.WriteString(fmt.Sprintf("}//@name %s\n", aliasName))

	// fmt.Printf("%s, %v | Sub Structs: %+v\n", name, public, childStructs)

	return sb.String()
}

func (this *StructBuilder) BuildInterface(name string, public bool, childStructs map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("export interface %s extends BaseModel {\n", name))
	fmt.Printf("\n\nBuilding struct %s (public=%v) with %d fields\n", name, public, len(this.Fields))
	for _, field := range this.Fields {
		if public && !field.IsPublic() {
			continue
		}
		sb.WriteString(field.BuildInterfaceDef(public))

		fmt.Printf("Field %s: IsStruct=%v, IsPublic=%v, TypeString=%s, FieldsCount=%d\n",
			field.Name, field.IsStruct(), field.IsPublic(), field.TypeString, len(field.Fields))

		if field.IsStruct() && public && field.IsPublic() {
			// Strip package prefix from TypeString for the key
			typeName := field.TypeString
			if strings.Contains(typeName, ".") {
				parts := strings.Split(typeName, ".")
				typeName = parts[len(parts)-1]
			}
			fmt.Printf("Creating child struct for %s -> %sPublic\n", field.TypeString, typeName)
			childStructs[typeName+"Public"] = field.BuildInterface(childStructs, public, typeName)
		}
	}
	sb.WriteString("}\n")

	// fmt.Printf("%s, %v | Sub Structs: %+v\n", name, public, childStructs)

	return sb.String()
}
