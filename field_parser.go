package swag

import (
	"fmt"
	"go/ast"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/go-openapi/spec"
)

var _ FieldParser = &tagBaseFieldParser{}

type tagBaseFieldParser struct {
	p     *Parser
	field *ast.Field
	tag   reflect.StructTag
}

func newTagBaseFieldParser(p *Parser, field *ast.Field) FieldParser {
	ps := &tagBaseFieldParser{
		p:     p,
		field: field,
	}
	if ps.field.Tag != nil {
		ps.tag = reflect.StructTag(strings.Replace(field.Tag.Value, "`", "", -1))
	}

	return ps
}

func (ps *tagBaseFieldParser) ShouldSkip() (bool, error) {
	// Skip non-exported fields.
	if !ast.IsExported(ps.field.Names[0].Name) {
		return true, nil
	}

	if ps.field.Tag == nil {
		return false, nil
	}

	ignoreTag := ps.tag.Get("swaggerignore")
	if strings.EqualFold(ignoreTag, "true") {
		return true, nil
	}

	// json:"tag,hoge"
	name := strings.TrimSpace(strings.Split(ps.tag.Get("json"), ",")[0])
	if name == "-" {
		return true, nil
	}

	return false, nil
}

func (ps *tagBaseFieldParser) FieldName() (string, error) {
	var name string
	if ps.field.Tag != nil {
		// json:"tag,hoge"
		name = strings.TrimSpace(strings.Split(ps.tag.Get("json"), ",")[0])

		if name != "" {
			return name, nil
		}
	}

	switch ps.p.PropNamingStrategy {
	case SnakeCase:
		return toSnakeCase(ps.field.Names[0].Name), nil
	case PascalCase:
		return ps.field.Names[0].Name, nil
	default:
		return toLowerCamelCase(ps.field.Names[0].Name), nil
	}
}

func toSnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) &&
			((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

func toLowerCamelCase(in string) string {
	runes := []rune(in)

	var out []rune
	flag := false
	for i, curr := range runes {
		if (i == 0 && unicode.IsUpper(curr)) || (flag && unicode.IsUpper(curr)) {
			out = append(out, unicode.ToLower(curr))
			flag = true
		} else {
			out = append(out, curr)
			flag = false
		}
	}

	return string(out)
}

func (ps *tagBaseFieldParser) CustomSchema() (*spec.Schema, error) {
	if ps.field.Tag == nil {
		return nil, nil
	}

	typeTag := ps.tag.Get("swaggertype")
	if typeTag != "" {
		return BuildCustomSchema(strings.Split(typeTag, ","))
	}

	return nil, nil
}

type structField struct {
	desc         string
	schemaType   string
	arrayType    string
	formatType   string
	maximum      *float64
	minimum      *float64
	multipleOf   *float64
	maxLength    *int64
	minLength    *int64
	maxItems     *int64
	minItems     *int64
	exampleValue interface{}
	defaultValue interface{}
	extensions   map[string]interface{}
	enums        []interface{}
	enumVarNames []interface{}
	readOnly     bool
	unique       bool
}

func (ps *tagBaseFieldParser) ComplementSchema(schema *spec.Schema) error {
	types := ps.p.GetSchemaTypePath(schema, 2)
	if len(types) == 0 {
		return fmt.Errorf("invalid type for field: %s", ps.field.Names[0])
	}

	if ps.field.Tag == nil {
		if ps.field.Doc != nil {
			schema.Description = strings.TrimSpace(ps.field.Doc.Text())
		}
		if schema.Description == "" && ps.field.Comment != nil {
			schema.Description = strings.TrimSpace(ps.field.Comment.Text())
		}
		return nil
	}

	structField := &structField{
		schemaType: types[0],
		formatType: ps.tag.Get(formatTag),
		readOnly:   ps.tag.Get(readOnlyTag) == "true",
	}

	if len(types) > 1 && (types[0] == ARRAY || types[0] == OBJECT) {
		structField.arrayType = types[1]
	}

	if ps.field.Doc != nil {
		structField.desc = strings.TrimSpace(ps.field.Doc.Text())
	}
	if structField.desc == "" && ps.field.Comment != nil {
		structField.desc = strings.TrimSpace(ps.field.Comment.Text())
	}

	jsonTag := ps.tag.Get(jsonTag)
	// json:"name,string" or json:",string"

	exampleTag, ok := ps.tag.Lookup(exampleTag)
	if ok {
		structField.exampleValue = exampleTag
		if !strings.Contains(jsonTag, ",string") {
			example, err := defineTypeOfExample(structField.schemaType, structField.arrayType, exampleTag)
			if err != nil {
				return err
			}
			structField.exampleValue = example
		}
	}

	bindingTag := ps.tag.Get(bindingTag)
	if bindingTag != "" {
		ps.parseValidTags(bindingTag, structField)
	}

	validateTag := ps.tag.Get(validateTag)
	if validateTag != "" {
		ps.parseValidTags(validateTag, structField)
	}

	extensionsTag := ps.tag.Get(extensionsTag)
	if extensionsTag != "" {
		structField.extensions = map[string]interface{}{}
		for _, val := range strings.Split(extensionsTag, ",") {
			parts := strings.SplitN(val, "=", 2)
			if len(parts) == 2 {
				structField.extensions[parts[0]] = parts[1]
			} else {
				if len(parts[0]) > 0 && string(parts[0][0]) == "!" {
					structField.extensions[parts[0][1:]] = false
				} else {
					structField.extensions[parts[0]] = true
				}
			}
		}
	}

	enumsTag := ps.tag.Get(enumsTag)
	if enumsTag != "" {
		enumType := structField.schemaType
		if structField.schemaType == ARRAY {
			enumType = structField.arrayType
		}

		structField.enums = nil
		for _, e := range strings.Split(enumsTag, ",") {
			value, err := defineType(enumType, e)
			if err != nil {
				return err
			}
			structField.enums = append(structField.enums, value)
		}
	}
	varnamesTag := ps.tag.Get("x-enum-varnames")
	if varnamesTag != "" {
		if structField.extensions == nil {
			structField.extensions = map[string]interface{}{}
		}
		varNames := strings.Split(varnamesTag, ",")
		if len(varNames) != len(structField.enums) {
			return fmt.Errorf("invalid count of x-enum-varnames. expected %d, got %d", len(structField.enums), len(varNames))
		}
		structField.enumVarNames = nil
		for _, v := range varNames {
			structField.enumVarNames = append(structField.enumVarNames, v)
		}
		structField.extensions["x-enum-varnames"] = structField.enumVarNames
	}
	defaultTag := ps.tag.Get(defaultTag)
	if defaultTag != "" {
		value, err := defineType(structField.schemaType, defaultTag)
		if err != nil {
			return err
		}
		structField.defaultValue = value
	}

	if IsNumericType(structField.schemaType) || IsNumericType(structField.arrayType) {
		maximum, err := getFloatTag(ps.tag, maximumTag)
		if err != nil {
			return err
		}
		if maximum != nil {
			structField.maximum = maximum
		}

		minimum, err := getFloatTag(ps.tag, minimumTag)
		if err != nil {
			return err
		}
		if minimum != nil {
			structField.minimum = minimum
		}

		multipleOf, err := getFloatTag(ps.tag, multipleOfTag)
		if err != nil {
			return err
		}
		if multipleOf != nil {
			structField.multipleOf = multipleOf
		}
	}

	if structField.schemaType == STRING || structField.arrayType == STRING {
		maxLength, err := getIntTag(ps.tag, "maxLength")
		if err != nil {
			return err
		}
		if maxLength != nil {
			structField.maxLength = maxLength
		}

		minLength, err := getIntTag(ps.tag, "minLength")
		if err != nil {
			return err
		}
		if minLength != nil {
			structField.minLength = minLength
		}
	}

	// perform this after setting everything else (min, max, etc...)
	if strings.Contains(jsonTag, ",string") { // @encoding/json: "It applies only to fields of string, floating point, integer, or boolean types."
		defaultValues := map[string]string{
			// Zero Values as string
			STRING:  "",
			INTEGER: "0",
			BOOLEAN: "false",
			NUMBER:  "0",
		}

		defaultValue, ok := defaultValues[structField.schemaType]
		if ok {
			structField.schemaType = STRING

			if structField.exampleValue == nil {
				// if exampleValue is not defined by the user,
				// we will force an example with a correct value
				// (eg: int->"0", bool:"false")
				structField.exampleValue = defaultValue
			}
		}
	}

	if structField.schemaType == STRING && types[0] != STRING {
		*schema = *PrimitiveSchema(structField.schemaType)
	}

	schema.Description = structField.desc
	schema.ReadOnly = structField.readOnly
	if !reflect.ValueOf(schema.Ref).IsZero() && schema.ReadOnly {
		schema.AllOf = []spec.Schema{*spec.RefSchema(schema.Ref.String())}
		schema.Ref = spec.Ref{} // clear out existing ref
	}
	schema.Default = structField.defaultValue
	schema.Example = structField.exampleValue
	if structField.schemaType != ARRAY {
		schema.Format = structField.formatType
	}
	schema.Extensions = structField.extensions
	eleSchema := schema
	if structField.schemaType == ARRAY {
		// For Array only
		schema.MaxItems = structField.maxItems
		schema.MinItems = structField.minItems
		schema.UniqueItems = structField.unique

		eleSchema = schema.Items.Schema
		eleSchema.Format = structField.formatType
	}
	eleSchema.Maximum = structField.maximum
	eleSchema.Minimum = structField.minimum
	eleSchema.MultipleOf = structField.multipleOf
	eleSchema.MaxLength = structField.maxLength
	eleSchema.MinLength = structField.minLength
	eleSchema.Enum = structField.enums
	return nil
}

func getFloatTag(structTag reflect.StructTag, tagName string) (*float64, error) {
	strValue := structTag.Get(tagName)
	if strValue == "" {
		return nil, nil
	}

	value, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		return nil, fmt.Errorf("can't parse numeric value of %q tag: %v", tagName, err)
	}

	return &value, nil
}

func getIntTag(structTag reflect.StructTag, tagName string) (*int64, error) {
	strValue := structTag.Get(tagName)
	if strValue == "" {
		return nil, nil
	}

	value, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("can't parse numeric value of %q tag: %v", tagName, err)
	}

	return &value, nil
}

func (ps *tagBaseFieldParser) IsRequired() (bool, error) {
	if ps.field.Tag == nil {
		return false, nil
	}

	bindingTag := ps.tag.Get(bindingTag)
	if bindingTag != "" {
		for _, val := range strings.Split(bindingTag, ",") {
			if val == "required" {
				return true, nil
			}
		}
	}

	validateTag := ps.tag.Get(validateTag)
	if validateTag != "" {
		for _, val := range strings.Split(validateTag, ",") {
			if val == "required" {
				return true, nil
			}
		}
	}

	return false, nil
}

func (ps *tagBaseFieldParser) parseValidTags(validTag string, sf *structField) {
	// `validate:"required,max=10,min=1"`
	// ps. required checked by IsRequired().
	for _, val := range strings.Split(validTag, ",") {
		var (
			valKey   string
			valValue string
		)
		kv := strings.Split(val, "=")
		switch len(kv) {
		case 1:
			valKey = kv[0]
		case 2:
			valKey = kv[0]
			valValue = kv[1]
		default:
			continue
		}
		valValue = strings.Replace(strings.Replace(valValue, utf8HexComma, ",", -1), utf8Pipe, "|", -1)

		switch valKey {
		case "max", "lte":
			sf.setMax(valValue)
		case "min", "gte":
			sf.setMin(valValue)
		case "oneof":
			sf.setOneOf(valValue)
		case "unique":
			if sf.schemaType == ARRAY {
				sf.unique = true
			}
		case "dive":
			// ignore dive
			return
		default:
			continue
		}
	}
}

func (sf *structField) setOneOf(valValue string) {
	if len(sf.enums) != 0 {
		return
	}

	enumType := sf.schemaType
	if sf.schemaType == ARRAY {
		enumType = sf.arrayType
	}

	valValues := parseOneOfParam2(valValue)
	for i := range valValues {
		value, err := defineType(enumType, valValues[i])
		if err != nil {
			continue
		}
		sf.enums = append(sf.enums, value)
	}
}

func (sf *structField) setMin(valValue string) {
	value, err := strconv.ParseFloat(valValue, 64)
	if err != nil {
		return
	}
	switch sf.schemaType {
	case INTEGER, NUMBER:
		sf.minimum = &value
	case STRING:
		intValue := int64(value)
		sf.minLength = &intValue
	case ARRAY:
		intValue := int64(value)
		sf.minItems = &intValue
	}
}

func (sf *structField) setMax(valValue string) {
	value, err := strconv.ParseFloat(valValue, 64)
	if err != nil {
		return
	}
	switch sf.schemaType {
	case INTEGER, NUMBER:
		sf.maximum = &value
	case STRING:
		intValue := int64(value)
		sf.maxLength = &intValue
	case ARRAY:
		intValue := int64(value)
		sf.maxItems = &intValue
	}
}

const (
	utf8HexComma = "0x2C"
	utf8Pipe     = "0x7C"
)

// These code copy from
// https://github.com/go-playground/validator/blob/d4271985b44b735c6f76abc7a06532ee997f9476/baked_in.go#L207
// ---
var oneofValsCache = map[string][]string{}
var oneofValsCacheRWLock = sync.RWMutex{}
var splitParamsRegex = regexp.MustCompile(`'[^']*'|\S+`)

func parseOneOfParam2(s string) []string {
	oneofValsCacheRWLock.RLock()
	values, ok := oneofValsCache[s]
	oneofValsCacheRWLock.RUnlock()
	if !ok {
		oneofValsCacheRWLock.Lock()
		values = splitParamsRegex.FindAllString(s, -1)
		for i := 0; i < len(values); i++ {
			values[i] = strings.Replace(values[i], "'", "", -1)
		}
		oneofValsCache[s] = values
		oneofValsCacheRWLock.Unlock()
	}
	return values
}

// ---
