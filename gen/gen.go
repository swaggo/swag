package gen

import (
	"bufio"
	"bytes"
	"embed"
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

	jsoniter "github.com/json-iterator/go"

	v2 "github.com/go-openapi/spec"
	v3 "github.com/sv-tools/openapi/spec"

	"github.com/nguyennm96/swag/v2"
	"sigs.k8s.io/yaml"
)

var open = os.Open

// DefaultOverridesFile is the location swagger will look for type overrides.
const DefaultOverridesFile = ".swaggo"

type genTypeWriter func(*Config, interface{}) error

// Gen presents a generate tool for swag.
type Gen struct {
	json          func(data interface{}) ([]byte, error)
	jsonIndent    func(data interface{}) ([]byte, error)
	jsonToYAML    func(data []byte) ([]byte, error)
	outputTypeMap map[string]genTypeWriter
	debug         Debugger
}

//go:embed src/*.tmpl
var tmpl embed.FS

// Debugger is the interface that wraps the basic Printf method.
type Debugger interface {
	Printf(format string, v ...interface{})
}

// New creates a new Gen.
func New() *Gen {
	gen := Gen{
		json: json.Marshal,
		jsonIndent: func(data interface{}) ([]byte, error) {
			return jsoniter.ConfigCompatibleWithStandardLibrary.MarshalIndent(&data, "", "    ")
		},
		jsonToYAML: yaml.JSONToYAML,
		debug:      log.New(os.Stdout, "", log.LstdFlags),
	}

	gen.outputTypeMap = map[string]genTypeWriter{
		"go":   gen.writeDoc,
		"json": gen.writeJSON,
		"yaml": gen.writeYAML,
		"yml":  gen.writeYAML,
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

	// outputs only specific extension
	ParseExtension string

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

	// LeftTemplateDelim defines the left delimiter for the template generation
	LeftTemplateDelim string

	// RightTemplateDelim defines the right delimiter for the template generation
	RightTemplateDelim string

	// GenerateOpenAPI3Doc if true, OpenAPI V3.1 spec will be generated
	GenerateOpenAPI3Doc bool

	// PackageName defines package name of generated `docs.go`
	PackageName string

	// CollectionFormat set default collection format
	CollectionFormat string
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

	if config.LeftTemplateDelim == "" {
		config.LeftTemplateDelim = "{{"
	}

	if config.RightTemplateDelim == "" {
		config.RightTemplateDelim = "}}"
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
		swag.SetParseExtension(config.ParseExtension),
		swag.SetCodeExamplesDirectory(config.CodeExampleFilesDir),
		swag.SetStrict(config.Strict),
		swag.SetOverrides(overrides),
		swag.ParseUsingGoList(config.ParseGoList),
		swag.SetTags(config.Tags),
		swag.GenerateOpenAPI3Doc(config.GenerateOpenAPI3Doc),
		swag.SetCollectionFormat(config.CollectionFormat),
	)

	p.PropNamingStrategy = config.PropNamingStrategy
	p.ParseVendor = config.ParseVendor
	p.ParseInternal = config.ParseInternal
	p.RequiredByDefault = config.RequiredByDefault

	if err := p.ParseAPIMultiSearchDir(searchDirs, config.MainAPIFile, config.ParseDepth); err != nil {
		return err
	}

	if err := os.MkdirAll(config.OutputDir, os.ModePerm); err != nil {
		return err
	}

	if config.GenerateOpenAPI3Doc {
		return g.writeOpenAPI(config, p.GetOpenAPI())
	}

	return g.writeOpenAPI(config, p.GetSwagger())
}

func (g *Gen) writeOpenAPI(config *Config, doc interface{}) error {
	for _, outputType := range config.OutputTypes {
		outputType = strings.ToLower(strings.TrimSpace(outputType))
		if typeWriter, ok := g.outputTypeMap[outputType]; ok {
			if err := typeWriter(config, doc); err != nil {
				return err
			}
		} else {
			log.Printf("output type '%s' not supported", outputType)
		}
	}

	return nil
}

func (g *Gen) writeDoc(config *Config, doc interface{}) error {
	var filename = "docs.go"

	if config.InstanceName != swag.Name {
		filename = config.InstanceName + "_" + filename
	}

	docFileName := path.Join(config.OutputDir, filename)

	absOutputDir, err := filepath.Abs(config.OutputDir)
	if err != nil {
		return err
	}

	var packageName string
	if len(config.PackageName) > 0 {
		packageName = config.PackageName
	} else {
		packageName = filepath.Base(absOutputDir)
		packageName = strings.ReplaceAll(packageName, "-", "_")
	}

	docs, err := os.Create(docFileName)
	if err != nil {
		return err
	}
	defer docs.Close()

	// Write doc
	switch spec := doc.(type) {
	case *v2.Swagger:
		err = g.writeGoDoc(packageName, docs, spec, config)
		if err != nil {
			return err

		}
	case *v3.OpenAPI:
		err = g.writeGoDocV3(packageName, docs, spec, config)
		if err != nil {
			return err
		}
	}
	g.debug.Printf("create docs.go at %+v", docFileName)

	return nil
}

func (g *Gen) writeJSON(config *Config, spec interface{}) error {
	var filename = "swagger.json"

	if config.InstanceName != swag.Name {
		filename = config.InstanceName + "_" + filename
	}

	jsonFileName := path.Join(config.OutputDir, filename)

	b, err := g.jsonIndent(spec)
	if err != nil {
		return err
	}

	err = g.writeFile(b, jsonFileName)
	if err != nil {
		return err
	}

	g.debug.Printf("create swagger.json at %+v", jsonFileName)

	return nil
}

func (g *Gen) writeYAML(config *Config, swagger interface{}) error {
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

	g.debug.Printf("create swagger.yaml at %+v", yamlFileName)

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

func (g *Gen) writeGoDoc(packageName string, output io.Writer, swagger *v2.Swagger, config *Config) error {
	generator, err := template.New("oas2.tmpl").Funcs(template.FuncMap{
		"printDoc": func(v string) string {
			// Add schemes
			v = "{\n    \"schemes\": " + config.LeftTemplateDelim + " marshal .Schemes " + config.RightTemplateDelim + "," + v[1:]
			// Sanitize backticks
			return strings.Replace(v, "`", "`+\"`\"+`", -1)
		},
	}).ParseFS(tmpl, "src/*.tmpl")
	if err != nil {
		return err
	}

	swaggerSpec := &v2.Swagger{
		VendorExtensible: swagger.VendorExtensible,
		SwaggerProps: v2.SwaggerProps{
			ID:       swagger.ID,
			Consumes: swagger.Consumes,
			Produces: swagger.Produces,
			Swagger:  swagger.Swagger,
			Info: &v2.Info{
				VendorExtensible: swagger.Info.VendorExtensible,
				InfoProps: v2.InfoProps{
					Description:    config.LeftTemplateDelim + "escape .Description" + config.RightTemplateDelim,
					Title:          config.LeftTemplateDelim + ".Title" + config.RightTemplateDelim,
					TermsOfService: swagger.Info.TermsOfService,
					Contact:        swagger.Info.Contact,
					License:        swagger.Info.License,
					Version:        config.LeftTemplateDelim + ".Version" + config.RightTemplateDelim,
				},
			},
			Host:                config.LeftTemplateDelim + ".Host" + config.RightTemplateDelim,
			BasePath:            config.LeftTemplateDelim + ".BasePath" + config.RightTemplateDelim,
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
		Timestamp          time.Time
		Doc                string
		Host               string
		PackageName        string
		BasePath           string
		Title              string
		Description        string
		Version            string
		InstanceName       string
		Schemes            []string
		GeneratedTime      bool
		LeftTemplateDelim  string
		RightTemplateDelim string
	}{
		Timestamp:          time.Now(),
		GeneratedTime:      config.GeneratedTime,
		Doc:                string(buf),
		Host:               swagger.Host,
		PackageName:        packageName,
		BasePath:           swagger.BasePath,
		Schemes:            swagger.Schemes,
		Title:              swagger.Info.Title,
		Description:        swagger.Info.Description,
		Version:            swagger.Info.Version,
		InstanceName:       config.InstanceName,
		LeftTemplateDelim:  config.LeftTemplateDelim,
		RightTemplateDelim: config.RightTemplateDelim,
	})
	if err != nil {
		return err
	}

	code := g.formatSource(buffer.Bytes())

	// write
	_, err = output.Write(code)

	return err
}

func (g *Gen) writeGoDocV3(packageName string, output io.Writer, openAPI *v3.OpenAPI, config *Config) error {
	generator, err := template.New("oas3.tmpl").Funcs(template.FuncMap{
		"printDoc": func(v string) string {
			// Add schemes
			v = "{\n    \"schemes\": " + config.LeftTemplateDelim + " marshal .Schemes " + config.RightTemplateDelim + "," + v[1:]
			// Sanitize backticks
			return strings.Replace(v, "`", "`+\"`\"+`", -1)
		},
	}).ParseFS(tmpl, "src/*.tmpl")
	if err != nil {
		return err
	}

	openAPISpec := v3.OpenAPI{
		Components: openAPI.Components,
		OpenAPI:    openAPI.OpenAPI,
		Info: &v3.Extendable[v3.Info]{
			Spec: &v3.Info{
				Description:    config.LeftTemplateDelim + "escape .Description" + config.RightTemplateDelim,
				Title:          config.LeftTemplateDelim + ".Title" + config.RightTemplateDelim,
				Version:        config.LeftTemplateDelim + ".Version" + config.RightTemplateDelim,
				TermsOfService: openAPI.Info.Spec.TermsOfService,
				Contact:        openAPI.Info.Spec.Contact,
				License:        openAPI.Info.Spec.License,
				Summary:        openAPI.Info.Spec.Summary,
			},
			Extensions: openAPI.Info.Extensions,
		},
		ExternalDocs:      openAPI.ExternalDocs,
		Paths:             openAPI.Paths,
		WebHooks:          openAPI.WebHooks,
		JsonSchemaDialect: openAPI.JsonSchemaDialect,
		Security:          openAPI.Security,
		Tags:              openAPI.Tags,
		Servers:           openAPI.Servers,
	}

	// crafted docs.json
	buf, err := g.jsonIndent(openAPISpec)
	if err != nil {
		return err
	}

	buffer := &bytes.Buffer{}

	err = generator.Execute(buffer, struct {
		Timestamp          time.Time
		Doc                string
		PackageName        string
		Title              string
		Description        string
		Version            string
		InstanceName       string
		GeneratedTime      bool
		LeftTemplateDelim  string
		RightTemplateDelim string
	}{
		Timestamp:          time.Now(),
		GeneratedTime:      config.GeneratedTime,
		Doc:                string(buf),
		PackageName:        packageName,
		Title:              openAPI.Info.Spec.Title,
		Description:        openAPI.Info.Spec.Description,
		Version:            openAPI.Info.Spec.Version,
		InstanceName:       config.InstanceName,
		LeftTemplateDelim:  config.LeftTemplateDelim,
		RightTemplateDelim: config.RightTemplateDelim,
	})
	if err != nil {
		return err
	}

	code := g.formatSource(buffer.Bytes())

	// write
	_, err = output.Write(code)

	return err
}
