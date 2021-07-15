package main

import (
	"os"

	"github.com/dbubel/api"
	"github.com/dbubel/config"
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/cli"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.JSONFormatter{})

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		logrus.WithError(err).Fatalln("Error parsing config")
	}

	c := cli.NewCLI("jackstand", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"serve": func() (cli.Command, error) {
			return &api.ServeCommand{
				Cfg: cfg,
				Log: log,
			}, nil
		},
	}

	_, err := c.Run()
	if err != nil {
		logrus.WithError(err).Fatalln("Error running serve command")
	}
}
