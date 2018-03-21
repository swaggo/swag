package main

import (
	"os"

	"github.com/swaggo/swag/gen"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "v1.2.0"
	app.Usage = "Automatically generate RESTful API documentation with Swagger 2.0 for Go."
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Create docs.go",
			Action: func(c *cli.Context) error {
				dir := c.String("dir")
				mainAPIFile := c.String("generalInfo")
				swaggerConfDir := c.String("swagger")
				gen.New().Build(dir, mainAPIFile, swaggerConfDir)
				return nil
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
					Name:  "swagger, s",
					Value: "./docs/swagger",
					Usage: "Output the swagger conf for json and yaml",
				},
			},
		},
	}
	app.Run(os.Args)
}
