package parse

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

		//Operation:spec.Operation:&spec.Operation{}

	}
}

func (operation *Operation) ParseComment(comment string) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "//"))
	if len(commentLine) == 0 {
		return nil
	}

	//fmt.Println(comment)
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
	case "@description":
		operation.Description = strings.TrimSpace(commentLine[len(attribute):])
	case "@success", "@failure":
		if err := operation.ParseResponseComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	//case "@param":
	//	if err := operation.ParseParamComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
	//		return err
	//	}
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
	operation.Path = matches[1]
	operation.HttpMethod = strings.ToUpper(matches[2])
	return nil
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

	//typeName, err := operation.registerType(matches[3])
	//if err != nil {
	//	return err
	//}
	schemaType := strings.Trim(matches[2], "{}")
	refType := matches[3]

	//TODO: checking refType has existing in 'TypeDefinitions'

	//TODO: if 'object' might ommited schema.type
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
