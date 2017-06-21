package parser

import (
	"errors"
	"fmt"
	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"regexp"
	"strconv"
	"strings"
)

type Operation struct {
	HttpMethod string
	Path       string
	spec.Operation

	parser *Parser
}

//map[int]Response
func NewOperation() *Operation {
	return &Operation{
		HttpMethod: "get",
		Operation: spec.Operation{
			OperationProps: spec.OperationProps{
				Responses: &spec.Responses{
					ResponsesProps: spec.ResponsesProps{
						StatusCodeResponses: make(map[int]spec.Response),
					},
				},
			},
		},
	}
}

func (operation *Operation) ParseComment(comment string) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "//"))
	if len(commentLine) == 0 {
		return nil
	}

	attribute := strings.Fields(commentLine)[0]
	switch strings.ToLower(attribute) {
	case "@router":
		if err := operation.ParseRouterComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	//case "@resource":
	//	resource := strings.TrimSpace(commentLine[len(attribute):])
	//	if resource[0:1] == "/" {
	//		resource = resource[1:]
	//	}
	//	operation.ForceResource = resource
	case "@summary":
		operation.Summary = strings.TrimSpace(commentLine[len(attribute):])
	case "@description":
		operation.Description = strings.TrimSpace(commentLine[len(attribute):])
	case "@success", "@failure":
		if err := operation.ParseResponseComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}

		//TODO: feat:ParseParamComment
	case "@param":
		if err := operation.ParseParamComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	case "@accept", "@consume":
		if err := operation.ParseAcceptComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	case "@produce":
		if err := operation.ParseProduceComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	}

	//operation.Models = operation.getUniqueModels()

	return nil
}

// Parse params return []string of param properties
// @Param	queryText		form	      string	  true		        "The email for login"
// 			[param name]    [param type] [data type]  [is mandatory?]   [Comment]
// @Param   some_id     path    int     true        "Some ID"
func (operation *Operation) ParseParamComment(commentLine string) error {
	paramString := commentLine

	re := regexp.MustCompile(`([-\w]+)[\s]+([\w]+)[\s]+([\S.]+)[\s]+([\w]+)[\s]+"([^"]+)"`)

	if matches := re.FindStringSubmatch(paramString); len(matches) != 6 {
		return fmt.Errorf("Can not parse param comment \"%s\", skipped.", paramString)
	} else {
		//typeName, err := operation.registerType(matches[3])
		//if err != nil {
		//	return err
		//}
		name := matches[1]
		kindIn := matches[2]

		paramTye := matches[3]

		requiredText := strings.ToLower(matches[4])
		required := (requiredText == "true" || requiredText == "required")
		description := matches[5]

		var param spec.Parameter

		if kindIn == "query" || kindIn == "path" {
			param = createParameter(kindIn, description, name, paramTye, required)
		}
		if kindIn == "body" {
			param = createParameter(kindIn, description, name, paramTye, required)
			refSplit := strings.Split(paramTye, ".")
			if len(refSplit) == 2 {
				pkgName := refSplit[0]
				typeName := refSplit[1]
				if typeSpec, ok := operation.parser.TypeDefinitions[pkgName][typeName]; ok {
					operation.parser.registerTypes[paramTye] = typeSpec
				} else {
					return fmt.Errorf("Can not find ref type:\"%s\".", paramTye)
				}

				param.Schema.Ref = spec.Ref{
					Ref: jsonreference.MustCreateRef("#/definitions/" + paramTye),
				}
			}
		}

		operation.Operation.Parameters = append(operation.Operation.Parameters, param)
	}

	return nil
}
func (operation *Operation) ParseAcceptComment(commentLine string) error {
	accepts := strings.Split(commentLine, ",")
	for _, a := range accepts {
		switch a {
		case "json", "application/json":
			operation.Consumes = append(operation.Consumes, "application/json")
		case "xml", "text/xml":
			operation.Consumes = append(operation.Consumes, "text/xml")
		case "plain", "text/plain":
			operation.Consumes = append(operation.Consumes, "text/plain")
		case "html", "text/html":
			operation.Consumes = append(operation.Consumes, "text/html")
		case "mpfd", "multipart/form-data":
			operation.Consumes = append(operation.Consumes, "multipart/form-data")
		}
	}
	return nil
}

func (operation *Operation) ParseProduceComment(commentLine string) error {
	produces := strings.Split(commentLine, ",")
	for _, a := range produces {
		switch a {
		case "json", "application/json":
			operation.Produces = append(operation.Produces, "application/json")
		case "xml", "text/xml":
			operation.Produces = append(operation.Produces, "text/xml")
		case "plain", "text/plain":
			operation.Produces = append(operation.Produces, "text/plain")
		case "html", "text/html":
			operation.Produces = append(operation.Produces, "text/html")
		case "mpfd", "multipart/form-data":
			operation.Produces = append(operation.Produces, "multipart/form-data")
		}
	}
	return nil
}

func (operation *Operation) ParseRouterComment(commentLine string) error {
	re := regexp.MustCompile(`([\w\.\/\-{}]+)[^\[]+\[([^\]]+)`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 3 {
		return fmt.Errorf("Can not parse router comment \"%s\", skipped.", commentLine)
	}
	path := matches[1]
	httpMethod := matches[2]

	//operation.ParseRouterParams(path)
	operation.Path = path
	operation.HttpMethod = strings.ToUpper(httpMethod)

	return nil
}

//func (operation *Operation) ParseRouterParams(path string) {
//	re := regexp.MustCompile(`\{(\w+)\}`)
//	matchs := re.FindAllStringSubmatch(path, -1)
//
//	if len(matchs) > 0 {
//		for _, match := range matchs {
//			group := match[1]
//			operation.Operation.Parameters = append(operation.Operation.Parameters, createPathParameter(group))
//		}
//	}
//}

func createParameter(kindIn, description, paramName, paramTye string, required bool) spec.Parameter {
	paramProps := spec.ParamProps{
		Name:        paramName,
		Description: description,
		Required:    required,
		In:          kindIn,
		Schema: &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{paramTye},
			},
		},
	}
	parameter := spec.Parameter{
		ParamProps: paramProps,
	}

	return parameter
}
func createPathParameter(paramName string) spec.Parameter {
	return createParameter("path", paramName, paramName, "string", true)
}

// @Success 200 {object} model.OrderRow "Error message, if code != 200"
func (operation *Operation) ParseResponseComment(commentLine string) error {
	re := regexp.MustCompile(`([\d]+)[\s]+([\w\{\}]+)[\s]+([\w\-\.\/]+)[^"]*(.*)?`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 5 {
		fmt.Println(len(matches))
		return fmt.Errorf("Can not parse response comment \"%s\", skipped.", commentLine)
	}

	response := spec.Response{}

	code, err := strconv.Atoi(matches[1])
	if err != nil {
		return errors.New("Success http code must be int")
	}

	response.Description = strings.Trim(matches[4], "\"")

	schemaType := strings.Trim(matches[2], "{}")
	refType := matches[3]

	if operation.parser != nil { // checking refType has existing in 'TypeDefinitions'
		refSplit := strings.Split(refType, ".")
		if len(refSplit) == 2 {
			pkgName := refSplit[0]
			typeName := refSplit[1]
			if typeSpec, ok := operation.parser.TypeDefinitions[pkgName][typeName]; ok {
				operation.parser.registerTypes[refType] = typeSpec
			} else {
				return fmt.Errorf("Can not find ref type:\"%s\".", refType)
			}

		}

	}
	// so we have to know all type in app
	//TODO: we might ommited schema.type if schemaType equals 'object'
	response.Schema = &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{schemaType}}}

	if schemaType == "object" {
		response.Schema.Ref = spec.Ref{
			Ref: jsonreference.MustCreateRef("#/definitions/" + refType),
		}
	}

	if schemaType == "array" {
		response.Schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Ref: spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/" + refType)},
				},
			},
		}

	}

	operation.Responses.StatusCodeResponses[code] = response

	return nil
}

func (operation *Operation) registerType(typeName string) error {
	if IsBasicType(typeName) {
		return nil
	}
	//TODO: extract type
	return fmt.Errorf("Not supported %+v type, only supported basic type now", typeName)
}

// refer to builtin.go
var basicTypes = map[string]bool{
	"bool":       true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"float32":    true,
	"float64":    true,
	"string":     true,
	"complex64":  true,
	"complex128": true,
	"byte":       true,
	"rune":       true,
	"uintptr":    true,
	"error":      true,
	"Time":       true,
	"file":       true,
}

func IsBasicType(typeName string) bool {
	_, ok := basicTypes[typeName]
	return ok || strings.Contains(typeName, "interface")
}
