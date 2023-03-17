package swag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
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
			scheme, err := parseSecAttributes(attribute, comments, &line)
			if err != nil {
				return err
			}

			parser.swagger.SecurityDefinitions[value] = scheme

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
				extensionName := attribute[1:]

				extExistsInSecurityDef := false
				// for each security definition
				for _, v := range parser.swagger.SecurityDefinitions {
					// check if extension exists
					_, extExistsInSecurityDef = v.VendorExtensible.Extensions.GetString(extensionName)
					// if it exists in at least one, then we stop iterating
					if extExistsInSecurityDef {
						break
					}
				}

				// if it is present on security def, don't add it again
				if extExistsInSecurityDef {
					break
				}

				if len(value) == 0 {
					return fmt.Errorf("annotation %s need a value", attribute)
				}

				var valueJSON interface{}
				err := json.Unmarshal([]byte(value), &valueJSON)
				if err != nil {
					return fmt.Errorf("annotation %s need a valid json value. error: %s", attribute, err.Error())
				}

				if strings.Contains(extensionName, "logo") {
					parser.swagger.Info.Extensions.Add(extensionName, valueJSON)
				} else {
					if parser.swagger.Extensions == nil {
						parser.swagger.Extensions = make(map[string]interface{})
					}

					parser.swagger.Extensions[attribute[1:]] = valueJSON
				}
			}
		}

		previousAttribute = attribute
	}

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

func parseSecAttributesV3(context string, lines []string, index *int) (*openapi.SecurityScheme, error) {
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
		return &scheme, nil
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

		for _, findterm := range search {
			if securityAttr == findterm {
				attrMap[securityAttr] = value

				break
			}
		}

		isExists, err := isExistsScope(securityAttr)
		if err != nil {
			return nil, err
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
		return nil, fmt.Errorf("%s is %v required", context, search)
	}

	scheme := &openapi.SecurityScheme{}

	switch attribute {
	case secAPIKeyAttr:
		scheme.Type = "apiKey"
		scheme.In = attrMap[in]
		scheme.Name = attrMap[name]
	case secApplicationAttr:
		scheme.Type = "oauth2"
		scheme.Flows = openapi.NewOAuthFlows()
		scheme.Flows.Spec.ClientCredentials = openapi.NewOAuthFlow()
		scheme.Flows.Spec.ClientCredentials.Spec.TokenURL = attrMap[tokenURL]
	case secImplicitAttr:
		scheme.Type = "oauth2"
		scheme.Flows = openapi.NewOAuthFlows()
		scheme.Flows.Spec.Implicit = openapi.NewOAuthFlow()
		scheme.Flows.Spec.Implicit.Spec.AuthorizationURL = attrMap[authorizationURL]
	case secPasswordAttr:
		scheme.Type = "oauth2"
		scheme.Flows = openapi.NewOAuthFlows()
		scheme.Flows.Spec.Password = openapi.NewOAuthFlow()
		scheme.Flows.Spec.Password.Spec.TokenURL = attrMap[tokenURL]
	case secAccessCodeAttr:
		scheme.Type = "oauth2"
		scheme.Flows = openapi.NewOAuthFlows()
		scheme.Flows.Spec.AuthorizationCode = openapi.NewOAuthFlow()
		scheme.Flows.Spec.AuthorizationCode.Spec.AuthorizationURL = attrMap[authorizationURL]
		scheme.Flows.Spec.AuthorizationCode.Spec.TokenURL = attrMap[tokenURL]
	}

	scheme.Description = description

	// TODO handle extensions
	// for extKey, extValue := range extensions {
	// 	scheme.AddExtension(extKey, extValue)
	// }

	// TODO handle scopes
	// for scope, scopeDescription := range scopes {
	// 	scheme.AddScope(scope, scopeDescription)
	// }

	return scheme, nil
}
