package swag

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/sv-tools/openapi/spec"
)

// FieldParserFactoryV3 create FieldParser.
type FieldParserFactoryV3 func(ps *Parser, file *ast.File, field *ast.Field) FieldParserV3

// FieldParserV3 parse struct field.
type FieldParserV3 interface {
	ShouldSkip() bool
	FieldName() (string, error)
	FormName() string
	CustomSchema() (*spec.RefOrSpec[spec.Schema], error)
	ComplementSchema(schema *spec.RefOrSpec[spec.Schema]) error
	IsRequired() (bool, error)
}

// GetOpenAPI returns *spec.OpenAPI which is the root document object for the API specification.
func (p *Parser) GetOpenAPI() *spec.OpenAPI {
	return p.openAPI
}

var (
	serversURLPattern       = regexp.MustCompile(`\{([^}]+)\}`)
	serversVariablesPattern = regexp.MustCompile(`^(\w+)\s+(.+)$`)
)

func (p *Parser) parseGeneralAPIInfoV3(comments []string) error {
	previousAttribute := ""

	// parsing classic meta data model
	for line := 0; line < len(comments); line++ {
		commentLine := comments[line]
		commentLine = strings.TrimSpace(commentLine)
		if len(commentLine) == 0 {
			continue
		}
		fields := FieldsByAnySpace(commentLine, 2)

		attribute := fields[0]
		var value string
		if len(fields) > 1 {
			value = fields[1]
		}

		switch attr := strings.ToLower(attribute); attr {
		case versionAttr, titleAttr, tosAttr, licNameAttr, licURLAttr, conNameAttr, conURLAttr, conEmailAttr:
			setspecInfo(p.openAPI, attr, value)
		case descriptionAttr:
			if previousAttribute == attribute {
				p.openAPI.Info.Spec.Description += "\n" + value

				continue
			}

			setspecInfo(p.openAPI, attr, value)
		case descriptionMarkdownAttr:
			commentInfo, err := getMarkdownForTag("api", p.markdownFileDir)
			if err != nil {
				return err
			}

			setspecInfo(p.openAPI, attr, string(commentInfo))
		case "@host":
			if len(p.openAPI.Servers) == 0 {
				server := spec.NewServer()
				server.Spec.URL = value
				p.openAPI.Servers = append(p.openAPI.Servers, server)
			}

			println("@host is deprecated use servers instead")
		case "@basepath":
			if len(p.openAPI.Servers) == 0 {
				server := spec.NewServer()
				p.openAPI.Servers = append(p.openAPI.Servers, server)
			}
			p.openAPI.Servers[0].Spec.URL += value

			println("@basepath is deprecated use servers instead")

		case acceptAttr:
			println("acceptAttribute is deprecated, as there is no such field on top level in spec V3.1")
		case produceAttr:
			println("produce is deprecated, as there is no such field on top level in spec V3.1")
		case "@schemes":
			println("@schemes is deprecated use servers instead")
		case "@tag.name":
			tag := &spec.Extendable[spec.Tag]{
				Spec: &spec.Tag{
					Name: value,
				},
			}

			p.openAPI.Tags = append(p.openAPI.Tags, tag)
		case "@tag.description":
			tag := p.openAPI.Tags[len(p.openAPI.Tags)-1]
			tag.Spec.Description = value
		case "@tag.description.markdown":
			tag := p.openAPI.Tags[len(p.openAPI.Tags)-1]

			commentInfo, err := getMarkdownForTag(tag.Spec.Name, p.markdownFileDir)
			if err != nil {
				return err
			}

			tag.Spec.Description = string(commentInfo)
		case "@tag.docs.url":
			tag := p.openAPI.Tags[len(p.openAPI.Tags)-1]
			tag.Spec.ExternalDocs = spec.NewExternalDocs()
			tag.Spec.ExternalDocs.Spec.URL = value
		case "@tag.docs.description":
			tag := p.openAPI.Tags[len(p.openAPI.Tags)-1]
			if tag.Spec.ExternalDocs == nil {
				return fmt.Errorf("%s needs to come after a @tags.docs.url", attribute)
			}

			tag.Spec.ExternalDocs.Spec.Description = value
		case secBasicAttr, secAPIKeyAttr, secApplicationAttr, secImplicitAttr, secPasswordAttr, secAccessCodeAttr, secBearerAuthAttr:
			key, scheme, err := parseSecAttributesV3(attribute, comments, &line)
			if err != nil {
				return err
			}

			schemeSpec := spec.NewSecuritySchemeSpec()
			schemeSpec.Spec.Spec = scheme

			if p.openAPI.Components.Spec.SecuritySchemes == nil {
				p.openAPI.Components.Spec.SecuritySchemes = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.SecurityScheme]])
			}

			p.openAPI.Components.Spec.SecuritySchemes[key] = schemeSpec

		case securityAttr:
			p.openAPI.Security = append(p.openAPI.Security, parseSecurity(value))

		case "@query.collection.format":
			p.collectionFormatInQuery = TransToValidCollectionFormat(value)

		case extDocsDescAttr, extDocsURLAttr:
			if p.openAPI.ExternalDocs == nil {
				p.openAPI.ExternalDocs = spec.NewExternalDocs()
			}

			switch attr {
			case extDocsDescAttr:
				p.openAPI.ExternalDocs.Spec.Description = value
			case extDocsURLAttr:
				p.openAPI.ExternalDocs.Spec.URL = value
			}

		case "@x-taggroups":
			originalAttribute := strings.Split(commentLine, " ")[0]
			if len(value) == 0 {
				return fmt.Errorf("annotation %s need a value", attribute)
			}

			var valueJSON interface{}
			if err := json.Unmarshal([]byte(value), &valueJSON); err != nil {
				return fmt.Errorf("annotation %s need a valid json value. error: %s", originalAttribute, err.Error())
			}

			p.openAPI.Info.Extensions[originalAttribute[1:]] = valueJSON
		case "@servers.url":
			server := spec.NewServer()
			server.Spec.URL = value
			matches := serversURLPattern.FindAllStringSubmatch(value, -1)
			server.Spec.Variables = make(map[string]*spec.Extendable[spec.ServerVariable])
			for _, match := range matches {
				server.Spec.Variables[match[1]] = spec.NewServerVariable()
			}

			p.openAPI.Servers = append(p.openAPI.Servers, server)
		case "@servers.description":
			server := p.openAPI.Servers[len(p.openAPI.Servers)-1]
			server.Spec.Description = value
		case "@servers.variables.enum":
			server := p.openAPI.Servers[len(p.openAPI.Servers)-1]
			matches := serversVariablesPattern.FindStringSubmatch(value)
			if len(matches) > 0 {
				variable, ok := server.Spec.Variables[matches[1]]
				if !ok {
					p.debug.Printf("Variables are not detected.")
					continue
				}
				variable.Spec.Enum = append(variable.Spec.Enum, matches[2])
			}
		case "@servers.variables.default":
			server := p.openAPI.Servers[len(p.openAPI.Servers)-1]
			matches := serversVariablesPattern.FindStringSubmatch(value)
			if len(matches) > 0 {
				variable, ok := server.Spec.Variables[matches[1]]
				if !ok {
					p.debug.Printf("Variables are not detected.")
					continue
				}
				variable.Spec.Default = matches[2]
			}
		case "@servers.variables.description":
			server := p.openAPI.Servers[len(p.openAPI.Servers)-1]
			matches := serversVariablesPattern.FindStringSubmatch(value)
			if len(matches) > 0 {
				variable, ok := server.Spec.Variables[matches[1]]
				if !ok {
					p.debug.Printf("Variables are not detected.")
					continue
				}
				variable.Spec.Default = matches[2]
			}
		case "@servers.variables.description.markdown":
			server := p.openAPI.Servers[len(p.openAPI.Servers)-1]
			matches := serversVariablesPattern.FindStringSubmatch(value)
			if len(matches) > 0 {
				variable, ok := server.Spec.Variables[matches[1]]
				if !ok {
					p.debug.Printf("Variables are not detected.")
					continue
				}
				commentInfo, err := getMarkdownForTag(matches[1], p.markdownFileDir)
				if err != nil {
					return err
				}
				variable.Spec.Description = string(commentInfo)
			}
		default:
			if strings.HasPrefix(attribute, "@x-") {
				err := p.parseExtensionsV3(value, attribute)
				if err != nil {
					return fmt.Errorf("could not parse extension comment: %w", err)
				}
			}
		}

		previousAttribute = attribute
	}

	return nil
}

func (p *Parser) parseExtensionsV3(value, attribute string) error {
	extensionName := attribute[1:]

	// // for each security definition
	// for _, v := range p.openAPI.Components.Spec.SecuritySchemes{
	// 	// check if extension exists
	// 	_, extExistsInSecurityDef := v.VendorExtensible.Extensions.GetString(extensionName)
	// 	// if it exists in at least one, then we stop iterating
	// 	if extExistsInSecurityDef {
	// 		return nil
	// 	}
	// }

	if len(value) == 0 {
		return fmt.Errorf("annotation %s need a value", attribute)
	}

	if p.openAPI.Info.Extensions == nil {
		p.openAPI.Info.Extensions = map[string]any{}
	}

	var valueJSON interface{}
	err := json.Unmarshal([]byte(value), &valueJSON)
	if err != nil {
		return fmt.Errorf("annotation %s need a valid json value. error: %s", attribute, err.Error())
	}

	if strings.Contains(extensionName, "logo") {
		p.openAPI.Info.Extensions[extensionName] = valueJSON
		return nil
	}

	p.openAPI.Info.Extensions[attribute[1:]] = valueJSON

	return nil
}

func setspecInfo(openAPI *spec.OpenAPI, attribute, value string) {
	switch attribute {
	case versionAttr:
		openAPI.Info.Spec.Version = value
	case titleAttr:
		openAPI.Info.Spec.Title = value
	case tosAttr:
		openAPI.Info.Spec.TermsOfService = value
	case descriptionAttr:
		openAPI.Info.Spec.Description = value
	case conNameAttr:
		if openAPI.Info.Spec.Contact == nil {
			openAPI.Info.Spec.Contact = spec.NewContact()
		}

		openAPI.Info.Spec.Contact.Spec.Name = value
	case conEmailAttr:
		if openAPI.Info.Spec.Contact == nil {
			openAPI.Info.Spec.Contact = spec.NewContact()
		}

		openAPI.Info.Spec.Contact.Spec.Email = value
	case conURLAttr:
		if openAPI.Info.Spec.Contact == nil {
			openAPI.Info.Spec.Contact = spec.NewContact()
		}

		openAPI.Info.Spec.Contact.Spec.URL = value
	case licNameAttr:
		if openAPI.Info.Spec.License == nil {
			openAPI.Info.Spec.License = spec.NewLicense()
		}
		openAPI.Info.Spec.License.Spec.Name = value
	case licURLAttr:
		if openAPI.Info.Spec.License == nil {
			openAPI.Info.Spec.License = spec.NewLicense()
		}
		openAPI.Info.Spec.License.Spec.URL = value
	}
}

func parseSecAttributesV3(context string, lines []string, index *int) (string, *spec.SecurityScheme, error) {
	const (
		in               = "@in"
		name             = "@name"
		descriptionAttr  = "@description"
		tokenURL         = "@tokenurl"
		authorizationURL = "@authorizationurl"
	)

	var search []string

	attribute := strings.ToLower(FieldsByAnySpace(lines[*index], 2)[0])
	key := getSecurityDefinitionKey(lines)
	switch attribute {
	case secBasicAttr:
		scheme := spec.SecurityScheme{
			Type:   "http",
			Scheme: "basic",
		}
		return key, &scheme, nil
	case secAPIKeyAttr:
		search = []string{in, name}
	case secApplicationAttr, secPasswordAttr:
		search = []string{tokenURL, in, name}
	case secImplicitAttr:
		search = []string{authorizationURL, in}
	case secAccessCodeAttr:
		search = []string{tokenURL, authorizationURL, in}
	case secBearerAuthAttr:
		// Support Bearer scheme with parameters
		scheme := spec.SecurityScheme{
			Type:   "http",
			Scheme: "bearer",
		}
		// Parse parameters
		*index++
		description := ""
		for ; *index < len(lines); *index++ {
			v := strings.TrimSpace(lines[*index])
			if len(v) == 0 {
				continue
			}
			fields := FieldsByAnySpace(v, 2)
			securityAttr := strings.ToLower(fields[0])
			var value string
			if len(fields) > 1 {
				value = fields[1]
			}
			if securityAttr == "@description" {
				description = value
			}
			if securityAttr == "@bearerformat" {
				scheme.BearerFormat = value
			}
			if strings.HasPrefix(securityAttr, "@securitydefinitions.") {
				*index--
				break
			}
		}
		scheme.Description = description
		return key, &scheme, nil
	}

	// For the first line we get the attributes in the context parameter, so we skip to the next one
	*index++

	attrMap, scopes := make(map[string]string), make(map[string]string)
	extensions, description := make(map[string]interface{}), ""

	for ; *index < len(lines); *index++ {
		v := strings.TrimSpace(lines[*index])
		if len(v) == 0 {
			continue
		}

		fields := FieldsByAnySpace(v, 2)
		securityAttr := strings.ToLower(fields[0])
		var value string
		if len(fields) > 1 {
			value = fields[1]
		}

		for _, findTerm := range search {
			if securityAttr == findTerm {
				attrMap[securityAttr] = value

				break
			}
		}

		isExists, err := isExistsScope(securityAttr)
		if err != nil {
			return "", nil, err
		}

		if isExists {
			scopes[securityAttr[len(scopeAttrPrefix):]] = v[len(securityAttr):]
		}

		if strings.HasPrefix(securityAttr, "@x-") {
			// Add the custom attribute without the @
			extensions[securityAttr[1:]] = value
		}

		// Not mandatory field
		if securityAttr == descriptionAttr {
			description = value
		}

		// next securityDefinitions
		if strings.Index(securityAttr, "@securitydefinitions.") == 0 {
			// Go back to the previous line and break
			*index--

			break
		}
	}

	if len(attrMap) != len(search) {
		return "", nil, fmt.Errorf("%s is %v required", context, search)
	}

	scheme := &spec.SecurityScheme{}
	key = getSecurityDefinitionKey(lines)

	switch attribute {
	case secAPIKeyAttr:
		scheme.Type = "apiKey"
		scheme.In = attrMap[in]
		scheme.Name = attrMap[name]
	case secApplicationAttr:
		scheme.Type = "oauth2"
		scheme.In = attrMap[in]
		scheme.Flows = spec.NewOAuthFlows()
		scheme.Flows.Spec.ClientCredentials = spec.NewOAuthFlow()
		scheme.Flows.Spec.ClientCredentials.Spec.TokenURL = attrMap[tokenURL]

		scheme.Flows.Spec.ClientCredentials.Spec.Scopes = make(map[string]string)
		for k, v := range scopes {
			scheme.Flows.Spec.ClientCredentials.Spec.Scopes[k] = v
		}
	case secImplicitAttr:
		scheme.Type = "oauth2"
		scheme.In = attrMap[in]
		scheme.Flows = spec.NewOAuthFlows()
		scheme.Flows.Spec.Implicit = spec.NewOAuthFlow()
		scheme.Flows.Spec.Implicit.Spec.AuthorizationURL = attrMap[authorizationURL]
		scheme.Flows.Spec.Implicit.Spec.Scopes = make(map[string]string)
		for k, v := range scopes {
			scheme.Flows.Spec.Implicit.Spec.Scopes[k] = v
		}
	case secPasswordAttr:
		scheme.Type = "oauth2"
		scheme.In = attrMap[in]
		scheme.Flows = spec.NewOAuthFlows()
		scheme.Flows.Spec.Password = spec.NewOAuthFlow()
		scheme.Flows.Spec.Password.Spec.TokenURL = attrMap[tokenURL]

		scheme.Flows.Spec.Password.Spec.Scopes = make(map[string]string)
		for k, v := range scopes {
			scheme.Flows.Spec.Password.Spec.Scopes[k] = v
		}

	case secAccessCodeAttr:
		scheme.Type = "oauth2"
		scheme.In = attrMap[in]
		scheme.Flows = spec.NewOAuthFlows()
		scheme.Flows.Spec.AuthorizationCode = spec.NewOAuthFlow()
		scheme.Flows.Spec.AuthorizationCode.Spec.AuthorizationURL = attrMap[authorizationURL]
		scheme.Flows.Spec.AuthorizationCode.Spec.TokenURL = attrMap[tokenURL]
	}

	scheme.Description = description

	if scheme.Flows != nil && scheme.Flows.Extensions == nil && len(extensions) > 0 {
		scheme.Flows.Extensions = make(map[string]interface{})
	}

	for k, v := range extensions {
		scheme.Flows.Extensions[k] = v
	}

	return key, scheme, nil
}

func getSecurityDefinitionKey(lines []string) string {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(strings.ToLower(line)), "@securitydefinitions") {
			splittedLine := strings.Fields(line)
			return splittedLine[len(splittedLine)-1]
		}
	}

	return ""
}

// ParseRouterAPIInfoV3 parses router api info for given astFile.
func (p *Parser) ParseRouterAPIInfoV3(fileInfo *AstFileInfo) error {
	for _, astDescription := range fileInfo.File.Decls {
		if (fileInfo.ParseFlag & ParseOperations) == ParseNone {
			continue
		}

		astDeclaration, ok := astDescription.(*ast.FuncDecl)
		if !ok || astDeclaration.Doc == nil || astDeclaration.Doc.List == nil {
			continue
		}

		if p.matchTags(astDeclaration.Doc.List) &&
			matchExtension(p.parseExtension, astDeclaration.Doc.List) {
			// for per 'function' comment, create a new 'Operation' object
			operation := NewOperationV3(p, SetCodeExampleFilesDirectoryV3(p.codeExampleFilesDir))

			for _, comment := range astDeclaration.Doc.List {
				err := operation.ParseComment(comment.Text, fileInfo.File)
				if err != nil {
					return fmt.Errorf("ParseComment error in file %s :%+v", fileInfo.Path, err)
				}
			}

			// workaround until we replace the produce comment with a new @Success syntax
			// We first need to setup all responses before we can set the mimetypes
			err := operation.ProcessProduceComment()
			if err != nil {
				return err
			}

			err = processRouterOperationV3(p, operation)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processRouterOperationV3(p *Parser, o *OperationV3) error {
	for _, routeProperties := range o.RouterProperties {
		var (
			pathItem *spec.RefOrSpec[spec.Extendable[spec.PathItem]]
			ok       bool
		)

		pathItem, ok = p.openAPI.Paths.Spec.Paths[routeProperties.Path]
		if !ok {
			pathItem = &spec.RefOrSpec[spec.Extendable[spec.PathItem]]{
				Spec: &spec.Extendable[spec.PathItem]{
					Spec: &spec.PathItem{},
				},
			}
		}

		op := refRouteMethodOpV3(pathItem.Spec.Spec, routeProperties.HTTPMethod)

		// check if we already have an operation for this path and method
		if *op != nil {
			err := fmt.Errorf("route %s %s is declared multiple times", routeProperties.HTTPMethod, routeProperties.Path)
			if p.Strict {
				return err
			}

			p.debug.Printf("warning: %s\n", err)
		}

		*op = &o.Operation

		p.openAPI.Paths.Spec.Paths[routeProperties.Path] = pathItem
	}

	return nil
}

func refRouteMethodOpV3(item *spec.PathItem, method string) **spec.Operation {
	switch method {
	case http.MethodGet:
		if item.Get == nil {
			item.Get = &spec.Extendable[spec.Operation]{}
		}
		return &item.Get.Spec
	case http.MethodPost:
		if item.Post == nil {
			item.Post = &spec.Extendable[spec.Operation]{}
		}
		return &item.Post.Spec
	case http.MethodDelete:
		if item.Delete == nil {
			item.Delete = &spec.Extendable[spec.Operation]{}
		}
		return &item.Delete.Spec
	case http.MethodPut:
		if item.Put == nil {
			item.Put = &spec.Extendable[spec.Operation]{}
		}
		return &item.Put.Spec
	case http.MethodPatch:
		if item.Patch == nil {
			item.Patch = &spec.Extendable[spec.Operation]{}
		}
		return &item.Patch.Spec
	case http.MethodHead:
		if item.Head == nil {
			item.Head = &spec.Extendable[spec.Operation]{}
		}
		return &item.Head.Spec
	case http.MethodOptions:
		if item.Options == nil {
			item.Options = &spec.Extendable[spec.Operation]{}
		}
		return &item.Options.Spec
	default:
		return nil
	}
}

func (p *Parser) getTypeSchemaV3(typeName string, file *ast.File, ref bool) (*spec.RefOrSpec[spec.Schema], error) {
	if override, ok := p.Overrides[typeName]; ok {
		p.debug.Printf("Override detected for %s: using %s instead", typeName, override)
		schema, err := parseObjectSchemaV3(p, override, file)
		if err != nil {
			return nil, err
		}

		return schema, nil

	}

	if IsInterfaceLike(typeName) {
		return spec.NewSchemaSpec(), nil
	}

	if IsGolangPrimitiveType(typeName) {
		return PrimitiveSchemaV3(TransToValidSchemeType(typeName)), nil
	}

	schemaType, err := convertFromSpecificToPrimitive(typeName)
	if err == nil {
		return PrimitiveSchemaV3(schemaType), nil
	}

	typeSpecDef := p.packages.FindTypeSpec(typeName, file)
	if typeSpecDef == nil {
		p.packages.FindTypeSpec(typeName, file) // uncomment for debugging
		return nil, fmt.Errorf("cannot find type definition: %s", typeName)
	}

	if override, ok := p.Overrides[typeSpecDef.FullPath()]; ok {
		if override == "" {
			p.debug.Printf("Override detected for %s: ignoring", typeSpecDef.FullPath())

			return nil, ErrSkippedField
		}

		p.debug.Printf("Override detected for %s: using %s instead", typeSpecDef.FullPath(), override)

		separator := strings.LastIndex(override, ".")
		if separator == -1 {
			// treat as a swaggertype tag
			parts := strings.Split(override, ",")
			return BuildCustomSchemaV3(parts)
		}

		typeSpecDef = p.packages.findTypeSpec(override[0:separator], override[separator+1:])
	}

	schema, ok := p.parsedSchemasV3[typeSpecDef]
	if !ok {
		var err error

		schema, err = p.ParseDefinitionV3(typeSpecDef)
		if err != nil {
			if err == ErrRecursiveParseStruct && ref {
				return p.getRefTypeSchemaV3(typeSpecDef, schema), nil
			}
			return nil, err
		}
	}

	if ref {
		if IsComplexSchemaV3(schema) {
			return p.getRefTypeSchemaV3(typeSpecDef, schema), nil
		}

		// if it is a simple schema, just return a copy
		newSchema := *schema.Schema
		return spec.NewRefOrSpec(nil, &newSchema), nil
	}

	return spec.NewRefOrSpec(nil, schema.Schema), nil
}

// ParseDefinitionV3 parses given type spec that corresponds to the type under
// given name and package, and populates swagger schema definitions registry
// with a schema for the given type
func (p *Parser) ParseDefinitionV3(typeSpecDef *TypeSpecDef) (*SchemaV3, error) {
	typeName := typeSpecDef.TypeName()
	schema, found := p.parsedSchemasV3[typeSpecDef]
	if found {
		p.debug.Printf("Skipping '%s', already parsed.", typeName)

		return schema, nil
	}

	if p.isInStructStack(typeSpecDef) {
		p.debug.Printf("Skipping '%s', recursion detected.", typeName)

		schemaName := typeName
		if typeSpecDef.SchemaName != "" {
			schemaName = typeSpecDef.SchemaName
		}

		schema := &SchemaV3{
			Name:    schemaName,
			PkgPath: typeSpecDef.PkgPath,
			Schema:  PrimitiveSchemaV3(OBJECT).Spec,
		}

		p.parsedSchemasV3[typeSpecDef] = schema

		if p.openAPI.Components.Spec.Schemas == nil {
			p.openAPI.Components.Spec.Schemas = make(map[string]*spec.RefOrSpec[spec.Schema])
		}
		p.openAPI.Components.Spec.Schemas[schema.Name] = spec.NewRefOrSpec(nil, schema.Schema)

		return schema, ErrRecursiveParseStruct
	}

	p.structStack = append(p.structStack, typeSpecDef)

	p.debug.Printf("Generating %s", typeName)

	definition, err := p.parseTypeExprV3(typeSpecDef.File, typeSpecDef.TypeSpec.Type, false)
	if err != nil {
		p.debug.Printf("Error parsing type definition '%s': %s", typeName, err)
		return nil, err
	}

	if definition.Spec.Description == "" {
		fillDefinitionDescriptionV3(p, definition.Spec, typeSpecDef.File, typeSpecDef)
	}

	if len(typeSpecDef.Enums) > 0 {
		var varNames []string
		var enumComments = make(map[string]string)
		for _, value := range typeSpecDef.Enums {
			definition.Spec.Enum = append(definition.Spec.Enum, value.Value)
			varNames = append(varNames, value.key)
			if len(value.Comment) > 0 {
				enumComments[value.key] = value.Comment
			}
		}

		if definition.Spec.Extensions == nil {
			definition.Spec.Extensions = make(map[string]any)
		}

		definition.Spec.Extensions[enumVarNamesExtension] = varNames
		if len(enumComments) > 0 {
			definition.Spec.Extensions[enumCommentsExtension] = enumComments
		}
	}
	schemaName := typeName
	if typeSpecDef.SchemaName != "" {
		schemaName = typeSpecDef.SchemaName
	}

	sch := SchemaV3{
		Name:    schemaName,
		PkgPath: typeSpecDef.PkgPath,
		Schema:  definition.Spec,
	}
	p.parsedSchemasV3[typeSpecDef] = &sch

	// update an empty schema as a result of recursion
	s2, found := p.outputSchemasV3[typeSpecDef]
	if found {
		p.openAPI.Components.Spec.Schemas[s2.Name] = definition
	}

	return &sch, nil
}

// fillDefinitionDescription additionally fills fields in definition (spec.Schema)
// TODO: If .go file contains many types, it may work for a long time
func fillDefinitionDescriptionV3(parser *Parser, definition *spec.Schema, file *ast.File, typeSpecDef *TypeSpecDef) {
	for _, astDeclaration := range file.Decls {
		generalDeclaration, ok := astDeclaration.(*ast.GenDecl)
		if !ok || generalDeclaration.Tok != token.TYPE {
			continue
		}

		for _, astSpec := range generalDeclaration.Specs {
			typeSpec, ok := astSpec.(*ast.TypeSpec)
			if !ok || typeSpec != typeSpecDef.TypeSpec {
				continue
			}

			var typeName string
			if typeSpec.Name != nil {
				typeName = typeSpec.Name.Name
			}

			text, err := parser.extractDeclarationDescription(typeName, typeSpec.Comment, generalDeclaration.Doc)
			if err != nil {
				parser.debug.Printf("Error extracting declaration description: %s", err)
				continue
			}

			definition.Description = text
		}
	}
}

// parseTypeExprV3 parses given type expression that corresponds to the type under
// given name and package, and returns swagger schema for it.
func (p *Parser) parseTypeExprV3(file *ast.File, typeExpr ast.Expr, ref bool) (*spec.RefOrSpec[spec.Schema], error) {
	const errMessage = "parse type expression v3"

	switch expr := typeExpr.(type) {
	// type Foo interface{}
	case *ast.InterfaceType:
		return spec.NewSchemaSpec(), nil

	// type Foo struct {...}
	case *ast.StructType:
		return p.parseStructV3(file, expr.Fields)

	// type Foo Baz
	case *ast.Ident:
		result, err := p.getTypeSchemaV3(expr.Name, file, ref)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errMessage, err)
		}

		return result, nil
	// type Foo *Baz
	case *ast.StarExpr:
		return p.parseTypeExprV3(file, expr.X, ref)

	// type Foo pkg.Bar
	case *ast.SelectorExpr:
		if xIdent, ok := expr.X.(*ast.Ident); ok {
			result, err := p.getTypeSchemaV3(fullTypeName(xIdent.Name, expr.Sel.Name), file, ref)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", errMessage, err)
			}

			return result, nil
		}
	// type Foo []Baz
	case *ast.ArrayType:
		itemSchema, err := p.parseTypeExprV3(file, expr.Elt, true)
		if err != nil {
			return nil, err
		}

		if itemSchema == nil {
			schema := &spec.Schema{}
			schema.Type = &spec.SingleOrArray[string]{ARRAY}
			schema.Items = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
			p.debug.Printf("Creating array with empty item schema %v", expr.Elt)

			return spec.NewRefOrSpec(nil, schema), nil
		}

		result := &spec.Schema{}
		result.Type = &spec.SingleOrArray[string]{ARRAY}
		result.Items = spec.NewBoolOrSchema(false, itemSchema)

		return spec.NewRefOrSpec(nil, result), nil
	// type Foo map[string]Bar
	case *ast.MapType:
		if _, ok := expr.Value.(*ast.InterfaceType); ok {
			result := &spec.Schema{}
			result.AdditionalProperties = spec.NewBoolOrSchema(false, spec.NewSchemaSpec())
			result.Type = &spec.SingleOrArray[string]{OBJECT}

			return spec.NewRefOrSpec(nil, result), nil
		}

		schema, err := p.parseTypeExprV3(file, expr.Value, true)
		if err != nil {
			return nil, err
		}

		result := &spec.Schema{}
		result.AdditionalProperties = spec.NewBoolOrSchema(false, schema)
		result.Type = &spec.SingleOrArray[string]{OBJECT}

		return spec.NewRefOrSpec(nil, result), nil
	case *ast.FuncType:
		return nil, ErrFuncTypeField
		// ...
	}

	return p.parseGenericTypeExprV3(file, typeExpr)
}

func (p *Parser) parseStructV3(file *ast.File, fields *ast.FieldList) (*spec.RefOrSpec[spec.Schema], error) {
	required, properties := make([]string, 0), make(map[string]*spec.RefOrSpec[spec.Schema])

	for _, field := range fields.List {
		fieldProps, requiredFromAnon, err := p.parseStructFieldV3(file, field)
		if err != nil {
			if err == ErrFuncTypeField || err == ErrSkippedField {
				continue
			}

			return nil, err
		}

		if len(fieldProps) == 0 {
			continue
		}

		required = append(required, requiredFromAnon...)

		for k, v := range fieldProps {
			properties[k] = v
		}
	}

	sort.Strings(required)

	result := spec.NewSchemaSpec()
	result.Spec.Type = &spec.SingleOrArray[string]{OBJECT}
	result.Spec.Properties = properties
	result.Spec.Required = required

	return result, nil
}

func (p *Parser) parseStructFieldV3(file *ast.File, field *ast.Field) (map[string]*spec.RefOrSpec[spec.Schema], []string, error) {
	if field.Tag != nil {
		skip, ok := reflect.StructTag(strings.ReplaceAll(field.Tag.Value, "`", "")).Lookup("swaggerignore")
		if ok && strings.EqualFold(skip, "true") {
			return nil, nil, nil
		}
	}

	ps := p.fieldParserFactoryV3(p, file, field)

	if ps.ShouldSkip() {
		return nil, nil, nil
	}

	fieldName, err := ps.FieldName()
	if err != nil {
		return nil, nil, err
	}

	if fieldName == "" {
		typeName, err := getFieldType(file, field.Type, nil)
		if err != nil {
			return nil, nil, err
		}

		schema, err := p.getTypeSchemaV3(typeName, file, false)
		if err != nil {
			return nil, nil, err
		}

		if len(*schema.Spec.Type) > 0 && (*schema.Spec.Type)[0] == OBJECT {
			if len(schema.Spec.Properties) == 0 {
				return nil, nil, nil
			}

			properties := make(map[string]*spec.RefOrSpec[spec.Schema])
			for k, v := range schema.Spec.Properties {
				properties[k] = v
			}

			return properties, schema.Spec.Required, nil
		}
		// for alias type of non-struct types ,such as array,map, etc. ignore field tag.
		return map[string]*spec.RefOrSpec[spec.Schema]{
			typeName: schema,
		}, nil, nil

	}

	schema, err := ps.CustomSchema()
	if err != nil {
		return nil, nil, err
	}

	if schema == nil {
		typeName, err := getFieldType(file, field.Type, nil)
		if err == nil {
			// named type
			schema, err = p.getTypeSchemaV3(typeName, file, true)
			if err != nil {
				return nil, nil, err
			}

		} else {
			// unnamed type
			parsedSchema, err := p.parseTypeExprV3(file, field.Type, false)
			if err != nil {
				return nil, nil, err
			}

			schema = parsedSchema
		}
	}

	err = ps.ComplementSchema(schema)
	if err != nil {
		return nil, nil, err
	}

	var tagRequired []string

	required, err := ps.IsRequired()
	if err != nil {
		return nil, nil, err
	}

	if required {
		tagRequired = append(tagRequired, fieldName)
	}

	if formName := ps.FormName(); len(formName) > 0 {
		if schema.Spec.Extensions == nil {
			schema.Spec.Extensions = make(map[string]any)
		}
		schema.Spec.Extensions[formTag] = formName
	}

	return map[string]*spec.RefOrSpec[spec.Schema]{fieldName: schema}, tagRequired, nil
}

func (p *Parser) getRefTypeSchemaV3(typeSpecDef *TypeSpecDef, schema *SchemaV3) *spec.RefOrSpec[spec.Schema] {
	_, ok := p.outputSchemasV3[typeSpecDef]
	if !ok {
		if p.openAPI.Components.Spec.Schemas == nil {
			p.openAPI.Components.Spec.Schemas = make(map[string]*spec.RefOrSpec[spec.Schema])
		}

		p.openAPI.Components.Spec.Schemas[schema.Name] = spec.NewSchemaSpec()

		if schema.Schema != nil {
			p.openAPI.Components.Spec.Schemas[schema.Name] = spec.NewRefOrSpec(nil, schema.Schema)
		}

		p.outputSchemasV3[typeSpecDef] = schema
	}

	refSchema := RefSchemaV3(schema.Name)

	return refSchema
}

// GetSchemaTypePathV3 get path of schema type.
func (p *Parser) GetSchemaTypePathV3(schema *spec.RefOrSpec[spec.Schema], depth int) []string {
	if schema == nil || depth == 0 {
		return nil
	}

	name := ""
	if schema.Ref != nil {
		name = schema.Ref.Ref
	}

	if name != "" {
		if pos := strings.LastIndexByte(name, '/'); pos >= 0 {
			name = name[pos+1:]
			if schema, ok := p.openAPI.Components.Spec.Schemas[name]; ok {
				return p.GetSchemaTypePathV3(schema, depth)
			}
		}

		return nil
	}

	if schema.Spec.Type != nil && len(*schema.Spec.Type) > 0 {
		switch (*schema.Spec.Type)[0] {
		case ARRAY:
			if schema.Spec.Items != nil && schema.Spec.Items.Schema != nil {
				depth--

				s := []string{(*schema.Spec.Type)[0]}

				return append(s, p.GetSchemaTypePathV3(schema.Spec.Items.Schema, depth)...)
			}
		case OBJECT:
			if schema.Spec.AdditionalProperties != nil && schema.Spec.AdditionalProperties.Schema != nil {
				// for map
				depth--

				s := []string{(*schema.Spec.Type)[0]}

				return append(s, p.GetSchemaTypePathV3(schema.Spec.AdditionalProperties.Schema, depth)...)
			}
		}

		return []string{(*schema.Spec.Type)[0]}
	}

	println("found schema with no Type, returning any")
	return []string{ANY}
}

func (p *Parser) getSchemaByRef(ref *spec.Ref) *spec.Schema {
	searchString := strings.ReplaceAll(ref.Ref, "#/components/schemas/", "")
	schemaRef, exists := p.openAPI.Components.Spec.Schemas[searchString]
	if !exists || schemaRef == nil {
		println(fmt.Sprintf("Schema not found for ref: %s, returning any", ref.Ref))
		return &spec.Schema{} // return empty schema if not found
	}

	return schemaRef.Spec
}
