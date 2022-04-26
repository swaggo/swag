package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/swaggo/swag"
	"github.com/swaggo/swag/format"
	"github.com/swaggo/swag/gen"
)

const (
	searchDirFlag        = "dir"
	excludeFlag          = "exclude"
	generalInfoFlag      = "generalInfo"
	propertyStrategyFlag = "propertyStrategy"
	outputFlag           = "output"
	outputTypesFlag      = "outputTypes"
	parseVendorFlag      = "parseVendor"
	parseDependencyFlag  = "parseDependency"
	markdownFilesFlag    = "markdownFiles"
	codeExampleFilesFlag = "codeExampleFiles"
	parseInternalFlag    = "parseInternal"
	generatedTimeFlag    = "generatedTime"
	parseDepthFlag       = "parseDepth"
	instanceNameFlag     = "instanceName"
	overridesFileFlag    = "overridesFile"
)

var initFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    generalInfoFlag,
		Aliases: []string{"g"},
		Value:   "main.go",
		Usage:   "Go file path in which 'swagger general API Info' is written",
	},
	&cli.StringFlag{
		Name:    searchDirFlag,
		Aliases: []string{"d"},
		Value:   "./",
		Usage:   "Directories you want to parse,comma separated and general-info file must be in the first one",
	},
	&cli.StringFlag{
		Name:  excludeFlag,
		Usage: "Exclude directories and files when searching, comma separated",
	},
	&cli.StringFlag{
		Name:    propertyStrategyFlag,
		Aliases: []string{"p"},
		Value:   swag.CamelCase,
		Usage:   "Property Naming Strategy like " + swag.SnakeCase + "," + swag.CamelCase + "," + swag.PascalCase,
	},
	&cli.StringFlag{
		Name:    outputFlag,
		Aliases: []string{"o"},
		Value:   "./docs",
		Usage:   "Output directory for all the generated files(swagger.json, swagger.yaml and docs.go)",
	},
	&cli.StringFlag{
		Name:    outputTypesFlag,
		Aliases: []string{"ot"},
		Value:   "go,json,yaml",
		Usage:   "Output types of generated files (docs.go, swagger.json, swagger.yaml) like go,json,yaml",
	},
	&cli.BoolFlag{
		Name:  parseVendorFlag,
		Usage: "Parse go files in 'vendor' folder, disabled by default",
	},
	&cli.BoolFlag{
		Name:    parseDependencyFlag,
		Aliases: []string{"pd"},
		Usage:   "Parse go files inside dependency folder, disabled by default",
	},
	&cli.StringFlag{
		Name:    markdownFilesFlag,
		Aliases: []string{"md"},
		Value:   "",
		Usage:   "Parse folder containing markdown files to use as description, disabled by default",
	},
	&cli.StringFlag{
		Name:    codeExampleFilesFlag,
		Aliases: []string{"cef"},
		Value:   "",
		Usage:   "Parse folder containing code example files to use for the x-codeSamples extension, disabled by default",
	},
	&cli.BoolFlag{
		Name:  parseInternalFlag,
		Usage: "Parse go files in internal packages, disabled by default",
	},
	&cli.BoolFlag{
		Name:  generatedTimeFlag,
		Usage: "Generate timestamp at the top of docs.go, disabled by default",
	},
	&cli.IntFlag{
		Name:  parseDepthFlag,
		Value: 100,
		Usage: "Dependency parse depth",
	},
	&cli.StringFlag{
		Name:  instanceNameFlag,
		Value: "",
		Usage: "This parameter can be used to name different swagger document instances. It is optional.",
	},
	&cli.StringFlag{
		Name:  overridesFileFlag,
		Value: gen.DefaultOverridesFile,
		Usage: "File to read global type overrides from.",
	},
}

func initAction(ctx *cli.Context) error {
	strategy := ctx.String(propertyStrategyFlag)

	switch strategy {
	case swag.CamelCase, swag.SnakeCase, swag.PascalCase:
	default:
		return fmt.Errorf("not supported %s propertyStrategy", strategy)
	}

	outputTypes := strings.Split(ctx.String(outputTypesFlag), ",")
	if len(outputTypes) == 0 {
		return fmt.Errorf("no output types specified")
	}

	return gen.New().Build(&gen.Config{
		SearchDir:           ctx.String(searchDirFlag),
		Excludes:            ctx.String(excludeFlag),
		MainAPIFile:         ctx.String(generalInfoFlag),
		PropNamingStrategy:  strategy,
		OutputDir:           ctx.String(outputFlag),
		OutputTypes:         outputTypes,
		ParseVendor:         ctx.Bool(parseVendorFlag),
		ParseDependency:     ctx.Bool(parseDependencyFlag),
		MarkdownFilesDir:    ctx.String(markdownFilesFlag),
		ParseInternal:       ctx.Bool(parseInternalFlag),
		GeneratedTime:       ctx.Bool(generatedTimeFlag),
		CodeExampleFilesDir: ctx.String(codeExampleFilesFlag),
		ParseDepth:          ctx.Int(parseDepthFlag),
		InstanceName:        ctx.String(instanceNameFlag),
		OverridesFile:       ctx.String(overridesFileFlag),
	})
}

func main() {
	app := cli.NewApp()
	app.Version = swag.Version
	app.Usage = "Automatically generate RESTful API documentation with Swagger 2.0 for Go."
	app.Commands = []*cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Create docs.go",
			Action:  initAction,
			Flags:   initFlags,
		},
		{
			Name:    "fmt",
			Aliases: []string{"f"},
			Usage:   "format swag comments",
			Action: func(c *cli.Context) error {
				searchDir := c.String(searchDirFlag)
				excludeDir := c.String(excludeFlag)
				mainFile := c.String(generalInfoFlag)

				return format.New().Build(&format.Config{
					SearchDir: searchDir,
					Excludes:  excludeDir,
					MainFile:  mainFile,
				})
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    searchDirFlag,
					Aliases: []string{"d"},
					Value:   "./",
					Usage:   "Directories you want to parse,comma separated and general-info file must be in the first one",
				},
				&cli.StringFlag{
					Name:  excludeFlag,
					Usage: "Exclude directories and files when searching, comma separated",
				},
				&cli.StringFlag{
					Name:    generalInfoFlag,
					Aliases: []string{"g"},
					Value:   "main.go",
					Usage:   "Go file path in which 'swagger general API Info' is written",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
