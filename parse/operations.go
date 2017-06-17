package parse

import (
	"errors"
	"fmt"
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

type Response struct {
	Code int
	spec.Response
}

func (operation *Operation) ParseComment(comment string) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "//"))
	if len(commentLine) == 0 {
		return nil
	}

	//fmt.Println(comment)
	attribute := strings.Fields(commentLine)[0]
	switch strings.ToLower(attribute) {
	//case "@router":
	//	if err := operation.ParseRouterComment(commentLine); err != nil {
	//		return err
	//	}
	//case "@resource":
	//	resource := strings.TrimSpace(commentLine[len(attribute):])
	//	if resource[0:1] == "/" {
	//		resource = resource[1:]
	//	}
	//	operation.ForceResource = resource
	//case "@title":
	//	operation.Nickname = strings.TrimSpace(commentLine[len(attribute):])
	//case "@description":
	//	operation.Summary = strings.TrimSpace(commentLine[len(attribute):])
	//case "@success", "@failure":
	//	if err := operation.ParseResponseComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
	//		return err
	//	}
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

// @Router /customer/get-wishlist/{wishlist_id} [get]
func (operation *Operation) ParseRouterComment(commentLine string) error {
	sourceString := strings.TrimSpace(commentLine[len("@Router"):])

	re := regexp.MustCompile(`([\w\.\/\-{}]+)[^\[]+\[([^\]]+)`)
	var matches []string

	if matches = re.FindStringSubmatch(sourceString); len(matches) != 3 {
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
		return fmt.Errorf("Can not parse response comment \"%s\", skipped.", commentLine)
	}

	response := Response{}
	if code, err := strconv.Atoi(matches[1]); err != nil {
		return errors.New("Success http code must be int")
	} else {
		response.Code = code
	}
	response.Description = strings.Trim(matches[4], "\"")

	//typeName, err := operation.registerType(matches[3])
	//if err != nil {
	//	return err
	//}
	response.Schema = &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"string"}}}
	//response.Schema.Type = strings.Trim(matches[2], "{}")

	//response.Schema.Ref = "/test/test"
	//if response.Code == 200 {
	//	if matches[2] == "{array}" {
	//		operation.SetItemsType(typeName)
	//		operation.Type = "array"
	//	} else {
	//		operation.Type = typeName
	//	}
	//}
	operation.Responses.StatusCodeResponses[response.Code] = response.Response

	return nil
}

func (operation *Operation) registerType(typeName string) (string, error) {
	//registerType := ""
	//
	//if translation, ok := typeDefTranslations[typeName]; ok {
	//	registerType = translation
	//} else if IsBasicType(typeName) {
	//	registerType = typeName
	//} else {
	//	model := NewModel(operation.parser)
	//	knownModelNames := map[string]bool{}
	//
	//	err, innerModels := model.ParseModel(typeName, operation.parser.CurrentPackage, knownModelNames)
	//	if err != nil {
	//		return registerType, err
	//	}
	//	if translation, ok := typeDefTranslations[typeName]; ok {
	//		registerType = translation
	//	} else {
	//		registerType = model.Id
	//
	//		operation.Models = append(operation.Models, model)
	//		operation.Models = append(operation.Models, innerModels...)
	//	}
	//}

	return "", nil
}
