package main

import (
	"os"

	"github.com/swaggo/swag/gen"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "v1.1.1"
	app.Usage = "Automatically generate RESTful API documentation with Swagger 2.0 for Go."

	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "create docs.go",
			Action: func(c *cli.Context) error {
				searchDir := "./"
				mainApiFile := "./main.go"
				gen.New().Build(searchDir, mainApiFile)
				return nil
			},
		},
	}

	app.Run(os.Args)
}
