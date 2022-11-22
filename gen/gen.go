package gen

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
	"github.com/swaggo/swag"
)

var open = os.Open

// DefaultOverridesFile is the location swagger will look for type overrides.
const DefaultOverridesFile = ".swaggo"

type genTypeWriter func(*Config, *spec.Swagger) error

// Gen presents a generate tool for swag.
type Gen struct {
	json          func(data interface{}) ([]byte, error)
	jsonIndent    func(data interface{}) ([]byte, error)
	jsonToYAML    func(data []byte) ([]byte, error)
	outputTypeMap map[string]genTypeWriter
	debug         Debugger
}

// Debugger is the interface that wraps the basic Printf method.
type Debugger interface {
	Printf(format string, v ...interface{})
}

// New creates a new Gen.
func New() *Gen {
	gen := Gen{
		json: json.Marshal,
		jsonIndent: func(data interface{}) ([]byte, error) {
			return json.MarshalIndent(data, "", "    ")
		},
		jsonToYAML: yaml.JSONToYAML,
		debug:      log.New(os.Stdout, "", log.LstdFlags),
	}

	gen.outputTypeMap = map[string]genTypeWriter{
		"go":   gen.writeDocSwagger,
		"json": gen.writeJSONSwagger,
		"yaml": gen.writeYAMLSwagger,
		"yml":  gen.writeYAMLSwagger,
	}

	return &gen
}

// Config presents Gen configurations.
type Config struct {
	Debugger swag.Debugger

	// SearchDir the swag would parse,comma separated if multiple
	SearchDir string

	// excludes dirs and files in SearchDir,comma separated
	Excludes string

	// OutputDir represents the output directory for all the generated files
	OutputDir string

	// OutputTypes define types of files which should be generated
	OutputTypes []string

	// MainAPIFile the Go file path in which 'swagger general API Info' is written
	MainAPIFile string

	// PropNamingStrategy represents property naming strategy like snake case,camel case,pascal case
	PropNamingStrategy string

	// MarkdownFilesDir used to find markdown files, which can be used for tag descriptions
	MarkdownFilesDir string

	// CodeExampleFilesDir used to find code example files, which can be used for x-codeSamples
	CodeExampleFilesDir string

	// InstanceName is used to get distinct names for different swagger documents in the
	// same project. The default value is "swagger".
	InstanceName string

	// ParseDepth dependency parse depth
	ParseDepth int

	// ParseVendor whether swag should be parse vendor folder
	ParseVendor bool

	// ParseDependencies whether swag should be parse outside dependency folder
	ParseDependency bool

	// ParseInternal whether swag should parse internal packages
	ParseInternal bool

	// Strict whether swag should error or warn when it detects cases which are most likely user errors
	Strict bool

	// GeneratedTime whether swag should generate the timestamp at the top of docs.go
	GeneratedTime bool

	// RequiredByDefault set validation required for all fields by default
	RequiredByDefault bool

	// OverridesFile defines global type overrides.
	OverridesFile string

	// ParseGoList whether swag use go list to parse dependency
	ParseGoList bool

	// include only tags mentioned when searching, comma separated
	Tags string
}

// Build builds swagger json file  for given searchDir and mainAPIFile. Returns json.
func (g *Gen) Build(config *Config) error {
	if config.Debugger != nil {
		g.debug = config.Debugger
	}
	if config.InstanceName == "" {
		config.InstanceName = swag.Name
	}

	searchDirs := strings.Split(config.SearchDir, ",")
	for _, searchDir := range searchDirs {
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return fmt.Errorf("dir: %s does not exist", searchDir)
		}
	}

	var overrides map[string]string

	if config.OverridesFile != "" {
		overridesFile, err := open(config.OverridesFile)
		if err != nil {
			// Don't bother reporting if the default file is missing; assume there are no overrides
			if !(config.OverridesFile == DefaultOverridesFile && os.IsNotExist(err)) {
				return fmt.Errorf("could not open overrides file: %w", err)
			}
		} else {
			g.debug.Printf("Using overrides from %s", config.OverridesFile)

			overrides, err = parseOverrides(overridesFile)
			if err != nil {
				return err
			}
		}
	}

	g.debug.Printf("Generate swagger docs....")

	p := swag.New(
		swag.SetParseDependency(config.ParseDependency),
		swag.SetMarkdownFileDirectory(config.MarkdownFilesDir),
		swag.SetDebugger(config.Debugger),
		swag.SetExcludedDirsAndFiles(config.Excludes),
		swag.SetCodeExamplesDirectory(config.CodeExampleFilesDir),
		swag.SetStrict(config.Strict),
		swag.SetOverrides(overrides),
		swag.ParseUsingGoList(config.ParseGoList),
		swag.SetTags(config.Tags),
	)

	p.PropNamingStrategy = config.PropNamingStrategy
	p.ParseVendor = config.ParseVendor
	p.ParseInternal = config.ParseInternal
	p.RequiredByDefault = config.RequiredByDefault

	if err := p.ParseAPIMultiSearchDir(searchDirs, config.MainAPIFile, config.ParseDepth); err != nil {
		return err
	}

	swagger := p.GetSwagger()

	if err := os.MkdirAll(config.OutputDir, os.ModePerm); err != nil {
		return err
	}

	for _, outputType := range config.OutputTypes {
		outputType = strings.ToLower(strings.TrimSpace(outputType))
		if typeWriter, ok := g.outputTypeMap[outputType]; ok {
			if err := typeWriter(config, swagger); err != nil {
				return err
			}
		} else {
			log.Printf("output type '%s' not supported", outputType)
		}
	}

	return nil
}

func (g *Gen) writeDocSwagger(config *Config, swagger *spec.Swagger) error {
	var filename = "docs.go"

	if config.InstanceName != swag.Name {
		filename = config.InstanceName + "_" + filename
	}

	docFileName := path.Join(config.OutputDir, filename)

	absOutputDir, err := filepath.Abs(config.OutputDir)
	if err != nil {
		return err
	}

	packageName := filepath.Base(absOutputDir)

	docs, err := os.Create(docFileName)
	if err != nil {
		return err
	}
	defer docs.Close()

	// Write doc
	err = g.writeGoDoc(packageName, docs, swagger, config)
	if err != nil {
		return err
	}

	g.debug.Printf("create docs.go at  %+v", docFileName)

	return nil
}

func (g *Gen) writeJSONSwagger(config *Config, swagger *spec.Swagger) error {
	var filename = "swagger.json"

	if config.InstanceName != swag.Name {
		filename = config.InstanceName + "_" + filename
	}

	jsonFileName := path.Join(config.OutputDir, filename)

	b, err := g.jsonIndent(swagger)
	if err != nil {
		return err
	}

	err = g.writeFile(b, jsonFileName)
	if err != nil {
		return err
	}

	g.debug.Printf("create swagger.json at  %+v", jsonFileName)

	return nil
}

func (g *Gen) writeYAMLSwagger(config *Config, swagger *spec.Swagger) error {
	var filename = "swagger.yaml"

	if config.InstanceName != swag.Name {
		filename = config.InstanceName + "_" + filename
	}

	yamlFileName := path.Join(config.OutputDir, filename)

	b, err := g.json(swagger)
	if err != nil {
		return err
	}

	y, err := g.jsonToYAML(b)
	if err != nil {
		return fmt.Errorf("cannot covert json to yaml error: %s", err)
	}

	err = g.writeFile(y, yamlFileName)
	if err != nil {
		return err
	}

	g.debug.Printf("create swagger.yaml at  %+v", yamlFileName)

	return nil
}

func (g *Gen) writeFile(b []byte, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(b)

	return err
}

func (g *Gen) formatSource(src []byte) []byte {
	code, err := format.Source(src)
	if err != nil {
		code = src // Formatter failed, return original code.
	}

	return code
}

// Read and parse the overrides file.
func parseOverrides(r io.Reader) (map[string]string, error) {
	overrides := make(map[string]string)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments
		if len(line) > 1 && line[0:2] == "//" {
			continue
		}

		parts := strings.Fields(line)

		switch len(parts) {
		case 0:
			// only whitespace
			continue
		case 2:
			// either a skip or malformed
			if parts[0] != "skip" {
				return nil, fmt.Errorf("could not parse override: '%s'", line)
			}

			overrides[parts[1]] = ""
		case 3:
			// either a replace or malformed
			if parts[0] != "replace" {
				return nil, fmt.Errorf("could not parse override: '%s'", line)
			}

			overrides[parts[1]] = parts[2]
		default:
			return nil, fmt.Errorf("could not parse override: '%s'", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading overrides file: %w", err)
	}

	return overrides, nil
}

func (g *Gen) writeGoDoc(packageName string, output io.Writer, swagger *spec.Swagger, config *Config) error {
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
					Description:    "{{escape .Description}}",
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
		Timestamp     time.Time
		Doc           string
		Host          string
		PackageName   string
		BasePath      string
		Title         string
		Description   string
		Version       string
		InstanceName  string
		Schemes       []string
		GeneratedTime bool
	}{
		Timestamp:     time.Now(),
		GeneratedTime: config.GeneratedTime,
		Doc:           string(buf),
		Host:          swagger.Host,
		PackageName:   packageName,
		BasePath:      swagger.BasePath,
		Schemes:       swagger.Schemes,
		Title:         swagger.Info.Title,
		Description:   swagger.Info.Description,
		Version:       swagger.Info.Version,
		InstanceName:  config.InstanceName,
	})
	if err != nil {
		return err
	}

	code := g.formatSource(buffer.Bytes())

	// write
	_, err = output.Write(code)

	return err
}

var packageTemplate = `// Package {{.PackageName}} GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag{{ if .GeneratedTime }} at
// {{ .Timestamp }}{{ end }}
package {{.PackageName}}

import "github.com/swaggo/swag"

const docTemplate{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }} = ` + "`{{ printDoc .Doc}}`" + `

// SwaggerInfo{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }} holds exported Swagger Info so clients can modify it
var SwaggerInfo{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }} = &swag.Spec{
	Version:     {{ printf "%q" .Version}},
	Host:        {{ printf "%q" .Host}},
	BasePath:    {{ printf "%q" .BasePath}},
	Schemes:     []string{ {{ range $index, $schema := .Schemes}}{{if gt $index 0}},{{end}}{{printf "%q" $schema}}{{end}} },
	Title:       {{ printf "%q" .Title}},
	Description: {{ printf "%q" .Description}},
	InfoInstanceName: {{ printf "%q" .InstanceName }},
	SwaggerTemplate: docTemplate{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }},
}

func init() {
	swag.Register(SwaggerInfo{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }}.InstanceName(), SwaggerInfo{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }})
}
`
