package swag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	openapi "github.com/sv-tools/openapi/spec"
)

func (parser *Parser) parseGeneralAPIInfoV3(comments []string) error {
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
			setOpenAPIInfo(parser.openAPI, attr, value)
		case descriptionAttr:
			if previousAttribute == attribute {
				parser.swagger.Info.Description += "\n" + value

				continue
			}

			setOpenAPIInfo(parser.openAPI, attr, value)
		case descriptionMarkdownAttr:
			commentInfo, err := getMarkdownForTag("api", parser.markdownFileDir)
			if err != nil {
				return err
			}

			setOpenAPIInfo(parser.openAPI, attr, string(commentInfo))
		case "@host":
			if len(parser.openAPI.Servers) == 0 {
				server := openapi.NewServer()
				server.Spec.URL = value
				parser.openAPI.Servers = append(parser.openAPI.Servers, server)
			}

			println("@host is deprecated use servers instead")
		case "@basepath":
			if len(parser.openAPI.Servers) == 0 {
				server := openapi.NewServer()
				parser.openAPI.Servers = append(parser.openAPI.Servers, server)
			}
			parser.openAPI.Servers[0].Spec.URL += value

			println("@basepath is deprecated use servers instead")

		case acceptAttr:
			println("acceptAttribute is deprecated, as there is no such field on top level in openAPI V3.1")
		case produceAttr:
			println("produce is deprecated, as there is no such field on top level in openAPI V3.1")
		case "@schemes":
			println("@basepath is deprecated use servers instead")
		case "@tag.name":
			tag := &openapi.Extendable[openapi.Tag]{
				Spec: &openapi.Tag{
					Name: value,
				},
			}

			parser.openAPI.Tags = append(parser.openAPI.Tags, tag)
		case "@tag.description":
			tag := parser.openAPI.Tags[len(parser.openAPI.Tags)-1]
			tag.Spec.Description = value
		case "@tag.description.markdown":
			tag := parser.openAPI.Tags[len(parser.openAPI.Tags)-1]

			commentInfo, err := getMarkdownForTag(tag.Spec.Name, parser.markdownFileDir)
			if err != nil {
				return err
			}

			tag.Spec.Description = string(commentInfo)
		case "@tag.docs.url":
			tag := parser.openAPI.Tags[len(parser.openAPI.Tags)-1]
			tag.Spec.ExternalDocs = openapi.NewExternalDocs()
			tag.Spec.ExternalDocs.Spec.URL = value
		case "@tag.docs.description":
			tag := parser.openAPI.Tags[len(parser.openAPI.Tags)-1]
			if tag.Spec.ExternalDocs == nil {
				return fmt.Errorf("%s needs to come after a @tags.docs.url", attribute)
			}

			tag.Spec.ExternalDocs.Spec.Description = value
		case secBasicAttr, secAPIKeyAttr, secApplicationAttr, secImplicitAttr, secPasswordAttr, secAccessCodeAttr:
			key, scheme, err := parseSecAttributesV3(attribute, comments, &line)
			if err != nil {
				return err
			}

			schemeSpec := openapi.NewSecuritySchemeSpec()
			schemeSpec.Spec.Spec = scheme

			if parser.openAPI.Components.Spec.SecuritySchemes == nil {
				parser.openAPI.Components.Spec.SecuritySchemes = make(map[string]*openapi.RefOrSpec[openapi.Extendable[openapi.SecurityScheme]])
			}

			parser.openAPI.Components.Spec.SecuritySchemes[key] = schemeSpec

		case "@query.collection.format":
			parser.collectionFormatInQuery = TransToValidCollectionFormat(value)

		case extDocsDescAttr, extDocsURLAttr:
			if parser.swagger.ExternalDocs == nil {
				parser.swagger.ExternalDocs = new(spec.ExternalDocumentation)
			}
			switch attr {
			case extDocsDescAttr:
				parser.swagger.ExternalDocs.Description = value
			case extDocsURLAttr:
				parser.swagger.ExternalDocs.URL = value
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

			parser.swagger.Extensions[originalAttribute[1:]] = valueJSON // don't use the method provided by spec lib, cause it will call toLower() on attribute names, which is wrongy
		default:
			if strings.HasPrefix(attribute, "@x-") {
				err := parser.parseExtensionsV3(value, attribute)
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
	// for each security definition
	for _, v := range p.swagger.SecurityDefinitions {
		// check if extension exists
		_, extExistsInSecurityDef := v.VendorExtensible.Extensions.GetString(extensionName)
		// if it exists in at least one, then we stop iterating
		if extExistsInSecurityDef {
			return nil
		}
	}

	if len(value) == 0 {
		return fmt.Errorf("annotation %s need a value", attribute)
	}

	if p.swagger.Extensions == nil {
		p.swagger.Extensions = make(map[string]interface{})
	}

	var valueJSON interface{}
	err := json.Unmarshal([]byte(value), &valueJSON)
	if err != nil {
		return fmt.Errorf("annotation %s need a valid json value. error: %s", attribute, err.Error())
	}

	if strings.Contains(extensionName, "logo") {
		p.swagger.Info.Extensions.Add(extensionName, valueJSON)
		return nil
	}

	p.swagger.Extensions[attribute[1:]] = valueJSON

	return nil
}

func setOpenAPIInfo(openAPI *openapi.OpenAPI, attribute, value string) {
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
			openAPI.Info.Spec.Contact = openapi.NewContact()
		}

		openAPI.Info.Spec.Contact.Spec.Name = value
	case conEmailAttr:
		if openAPI.Info.Spec.Contact == nil {
			openAPI.Info.Spec.Contact = openapi.NewContact()
		}

		openAPI.Info.Spec.Contact.Spec.Email = value
	case conURLAttr:
		if openAPI.Info.Spec.Contact == nil {
			openAPI.Info.Spec.Contact = openapi.NewContact()
		}

		openAPI.Info.Spec.Contact.Spec.URL = value
	case licNameAttr:
		if openAPI.Info.Spec.License == nil {
			openAPI.Info.Spec.License = openapi.NewLicense()
		}
		openAPI.Info.Spec.License.Spec.Name = value
	case licURLAttr:
		if openAPI.Info.Spec.License == nil {
			openAPI.Info.Spec.License = openapi.NewLicense()
		}
		openAPI.Info.Spec.License.Spec.URL = value
	}
}

func parseSecAttributesV3(context string, lines []string, index *int) (string, *openapi.SecurityScheme, error) {
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
		scheme := openapi.SecurityScheme{
			Type:   "http",
			Scheme: "basic",
		}
		return "basic", &scheme, nil
	case secAPIKeyAttr:
		search = []string{in, name}
	case secApplicationAttr, secPasswordAttr:
		search = []string{tokenURL}
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

	scheme := &openapi.SecurityScheme{}
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
		scheme.Flows = openapi.NewOAuthFlows()
		scheme.Flows.Spec.ClientCredentials = openapi.NewOAuthFlow()
		scheme.Flows.Spec.ClientCredentials.Spec.TokenURL = attrMap[tokenURL]
	case secImplicitAttr:
		key = "oauth2"
		scheme.Type = "oauth2"
		scheme.Flows = openapi.NewOAuthFlows()
		scheme.Flows.Spec.Implicit = openapi.NewOAuthFlow()
		scheme.Flows.Spec.Implicit.Spec.AuthorizationURL = attrMap[authorizationURL]

		scheme.Flows.Spec.Password.Spec.Scopes = make(map[string]string)
		for k, v := range scopes {
			scheme.Flows.Spec.Password.Spec.Scopes[k] = v
		}
	case secPasswordAttr:
		key = "oauth2"
		scheme.Type = "oauth2"
		scheme.Flows = openapi.NewOAuthFlows()
		scheme.Flows.Spec.Password = openapi.NewOAuthFlow()
		scheme.Flows.Spec.Password.Spec.TokenURL = attrMap[tokenURL]

		scheme.Flows.Spec.Password.Spec.Scopes = make(map[string]string)
		for k, v := range scopes {
			scheme.Flows.Spec.Password.Spec.Scopes[k] = v
		}

	case secAccessCodeAttr:
		key = "oauth2"
		scheme.Type = "oauth2"
		scheme.Flows = openapi.NewOAuthFlows()
		scheme.Flows.Spec.AuthorizationCode = openapi.NewOAuthFlow()
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
