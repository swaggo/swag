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

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
)

var _ FieldParser = &tagBaseFieldParser{p: nil, field: nil, tag: ""}

const (
	requiredLabel    = "required"
	swaggerTypeTag   = "swaggertype"
	swaggerIgnoreTag = "swaggerignore"
)

type tagBaseFieldParser struct {
	p     *Parser
	field *ast.Field
	tag   reflect.StructTag
}

func newTagBaseFieldParser(p *Parser, field *ast.Field) FieldParser {
	ps := tagBaseFieldParser{
		p:     p,
		field: field,
		tag:   "",
	}
	if ps.field.Tag != nil {
		ps.tag = reflect.StructTag(strings.ReplaceAll(field.Tag.Value, "`", ""))
	}

	return &ps
}

func (ps *tagBaseFieldParser) ShouldSkip() (bool, error) {
	// Skip non-exported fields.
	if !ast.IsExported(ps.field.Names[0].Name) {
		return true, nil
	}

	if ps.field.Tag == nil {
		return false, nil
	}

	ignoreTag := ps.tag.Get(swaggerIgnoreTag)
	if strings.EqualFold(ignoreTag, "true") {
		return true, nil
	}

	// json:"tag,hoge"
	name := strings.TrimSpace(strings.Split(ps.tag.Get(jsonTag), ",")[0])
	if name == "-" {
		return true, nil
	}

	return false, nil
}

func (ps *tagBaseFieldParser) FieldName() (string, error) {
	var name string
	if ps.field.Tag != nil {
		// json:"tag,hoge"
		name = strings.TrimSpace(strings.Split(ps.tag.Get(jsonTag), ",")[0])

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
	var (
		runes  = []rune(in)
		length = len(runes)
		out    []rune
	)

	for idx := 0; idx < length; idx++ {
		if idx > 0 && unicode.IsUpper(runes[idx]) &&
			((idx+1 < length && unicode.IsLower(runes[idx+1])) || unicode.IsLower(runes[idx-1])) {
			out = append(out, '_')
		}

		out = append(out, unicode.ToLower(runes[idx]))
	}

	return string(out)
}

func toLowerCamelCase(in string) string {
	runes := []rune(in)

	var (
		out  []rune
		flag bool
	)

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

	typeTag := ps.tag.Get(swaggerTypeTag)
	if typeTag != "" {
		return BuildCustomSchema(strings.Split(typeTag, ","))
	}

	return nil, nil
}

type structField struct {
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
	enums        []interface{}
	enumVarNames []interface{}
	unique       bool
}

// splitNotWrapped slices s into all substrings separated by sep if sep is not
// wrapped by brackets and returns a slice of the substrings between those separators.
func splitNotWrapped(s string, sep rune) []string {
	openCloseMap := map[rune]rune{
		'(': ')',
		'[': ']',
		'{': '}',
	}

	var (
		result    = make([]string, 0)
		current   = strings.Builder{}
		openCount = 0
		openChar  rune
	)

	for _, char := range s {
		switch {
		case openChar == 0 && openCloseMap[char] != 0:
			openChar = char

			openCount++

			current.WriteRune(char)
		case char == openChar:
			openCount++

			current.WriteRune(char)
		case openCount > 0 && char == openCloseMap[openChar]:
			openCount--

			current.WriteRune(char)
		case openCount == 0 && char == sep:
			result = append(result, current.String())

			openChar = 0

			current = strings.Builder{}
		default:
			current.WriteRune(char)
		}
	}

	if current.String() != "" {
		result = append(result, current.String())
	}

	return result
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

	field := &structField{
		schemaType: types[0],
		formatType: ps.tag.Get(formatTag),
	}

	if len(types) > 1 && (types[0] == ARRAY || types[0] == OBJECT) {
		field.arrayType = types[1]
	}

	jsonTag := ps.tag.Get(jsonTag)

	bindingTag := ps.tag.Get(bindingTag)
	if bindingTag != "" {
		ps.parseValidTags(bindingTag, field)
	}

	validateTag := ps.tag.Get(validateTag)
	if validateTag != "" {
		ps.parseValidTags(validateTag, field)
	}

	enumsTag := ps.tag.Get(enumsTag)
	if enumsTag != "" {
		enumType := field.schemaType
		if field.schemaType == ARRAY {
			enumType = field.arrayType
		}

		field.enums = nil

		for _, e := range strings.Split(enumsTag, ",") {
			value, err := defineType(enumType, e)
			if err != nil {
				return err
			}

			field.enums = append(field.enums, value)
		}
	}

	if IsNumericType(field.schemaType) || IsNumericType(field.arrayType) {
		maximum, err := getFloatTag(ps.tag, maximumTag)
		if err != nil {
			return err
		}

		if maximum != nil {
			field.maximum = maximum
		}

		minimum, err := getFloatTag(ps.tag, minimumTag)
		if err != nil {
			return err
		}

		if minimum != nil {
			field.minimum = minimum
		}

		multipleOf, err := getFloatTag(ps.tag, multipleOfTag)
		if err != nil {
			return err
		}

		if multipleOf != nil {
			field.multipleOf = multipleOf
		}
	}

	if field.schemaType == STRING || field.arrayType == STRING {
		maxLength, err := getIntTag(ps.tag, "maxLength")
		if err != nil {
			return err
		}

		if maxLength != nil {
			field.maxLength = maxLength
		}

		minLength, err := getIntTag(ps.tag, "minLength")
		if err != nil {
			return err
		}

		if minLength != nil {
			field.minLength = minLength
		}
	}

	// json:"name,string" or json:",string"
	exampleTag, ok := ps.tag.Lookup(exampleTag)
	if ok {
		field.exampleValue = exampleTag

		if !strings.Contains(jsonTag, ",string") {
			example, err := defineTypeOfExample(field.schemaType, field.arrayType, exampleTag)
			if err != nil {
				return err
			}

			field.exampleValue = example
		}
	}

	// perform this after setting everything else (min, max, etc...)
	if strings.Contains(jsonTag, ",string") {
		// @encoding/json: "It applies only to fields of string, floating point, integer, or boolean types."
		defaultValues := map[string]string{
			// Zero Values as string
			STRING:  "",
			INTEGER: "0",
			BOOLEAN: "false",
			NUMBER:  "0",
		}

		defaultValue, ok := defaultValues[field.schemaType]
		if ok {
			field.schemaType = STRING
			*schema = *PrimitiveSchema(field.schemaType)

			if field.exampleValue == nil {
				// if exampleValue is not defined by the user,
				// we will force an example with a correct value
				// (eg: int->"0", bool:"false")
				field.exampleValue = defaultValue
			}
		}
	}

	if ps.field.Doc != nil {
		schema.Description = strings.TrimSpace(ps.field.Doc.Text())
	}

	if schema.Description == "" && ps.field.Comment != nil {
		schema.Description = strings.TrimSpace(ps.field.Comment.Text())
	}

	schema.ReadOnly = ps.tag.Get(readOnlyTag) == "true"

	if !reflect.ValueOf(schema.Ref).IsZero() && schema.ReadOnly {
		schema.AllOf = []spec.Schema{*spec.RefSchema(schema.Ref.String())}
		schema.Ref = spec.Ref{
			Ref: jsonreference.Ref{
				HasFullURL:      false,
				HasURLPathOnly:  false,
				HasFragmentOnly: false,
				HasFileScheme:   false,
				HasFullFilePath: false,
			},
		} // clear out existing ref
	}

	defaultTag := ps.tag.Get(defaultTag)
	if defaultTag != "" {
		value, err := defineType(field.schemaType, defaultTag)
		if err != nil {
			return err
		}

		schema.Default = value
	}

	schema.Example = field.exampleValue

	if field.schemaType != ARRAY {
		schema.Format = field.formatType
	}

	extensionsTag := ps.tag.Get(extensionsTag)
	if extensionsTag != "" {
		schema.Extensions = map[string]interface{}{}

		for _, val := range splitNotWrapped(extensionsTag, ',') {
			parts := strings.SplitN(val, "=", 2)
			if len(parts) == 2 {
				schema.Extensions[parts[0]] = parts[1]
			} else {
				if len(parts[0]) > 0 && string(parts[0][0]) == "!" {
					schema.Extensions[parts[0][1:]] = false
				} else {
					schema.Extensions[parts[0]] = true
				}
			}
		}
	}

	varnamesTag := ps.tag.Get("x-enum-varnames")
	if varnamesTag != "" {
		if schema.Extensions == nil {
			schema.Extensions = map[string]interface{}{}
		}

		varNames := strings.Split(varnamesTag, ",")
		if len(varNames) != len(field.enums) {
			return fmt.Errorf("invalid count of x-enum-varnames. expected %d, got %d", len(field.enums), len(varNames))
		}

		field.enumVarNames = nil

		for _, v := range varNames {
			field.enumVarNames = append(field.enumVarNames, v)
		}

		schema.Extensions["x-enum-varnames"] = field.enumVarNames
	}

	eleSchema := schema

	if field.schemaType == ARRAY {
		// For Array only
		schema.MaxItems = field.maxItems
		schema.MinItems = field.minItems
		schema.UniqueItems = field.unique

		eleSchema = schema.Items.Schema
		eleSchema.Format = field.formatType
	}

	eleSchema.Maximum = field.maximum
	eleSchema.Minimum = field.minimum
	eleSchema.MultipleOf = field.multipleOf
	eleSchema.MaxLength = field.maxLength
	eleSchema.MinLength = field.minLength
	eleSchema.Enum = field.enums

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
			if val == requiredLabel {
				return true, nil
			}
		}
	}

	validateTag := ps.tag.Get(validateTag)
	if validateTag != "" {
		for _, val := range strings.Split(validateTag, ",") {
			if val == requiredLabel {
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
			kv       = strings.Split(val, "=")
		)

		switch len(kv) {
		case 1:
			valKey = kv[0]
		case 2:
			valKey = kv[0]
			valValue = kv[1]
		default:
			continue
		}

		valValue = strings.ReplaceAll(strings.ReplaceAll(valValue, utf8HexComma, ","), utf8Pipe, "|")

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

func parseOneOfParam2(param string) []string {
	oneofValsCacheRWLock.RLock()
	values, ok := oneofValsCache[param]
	oneofValsCacheRWLock.RUnlock()

	if !ok {
		oneofValsCacheRWLock.Lock()
		values = splitParamsRegex.FindAllString(param, -1)

		for i := 0; i < len(values); i++ {
			values[i] = strings.ReplaceAll(values[i], "'", "")
		}

		oneofValsCache[param] = values

		oneofValsCacheRWLock.Unlock()
	}

	return values
}

// ---
