package gen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/go-openapi/spec"

	"github.com/ghodss/yaml"
	"github.com/swaggo/swag"
)

// Gen presents a generate tool for swag.
type Gen struct{}

// New creates a new Gen.
func New() *Gen {
	return &Gen{}
}

// Config presents Gen configurations.
type Config struct {
	// SearchDir the swag would be parse
	SearchDir string

	// OutputDir represents the output directory for al the generated files
	OutputDir string

	// MainAPIFile the Go file path in which 'swagger general API Info' is written
	MainAPIFile string

	// PropNamingStrategy represents property naming strategy like snakecase,camelcase,pascalcase
	PropNamingStrategy string

	// ParseVendor whether swag should be parse vendor folder
	ParseVendor bool

	// ParseDependencies whether swag should be parse outside dependency folder
	ParseDependency bool

	// MarkdownFilesDir used to find markdownfiles, which can be used for tag descriptions
	MarkdownFilesDir string
}

// Build builds swagger json file  for given searchDir and mainAPIFile. Returns json
func (g *Gen) Build(config *Config) error {
	if _, err := os.Stat(config.SearchDir); os.IsNotExist(err) {
		return fmt.Errorf("dir: %s is not exist", config.SearchDir)
	}

	log.Println("Generate swagger docs....")
	p := swag.New(swag.SetMarkdownFileDirectory(config.MarkdownFilesDir))
	p.PropNamingStrategy = config.PropNamingStrategy
	p.ParseVendor = config.ParseVendor
	p.ParseDependency = config.ParseDependency

	if err := p.ParseAPI(config.SearchDir, config.MainAPIFile); err != nil {
		return err
	}
	swagger := p.GetSwagger()

	b, err := g.jsonIndent(swagger)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(config.OutputDir, os.ModePerm); err != nil {
		return err
	}

	docs, err := os.Create(path.Join(config.OutputDir, "docs.go"))
	if err != nil {
		return err
	}
	defer docs.Close()

	swaggerJSON, err := os.Create(path.Join(config.OutputDir, "swagger.json"))
	if err != nil {
		return err
	}
	defer swaggerJSON.Close()

	if _, err := swaggerJSON.Write(b); err != nil {
		return err
	}

	swaggerYAML, err := os.Create(path.Join(config.OutputDir, "swagger.yaml"))
	if err != nil {
		return err
	}

	defer swaggerYAML.Close()
	y, err := yaml.JSONToYAML(b)
	if err != nil {
		return fmt.Errorf("cannot covert json to yaml error: %s", err)
	}

	if _, err := swaggerYAML.Write(y); err != nil {
		return err
	}

	// Write doc
	err = g.writeGoDoc(docs, swagger)
	if err != nil {
		return err
	}

	log.Printf("create docs.go at  %+v", docs.Name())
	log.Printf("create swagger.json at  %+v", swaggerJSON.Name())
	log.Printf("create swagger.yaml at  %+v", swaggerYAML.Name())

	return nil
}

func (g *Gen) jsonIndent(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "    ")
}

func (g *Gen) formatSource(src []byte) []byte {
	code, err := format.Source(src)
	if err != nil {
		code = src // Output the unformated code anyway
	}
	return code
}

func (g *Gen) writeGoDoc(output io.Writer, swagger *spec.Swagger) error {

	generator, err := template.New("swagger_info").Funcs(template.FuncMap{
		"printDoc": func(v string) string {
			// Add schemes
			v = "{\n    \"schemes\": {{ marshal .Schemes }}," + v[1:]
			// Sanitize backticks
			return strings.Replace(v, "`", "`+\"`\"+`", -1)
		},
	}).Parse(packageTemplate)
	if err != nil {
		return err
	}

	swaggerSpec := &spec.Swagger{
		VendorExtensible: swagger.VendorExtensible,
		SwaggerProps: spec.SwaggerProps{
			ID:       swagger.ID,
			Consumes: swagger.Consumes,
			Produces: swagger.Produces,
			Swagger:  swagger.Swagger,
			Info: &spec.Info{
				VendorExtensible: swagger.Info.VendorExtensible,
				InfoProps: spec.InfoProps{
					Description:    "{{.Description}}",
					Title:          "{{.Title}}",
					TermsOfService: swagger.Info.TermsOfService,
					Contact:        swagger.Info.Contact,
					License:        swagger.Info.License,
					Version:        "{{.Version}}",
				},
			},
			Host:                "{{.Host}}",
			BasePath:            "{{.BasePath}}",
			Paths:               swagger.Paths,
			Definitions:         swagger.Definitions,
			Parameters:          swagger.Parameters,
			Responses:           swagger.Responses,
			SecurityDefinitions: swagger.SecurityDefinitions,
			Security:            swagger.Security,
			Tags:                swagger.Tags,
			ExternalDocs:        swagger.ExternalDocs,
		},
	}

	// crafted docs.json
	buf, err := g.jsonIndent(swaggerSpec)
	if err != nil {
		return err
	}

	buffer := &bytes.Buffer{}
	err = generator.Execute(buffer, struct {
		Timestamp   time.Time
		Doc         string
		Host        string
		BasePath    string
		Schemes     []string
		Title       string
		Description string
		Version     string
	}{
		Timestamp:   time.Now(),
		Doc:         string(buf),
		Host:        swagger.Host,
		BasePath:    swagger.BasePath,
		Schemes:     swagger.Schemes,
		Title:       swagger.Info.Title,
		Description: swagger.Info.Description,
		Version:     swagger.Info.Version,
	})
	if err != nil {
		return err
	}

	code := g.formatSource(buffer.Bytes())

	// write
	_, err = output.Write(code)
	return err

}

var packageTemplate = `// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// {{ .Timestamp }}

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = ` + "`{{ printDoc .Doc}}`" + `

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{ 
	Version:     {{ printf "%q" .Version}},
 	Host:        {{ printf "%q" .Host}},
	BasePath:    {{ printf "%q" .BasePath}},
	Schemes:     []string{ {{ range $index, $schema := .Schemes}}{{if gt $index 0}},{{end}}{{printf "%q" $schema}}{{end}} },
	Title:       {{ printf "%q" .Title}},
	Description: {{ printf "%q" .Description}},
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
`
