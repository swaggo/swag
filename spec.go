package swag

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"
)

// Spec holds exported Swagger Info so clients can modify it.
type Spec struct {
	Version          string
	Host             string
	BasePath         string
	Schemes          []string
	Title            string
	Description      string
	InfoInstanceName string
	SwaggerTemplate  string
}

// ReadDoc parses SwaggerTemplate into swagger document.
func (i *Spec) ReadDoc() string {
	i.Description = strings.Replace(i.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(i.SwaggerTemplate)
	if err != nil {
		return i.SwaggerTemplate
	}

	var tpl bytes.Buffer
	if err = t.Execute(&tpl, i); err != nil {
		return i.SwaggerTemplate
	}

	return tpl.String()
}

// InstanceName returns Spec instance name.
func (i *Spec) InstanceName() string {
	return i.InfoInstanceName
}
