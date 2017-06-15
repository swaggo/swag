package main

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/easonlin404/gin-swagger/parse"
	"os"
	"github.com/easonlin404/gin-swagger/gen"
)

var framework string

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "framework,f",
			Value:       "gin",
			Usage:       "web framework for the swagger",
			Destination: &framework,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "create doc.go",
			Action: func(c *cli.Context) error {
				if framework == "gin" {
					gen.New().Build()
				} else {
					fmt.Printf("%v not support.\n", framework)
				}
				return nil
			},
		},
		{
			Name:    "update",
			Aliases: []string{"a"},
			Usage:   "update doc.go",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	}

	app.Run(os.Args)
}
