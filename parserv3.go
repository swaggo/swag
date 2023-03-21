package swag

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/sv-tools/openapi/spec"
)

// GetOpenAPI returns *spec.OpenAPI which is the root document object for the API specification.
func (parser *Parser) GetOpenAPI() *spec.OpenAPI {
	return parser.openAPI
}

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
			println("@basepath is deprecated use servers instead")
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
		case secBasicAttr, secAPIKeyAttr, secApplicationAttr, secImplicitAttr, secPasswordAttr, secAccessCodeAttr:
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
				p.openAPI.ExternalDocs.Spec.Description = value
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
		default:
			if strings.HasPrefix(attribute, "@x-") {
				err := p.parseExtensionsV3(value, attribute)
				if err != nil {
					return errors.Wrap(err, "could not parse extension comment")
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
	switch attribute {
	case secBasicAttr:
		scheme := spec.SecurityScheme{
			Type:   "http",
			Scheme: "basic",
		}
		return "basic", &scheme, nil
	case secAPIKeyAttr:
		search = []string{in, name}
	case secApplicationAttr, secPasswordAttr:
		search = []string{tokenURL, in, name}
	case secImplicitAttr:
		search = []string{authorizationURL}
	case secAccessCodeAttr:
		search = []string{tokenURL, authorizationURL}
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
	key := ""

	switch attribute {
	case secAPIKeyAttr:
		key = "apiKey"
		scheme.Type = "apiKey"
		scheme.In = attrMap[in]
		scheme.Name = attrMap[name]
	case secApplicationAttr:
		key = "oauth2"
		scheme.Type = "oauth2"
		scheme.In = attrMap[in]
		scheme.Flows = spec.NewOAuthFlows()
		scheme.Flows.Spec.ClientCredentials = spec.NewOAuthFlow()
		scheme.Flows.Spec.ClientCredentials.Spec.TokenURL = attrMap[tokenURL]
	case secImplicitAttr:
		key = "oauth2"
		scheme.Type = "oauth2"
		scheme.Flows = spec.NewOAuthFlows()
		scheme.Flows.Spec.Implicit = spec.NewOAuthFlow()
		scheme.Flows.Spec.Implicit.Spec.AuthorizationURL = attrMap[authorizationURL]

		scheme.Flows.Spec.Password.Spec.Scopes = make(map[string]string)
		for k, v := range scopes {
			scheme.Flows.Spec.Password.Spec.Scopes[k] = v
		}
	case secPasswordAttr:
		key = "oauth2"
		scheme.Type = "oauth2"
		scheme.Flows = spec.NewOAuthFlows()
		scheme.Flows.Spec.Password = spec.NewOAuthFlow()
		scheme.Flows.Spec.Password.Spec.TokenURL = attrMap[tokenURL]

		scheme.Flows.Spec.Password.Spec.Scopes = make(map[string]string)
		for k, v := range scopes {
			scheme.Flows.Spec.Password.Spec.Scopes[k] = v
		}

	case secAccessCodeAttr:
		key = "oauth2"
		scheme.Type = "oauth2"
		scheme.Flows = spec.NewOAuthFlows()
		scheme.Flows.Spec.AuthorizationCode = spec.NewOAuthFlow()
		scheme.Flows.Spec.AuthorizationCode.Spec.AuthorizationURL = attrMap[authorizationURL]
		scheme.Flows.Spec.AuthorizationCode.Spec.TokenURL = attrMap[tokenURL]
	}

	scheme.Description = description

	if scheme.Flows.Extensions == nil && len(extensions) > 0 {
		scheme.Flows.Extensions = make(map[string]interface{})
	}

	for k, v := range extensions {
		scheme.Flows.Extensions[k] = v
	}

	return key, scheme, nil
}

// ParseRouterAPIInfo parses router api info for given astFile.
func (parser *Parser) ParseRouterAPIInfoV3(fileInfo *AstFileInfo) error {
	for _, astDescription := range fileInfo.File.Decls {
		if (fileInfo.ParseFlag & ParseOperations) == ParseNone {
			continue
		}

		astDeclaration, ok := astDescription.(*ast.FuncDecl)
		if !ok || astDeclaration.Doc == nil || astDeclaration.Doc.List == nil {
			continue
		}

		if parser.matchTags(astDeclaration.Doc.List) &&
			matchExtension(parser.parseExtension, astDeclaration.Doc.List) {
			// for per 'function' comment, create a new 'Operation' object
			operation := NewOperationV3(parser, SetCodeExampleFilesDirectoryV3(parser.codeExampleFilesDir))

			for _, comment := range astDeclaration.Doc.List {
				err := operation.ParseCommentV3(comment.Text, fileInfo.File)
				if err != nil {
					return fmt.Errorf("ParseComment error in file %s :%+v", fileInfo.Path, err)
				}
			}
			err := processRouterOperationV3(parser, operation)
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

func (p *Parser) getTypeSchemaV3(typeName string, file *ast.File, ref bool) (*spec.Schema, error) {
	if override, ok := p.Overrides[typeName]; ok {
		p.debug.Printf("Override detected for %s: using %s instead", typeName, override)
		return parseObjectSchemaV3(p, override, file)
	}

	if IsInterfaceLike(typeName) {
		return &spec.Schema{}, nil
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
			// parts := strings.Split(override, ",")
			// TODO
			// return BuildCustomSchema(parts)
		}

		typeSpecDef = p.packages.findTypeSpec(override[0:separator], override[separator+1:])
	}

	// schema, ok := p.parsedSchemas[typeSpecDef]
	// if !ok {
	// 	var err error

	// 	schema, err = p.ParseDefinition(typeSpecDef)
	// 	if err != nil {
	// 		if err == ErrRecursiveParseStruct && ref {
	// 			// TODO
	// 			// return p.getRefTypeSchema(typeSpecDef, schema), nil
	// 		}
	// 		return nil, err
	// 	}
	// }

	if ref {
		// if IsComplexSchema(schema.Schema) {
		// 	// TODO
		// 	// return p.getRefTypeSchema(typeSpecDef, schema), nil
		// }
		// // if it is a simple schema, just return a copy
		// newSchema := *schema.Schema
		// return &newSchema, nil
	}

	// return schema.Schema, nil
	return nil, nil
}
