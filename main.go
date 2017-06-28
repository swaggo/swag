package main

import (
	"github.com/swag-gonic/swag/gen"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()

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
