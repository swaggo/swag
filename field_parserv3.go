package swag

import (
	"fmt"
	"go/ast"
	"reflect"
	"strconv"
	"strings"

	"github.com/sv-tools/openapi/spec"
)

type structFieldV3 struct {
	schemaType   string
	arrayType    string
	formatType   string
	maximum      *int
	minimum      *int
	multipleOf   *int
	maxLength    *int
	minLength    *int
	maxItems     *int
	minItems     *int
	exampleValue interface{}
	enums        []interface{}
	enumVarNames []interface{}
	unique       bool
	pattern      string
}

func (sf *structFieldV3) setOneOf(valValue string) {
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

func (sf *structFieldV3) setMin(valValue string) {
	value, err := strconv.Atoi(valValue)
	if err != nil {
		return
	}

	switch sf.schemaType {
	case INTEGER, NUMBER:
		sf.minimum = &value
	case STRING:
		sf.minLength = &value
	case ARRAY:
		sf.minItems = &value
	}
}

func (sf *structFieldV3) setMax(valValue string) {
	value, err := strconv.Atoi(valValue)
	if err != nil {
		return
	}

	switch sf.schemaType {
	case INTEGER, NUMBER:
		sf.maximum = &value
	case STRING:
		sf.maxLength = &value
	case ARRAY:
		sf.maxItems = &value
	}
}

type tagBaseFieldParserV3 struct {
	p     *Parser
	file  *ast.File
	field *ast.Field
	tag   reflect.StructTag
}

func newTagBaseFieldParserV3(p *Parser, file *ast.File, field *ast.Field) FieldParserV3 {
	fieldParser := tagBaseFieldParserV3{
		p:     p,
		file:  file,
		field: field,
		tag:   "",
	}
	if fieldParser.field.Tag != nil {
		fieldParser.tag = reflect.StructTag(strings.ReplaceAll(field.Tag.Value, "`", ""))
	}

	return &fieldParser
}

func (ps *tagBaseFieldParserV3) CustomSchema() (*spec.RefOrSpec[spec.Schema], error) {
	if ps.field.Tag == nil {
		return nil, nil
	}

	typeTag := ps.tag.Get(swaggerTypeTag)
	if typeTag != "" {
		return BuildCustomSchemaV3(strings.Split(typeTag, ","))
	}

	return nil, nil
}

// ComplementSchema complement schema with field properties
func (ps *tagBaseFieldParserV3) ComplementSchema(schema *spec.RefOrSpec[spec.Schema]) error {
	if schema.Spec == nil {
		componentSchema := ps.p.openAPI.Components.Spec.Schemas[strings.ReplaceAll(schema.Ref.Ref, "#/components/schemas/", "")]
		if componentSchema == nil {
			return fmt.Errorf("could not resolve schema for ref %s", schema.Ref.Ref)
		}
		schema = componentSchema
	}

	types := ps.p.GetSchemaTypePathV3(schema, 2)
	if len(types) == 0 {
		return fmt.Errorf("invalid type for field: %s", ps.field.Names[0])
	}

	if schema.Ref != nil { //IsRefSchema(schema)
		// TODO fetch existing schema from components
		var newSchema = spec.Schema{}
		err := ps.complementSchema(&newSchema, types)
		if err != nil {
			return err
		}
		if !reflect.ValueOf(newSchema).IsZero() {
			newSchema.AllOf = []*spec.RefOrSpec[spec.Schema]{{Spec: schema.Spec}}
			*schema = spec.RefOrSpec[spec.Schema]{Spec: &newSchema}
		}
		return nil
	}

	return ps.complementSchema(schema.Spec, types)
}

// complementSchema complement schema with field properties
func (ps *tagBaseFieldParserV3) complementSchema(schema *spec.Schema, types []string) error {
	if ps.field.Tag == nil {
		if ps.field.Doc != nil {
			schema.Description = strings.TrimSpace(ps.field.Doc.Text())
		}

		if schema.Description == "" && ps.field.Comment != nil {
			schema.Description = strings.TrimSpace(ps.field.Comment.Text())
		}

		return nil
	}

	field := &structFieldV3{
		schemaType: types[0],
		formatType: ps.tag.Get(formatTag),
	}

	if len(types) > 1 && (types[0] == ARRAY || types[0] == OBJECT) {
		field.arrayType = types[1]
	}

	jsonTagValue := ps.tag.Get(jsonTag)

	bindingTagValue := ps.tag.Get(bindingTag)
	if bindingTagValue != "" {
		field.parseValidTags(bindingTagValue)
	}

	validateTagValue := ps.tag.Get(validateTag)
	if validateTagValue != "" {
		field.parseValidTags(validateTagValue)
	}

	enumsTagValue := ps.tag.Get(enumsTag)
	if enumsTagValue != "" {
		err := field.parseEnumTags(enumsTagValue)
		if err != nil {
			return err
		}
	}

	if IsNumericType(field.schemaType) || IsNumericType(field.arrayType) {
		maximum, err := getIntTagV3(ps.tag, maximumTag)
		if err != nil {
			return err
		}

		if maximum != nil {
			field.maximum = maximum
		}

		minimum, err := getIntTagV3(ps.tag, minimumTag)
		if err != nil {
			return err
		}

		if minimum != nil {
			field.minimum = minimum
		}

		multipleOf, err := getIntTagV3(ps.tag, multipleOfTag)
		if err != nil {
			return err
		}

		if multipleOf != nil {
			field.multipleOf = multipleOf
		}
	}

	if field.schemaType == STRING || field.arrayType == STRING {
		maxLength, err := getIntTagV3(ps.tag, maxLengthTag)
		if err != nil {
			return err
		}

		if maxLength != nil {
			field.maxLength = maxLength
		}

		minLength, err := getIntTagV3(ps.tag, minLengthTag)
		if err != nil {
			return err
		}

		if minLength != nil {
			field.minLength = minLength
		}

		pattern, ok := ps.tag.Lookup(patternTag)
		if ok {
			field.pattern = pattern
		}
	}

	// json:"name,string" or json:",string"
	exampleTagValue, ok := ps.tag.Lookup(exampleTag)
	if ok {
		field.exampleValue = exampleTagValue

		if !strings.Contains(jsonTagValue, ",string") {
			example, err := defineTypeOfExample(field.schemaType, field.arrayType, exampleTagValue)
			if err != nil {
				return err
			}

			field.exampleValue = example
		}
	}

	// perform this after setting everything else (min, max, etc...)
	if strings.Contains(jsonTagValue, ",string") {
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
			*schema = *PrimitiveSchemaV3(field.schemaType).Spec

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

	defaultTagValue := ps.tag.Get(defaultTag)
	if defaultTagValue != "" {
		value, err := defineType(field.schemaType, defaultTagValue)
		if err != nil {
			return err
		}

		schema.Default = value
	}

	schema.Example = field.exampleValue

	if field.schemaType != ARRAY {
		schema.Format = field.formatType
	}

	extensionsTagValue := ps.tag.Get(extensionsTag)
	if extensionsTagValue != "" {
		schema.Extensions = setExtensionParam(extensionsTagValue)
	}

	varNamesTag := ps.tag.Get("x-enum-varnames")
	if varNamesTag != "" {
		varNames := strings.Split(varNamesTag, ",")
		if len(varNames) != len(field.enums) {
			return fmt.Errorf("invalid count of x-enum-varnames. expected %d, got %d", len(field.enums), len(varNames))
		}

		field.enumVarNames = nil

		for _, v := range varNames {
			field.enumVarNames = append(field.enumVarNames, v)
		}

		if field.schemaType == ARRAY {
			// Add the var names in the items schema
			if schema.Items.Schema.Spec.Extensions == nil {
				schema.Items.Schema.Spec.Extensions = map[string]interface{}{}
			}
			schema.Items.Schema.Spec.Extensions[enumVarNamesExtension] = field.enumVarNames
		} else {
			// Add to top level schema
			if schema.Extensions == nil {
				schema.Extensions = map[string]interface{}{}
			}
			schema.Extensions[enumVarNamesExtension] = field.enumVarNames
		}
	}

	var oneOfSchemas []*spec.RefOrSpec[spec.Schema]
	oneOfTagValue := ps.tag.Get(oneOfTag)
	if oneOfTagValue != "" {
		oneOfTypes := strings.Split((oneOfTagValue), ",")
		for _, oneOfType := range oneOfTypes {
			oneOfSchema, err := ps.p.getTypeSchemaV3(oneOfType, ps.file, true)
			if err != nil {
				return fmt.Errorf("can't find oneOf type %q: %v", oneOfType, err)
			}
			oneOfSchemas = append(oneOfSchemas, oneOfSchema)
		}
	}

	elemSchema := schema

	if field.schemaType == ARRAY {
		// For Array only
		schema.MaxItems = field.maxItems
		schema.MinItems = field.minItems
		schema.UniqueItems = &field.unique

		elemSchema = schema.Items.Schema.Spec
		if elemSchema == nil {
			elemSchema = ps.p.getSchemaByRef(schema.Items.Schema.Ref)
		}

		elemSchema.Format = field.formatType
	}

	elemSchema.Maximum = field.maximum
	elemSchema.Minimum = field.minimum
	elemSchema.MultipleOf = field.multipleOf
	elemSchema.MaxLength = field.maxLength
	elemSchema.MinLength = field.minLength
	elemSchema.Enum = append(elemSchema.Enum, field.enums...)
	elemSchema.Pattern = field.pattern
	elemSchema.OneOf = oneOfSchemas

	return nil
}

func getIntTagV3(structTag reflect.StructTag, tagName string) (*int, error) {
	strValue := structTag.Get(tagName)
	if strValue == "" {
		return nil, nil
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		return nil, fmt.Errorf("can't parse numeric value of %q tag: %v", tagName, err)
	}

	return &value, nil
}

func (sf *structFieldV3) parseValidTags(validTag string) {

	// `validate:"required,max=10,min=1"`
	// ps. required checked by IsRequired().
	for _, val := range strings.Split(validTag, ",") {
		var (
			valValue string
			keyVal   = strings.Split(val, "=")
		)

		switch len(keyVal) {
		case 1:
		case 2:
			valValue = strings.ReplaceAll(strings.ReplaceAll(keyVal[1], utf8HexComma, ","), utf8Pipe, "|")
		default:
			continue
		}

		switch keyVal[0] {
		case "max", "lte":
			sf.setMax(valValue)
		case "min", "gte":
			sf.setMin(valValue)
		case "oneof":
			if strings.Contains(validTag, "swaggerIgnore") {
				continue
			}

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

func (sf *structFieldV3) parseEnumTags(enumTag string) error {
	enumType := sf.schemaType
	if sf.schemaType == ARRAY {
		enumType = sf.arrayType
	}

	sf.enums = nil

	for _, e := range strings.Split(enumTag, ",") {
		value, err := defineType(enumType, e)
		if err != nil {
			return err
		}

		sf.enums = append(sf.enums, value)
	}

	return nil
}

func (ps *tagBaseFieldParserV3) ShouldSkip() bool {
	// Skip non-exported fields.
	if ps.field.Names != nil && !ast.IsExported(ps.field.Names[0].Name) {
		return true
	}

	if ps.field.Tag == nil {
		return false
	}

	ignoreTag := ps.tag.Get(swaggerIgnoreTag)
	if strings.EqualFold(ignoreTag, "true") {
		return true
	}

	// json:"tag,hoge"
	name := ps.JsonName()
	if name == "" {
		return true
	}

	return false
}

func (ps *tagBaseFieldParserV3) FieldName() (string, error) {
	var name string

	// json:"tag,hoge"
	name = ps.JsonName()
	if name != "" {
		return name, nil
	}

	// use "form" tag over json tag
	name = ps.FormName()
	if name != "" {
		return name, nil
	}

	if ps.field.Names == nil {
		return "", nil
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

func (ps *tagBaseFieldParserV3) FormName() string {
	if ps.field.Tag != nil {
		name := strings.TrimSpace(strings.Split(ps.tag.Get(formTag), ",")[0])
		if name != "-" {
			return name
		}
	}
	return ""
}

func (ps *tagBaseFieldParserV3) JsonName() string {
	if ps.field.Tag != nil {
		name := strings.TrimSpace(strings.Split(ps.tag.Get(jsonTag), ",")[0])
		if name != "-" {
			return name
		}
	}
	return ""
}

func (ps *tagBaseFieldParserV3) IsRequired() (bool, error) {
	if ps.field.Tag == nil {
		return false, nil
	}

	bindingTag := ps.tag.Get(bindingTag)
	if bindingTag != "" {
		for _, val := range strings.Split(bindingTag, ",") {
			switch val {
			case requiredLabel:
				return true, nil
			case optionalLabel:
				return false, nil
			}
		}
	}

	validateTag := ps.tag.Get(validateTag)
	if validateTag != "" {
		for _, val := range strings.Split(validateTag, ",") {
			switch val {
			case requiredLabel:
				return true, nil
			case optionalLabel:
				return false, nil
			}
		}
	}

	return ps.p.RequiredByDefault, nil
}
