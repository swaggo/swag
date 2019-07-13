package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/swaggo/swag"
	"github.com/swaggo/swag/gen"
	"github.com/urfave/cli"
)

const searchDirFlag = "dir"
const generalInfoFlag = "generalInfo"
const propertyStrategyFlag = "propertyStrategy"
const outputFlag = "output"
const parseVendorFlag = "parseVendor"
const parseDependency = "parseDependency"
const markdownFilesDirFlag = "markdownFiles"

func main() {
	app := cli.NewApp()
	app.Version = swag.Version
	app.Usage = "Automatically generate RESTful API documentation with Swagger 2.0 for Go."
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Create docs.go",
			Action: func(c *cli.Context) error {
				searchDir := c.String(searchDirFlag)
				mainAPIFile := c.String(generalInfoFlag)
				strategy := c.String(propertyStrategyFlag)
				outputDir := c.String(outputFlag)
				parseVendor := c.Bool(parseVendorFlag)
				parseDependency := c.Bool(parseDependency)
				markdownFilesDir := c.String(markdownFilesDirFlag)

				switch strategy {
				case swag.CamelCase, swag.SnakeCase, swag.PascalCase:
				default:
					return errors.Errorf("not supported %s propertyStrategy", strategy)
				}

				return gen.New().Build(&gen.Config{
					SearchDir:          searchDir,
					MainAPIFile:        mainAPIFile,
					PropNamingStrategy: strategy,
					OutputDir:          outputDir,
					ParseVendor:        parseVendor,
					ParseDependency:    parseDependency,
					MarkdownFilesDir:   markdownFilesDir,
				})
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "generalInfo, g",
					Value: "main.go",
					Usage: "Go file path in which 'swagger general API Info' is written",
				},
				cli.StringFlag{
					Name:  "dir, d",
					Value: "./",
					Usage: "Directory you want to parse",
				},
				cli.StringFlag{
					Name:  "propertyStrategy, p",
					Value: "camelcase",
					Usage: "Property Naming Strategy like snakecase,camelcase,pascalcase",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "./docs",
					Usage: "Output directory for al the generated files(swagger.json, swagger.yaml and doc.go)",
				},
				cli.BoolFlag{
					Name:  "parseVendor",
					Usage: "Parse go files in 'vendor' folder, disabled by default",
				},
				cli.BoolFlag{
					Name:  "parseDependency",
					Usage: "Parse go files in outside dependency folder, disabled by default",
				},
				cli.StringFlag{
					Name:  "markdownFiles, md",
					Value: "",
					Usage: "Parse folder containing markdown files to use as description, disabled by default",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
