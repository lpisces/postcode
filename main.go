package main

import (
	"github.com/lpisces/postcode/cmd/collect"
	"github.com/lpisces/postcode/cmd/serve/mvc"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "bootstrap"
	app.Usage = "bootstrap for website server development"

	app.Commands = []cli.Command{
		{
			Name:    "serve",
			Aliases: []string{"s"},
			Usage:   "start web server",
			Action:  mvc.Run,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "env, e",
					Usage: "set run env",
					Value: "development",
				},
				cli.StringFlag{
					Name:  "port, p",
					Usage: "listen port",
				},
				cli.StringFlag{
					Name:  "bind, b",
					Usage: "bind host",
				},
				cli.StringFlag{
					Name:  "config, c",
					Usage: "load config file",
				},
				cli.StringFlag{
					Name:  "source, s",
					Usage: "datasource file",
				},
			},
		},
		{
			Name:    "collect",
			Aliases: []string{"c"},
			Usage:   "collect cn postcode",
			Action:  collect.Run,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "key, k",
					Usage: "app key",
					Value: "",
				},
				cli.StringFlag{
					Name:  "env, e",
					Usage: "set run env",
					Value: "development",
				},
				cli.StringFlag{
					Name:  "config, c",
					Usage: "load config file",
				},
				cli.StringFlag{
					Name:  "cache",
					Usage: "cache path",
				},
				cli.StringFlag{
					Name:  "output, o",
					Usage: "output file",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
