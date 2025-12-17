package model

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/swaggo/swag/console"
)

type StructField struct {
	Name       string         `json:"name"`
	Type       types.Type     `json:"type"`
	TypeString string         `json:"type_string"` // For easier JSON serialization
	Tag        string         `json:"tag"`
	Fields     []*StructField `json:"fields"` // For nested structs
}

func (this *StructField) GetColumn() string {
	tags := this.GetTags()
	if columnTag, ok := tags["column"]; ok {
		return columnTag
	} else if jsonTag, ok := tags["json"]; ok {
		parts := strings.Split(jsonTag, ",")
		return parts[0]
	}
	return ""
}

func (this *StructField) GetColumnType() string {
	tags := this.GetTags()
	sanitizedType := strings.ReplaceAll(tags["type"], "\"", "")
	sanitizedType = strings.Split(sanitizedType, ",")[0]
	return sanitizedType
}

func (this *StructField) IsStruct() bool {
	return len(this.Fields) > 0
}

func (this *StructField) IsJoined() bool {
	tags := this.GetTags()
	_, ok := tags["json"]
	return ok
}

func (this *StructField) BuildStruct(structs map[string]string, public bool, typeName string) string {
	var sb strings.Builder

	name := typeName
	if public {
		name = fmt.Sprintf("%sPublic", typeName)
	}

	sb.WriteString(fmt.Sprintf("type %s struct {\n", name))
	for _, field := range this.Fields {
		if public && !field.IsPublic() {
			continue
		}

		sb.WriteString(field.BuildStructDef(public))
		if field.IsStruct() {
			// Strip package prefix for nested types
			nestedTypeName := field.TypeString
			if strings.Contains(nestedTypeName, ".") {
				parts := strings.Split(nestedTypeName, ".")
				nestedTypeName = parts[len(parts)-1]
			}
			structs[nestedTypeName+"Public"] = field.BuildStruct(structs, public, nestedTypeName)
		}
	}
	sb.WriteString("}\n")
	return sb.String()
}

func (this *StructField) BuildStructDef(public bool) string {
	var sb strings.Builder

	tags := this.GetTags()

	jsonKey := ""

	if tags["column"] != "" {
		jsonKey = tags["column"]
	} else {
		jsonKey = tags["json"]
	}

	fieldType := ""
	if strings.Contains(this.TypeString, "fields.") {
		fieldType = this.GetGoType()
	} else {
		fieldType = this.TypeString
	}

	if this.IsStruct() && public {
		// Strip package prefix if present (e.g., "billing_plan.FeatureSet" -> "FeatureSet")
		if strings.Contains(fieldType, ".") {
			parts := strings.Split(fieldType, ".")
			fieldType = parts[len(parts)-1]
		}
		fieldType = fmt.Sprintf("%sPublic", fieldType)
	}

	sb.WriteString(fmt.Sprintf("\t%s %s `json:%s`\n", this.Name, fieldType, jsonKey))
	return sb.String()
}

func (this *StructField) BuildInterface(structs map[string]string, public bool, typeName string) string {
	var sb strings.Builder

	name := typeName
	if public {
		name = fmt.Sprintf("%sPublic", typeName)
	}

	sb.WriteString(fmt.Sprintf("interface %s {\n", name))
	for _, field := range this.Fields {
		if public && !field.IsPublic() {
			continue
		}

		sb.WriteString(field.BuildInterfaceDef(public))
		if field.IsStruct() {
			// Strip package prefix for nested types
			nestedTypeName := field.TypeString
			if strings.Contains(nestedTypeName, ".") {
				parts := strings.Split(nestedTypeName, ".")
				nestedTypeName = parts[len(parts)-1]
			}
			structs[nestedTypeName+"Public"] = field.BuildInterface(structs, public, nestedTypeName)
		}
	}
	sb.WriteString("}\n")
	return sb.String()
}

func (this *StructField) BuildInterfaceDef(public bool) string {
	nullableSuffix := ""
	if this.IsNullable() {
		nullableSuffix = " | null"
	}

	if this.IsStruct() && public {
		// Strip package prefix if present (e.g., "billing_plan.FeatureSet" -> "FeatureSet")

		publicType := fmt.Sprintf("%sPublic", this.GetInterfaceType())
		return fmt.Sprintf("  %s: %s%s;\n", this.GetColumn(), publicType, nullableSuffix)
	}

	return fmt.Sprintf("  %s: %s%s;\n", this.GetColumn(), this.GetInterfaceType(), nullableSuffix)
}

func (this *StructField) IsPublic() bool {
	_, ok := this.GetTags()["public"]
	return ok
}

func (this *StructField) GetTags() map[string]string {
	tags := strings.Split(this.Tag, " ")
	result := make(map[string]string)
	for _, tag := range tags {
		parts := strings.SplitN(tag, ":", 2)
		if len(parts) == 2 {
			key := strings.Trim(parts[0], "`")
			value := strings.Trim(parts[1], "`")
			result[key] = strings.Trim(value, "\"")
		}
	}
	return result
}

func (this *StructField) GetGoType() string {
	columnType := this.GetColumnType()
	switch columnType {
	case "boolean":
		return "boolean"
	case "uuid":
		return "string"
	case "text", "varchar":
		return "string"
	case "integer", "smallint", "bigint":
		return "int64"
	case "numeric":
		return "decimal.Decimal"
	case "date":
		return "time.Time"
	case "timestamp with time zone", "tswtz":
		return "time.Time"
	default:
		return "any"
	}
}

func (this *StructField) GetAttrType() string {
	switch this.GetColumnType() {
	case "boolean":
		return "bool"
	case "uuid":
		return "uuid"
	case "text", "varchar":
		return "string"
	case "integer", "smallint", "bigint", "int":
		if strings.HasSuffix(this.GetColumn(), "_ts") {
			return "ts-dayjs"
		}
		return "number"
	case "numeric":
		return "decimal"
	case "jsonb":
		return "json"
	case "date":
		return "date-dayjs"
	case "timestamp with time zone", "tswtz":
		return "date-dayjs"
	default:
		console.Printf("$Red{Unknown Go type '%s' for column '%s', defaulting to 'any'}", this.GetColumnType(), this.GetColumn())
		return "any"
	}
}

func (this *StructField) GetInterfaceTypeFromGoType() string {
	switch this.TypeString {
	case "boolean":
		return "boolean"
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return "number"
	case "time.Time", "*time.Time":
		return "dayjs.Dayjs | null"
	default:

		fmt.Printf("Unknown type string: %s\n", this.TypeString)

		// Try to infer from Go type if db type is unknown
		if strings.Contains(this.TypeString, "[]") {
			elementType := strings.TrimPrefix(this.TypeString, "[]")
			elementType = strings.TrimPrefix(elementType, "*")
			if strings.Contains(elementType, ".") {
				parts := strings.Split(elementType, ".")
				elementType = parts[len(parts)-1]
			}
			return fmt.Sprintf("%s[]", elementType)
		} else if strings.Contains(this.TypeString, "map[") {
			return "Record<string, any>"
		}
		return "any"
	}
}

// goTypeToTSInterface converts a database type or Go type to TypeScript interface type
func (this *StructField) GetInterfaceType() string {
	switch this.GetColumnType() {
	case "boolean":
		return "boolean"
	case "uuid":
		return "string"
	case "text", "varchar":
		return "string"
	case "integer", "smallint", "bigint":
		if strings.HasSuffix(this.GetColumn(), "_ts") {
			return "dayjs.Dayjs | null"
		}
		return "number"
	case "numeric":
		return "number"
	case "jsonb", "json":
		return snakeToPascal(this.GetColumn())
	case "date":
		return "dayjs.Dayjs | null"
	case "timestamp with time zone", "tswtz":
		return "dayjs.Dayjs | null"
	default:
		return this.GetInterfaceTypeFromGoType()
	}
}

func (this *StructField) GetModelAttrAndDefault() (string, string) {
	defaultValue := this.GetDefault()

	nullableString := ""
	if this.IsNullable() {
		nullableString = " | null"
		if defaultValue == "" {
			defaultValue = "null"
		}
	}

	switch this.GetAttrType() {
	case "uuid":
		if !this.IsNullable() {
			if defaultValue == "" {
				defaultValue = "''"
			}
		}
		return fmt.Sprintf("string %s", nullableString), defaultValue
	case "bool":
		return fmt.Sprintf("boolean %s", nullableString), defaultValue
	case "decimal":
		if defaultValue == "" && !this.IsNullable() {
			defaultValue = "0"
		}
		return fmt.Sprintf("number %s", nullableString), defaultValue
	case "date-dayjs":
		if defaultValue == "" && this.IsNullable() {
			defaultValue = "null"
		}
		return fmt.Sprintf("dayjs.Dayjs %s", nullableString), defaultValue
	case "ts-dayjs":
		return "dayjs.Dayjs | null", "null"
	case "string":
		if defaultValue == "" && !this.IsNullable() {
			defaultValue = "''"
		}
		return fmt.Sprintf("%s %s", this.GetAttrType(), nullableString), defaultValue
	case "number":
		if defaultValue == "" && !this.IsNullable() {
			defaultValue = "0"
		}

		return fmt.Sprintf("%s %s", this.GetAttrType(), nullableString), defaultValue
	case "json":
		if defaultValue == "{}" && !this.IsNullable() {
			defaultValue = fmt.Sprintf("new %s()", snakeToPascal(this.GetColumn()))
		}
		return fmt.Sprintf("%s %s", snakeToPascal(this.GetColumn()), nullableString), defaultValue
	default:
		return "string | null", "null"
	}
}

func (this *StructField) IsNullable() bool {
	tags := this.GetTags()
	if nullableTag, ok := tags["null"]; ok {
		return nullableTag == "true"
	}
	return false
}

func (this *StructField) GetDefault() string {
	tags := this.GetTags()
	return tags["default"]
}

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
