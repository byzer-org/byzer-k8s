package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
	"mlsql.tech/allwefantasy/deploy/pkg/version"
	"os"
)

var logger = utils.GetLogger("mlsql-deploy")

func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"debug", "v"},
			Usage:   "enable debug log",
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
			Usage:   "only warning and errors",
		},
		&cli.BoolFlag{
			Name:  "trace",
			Usage: "enable trace log",
		},
	}
}

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name: "version", Aliases: []string{"V"},
		Usage: "print only the version",
	}

	app := cli.App{
		Name:                 "mlsql-deploy",
		Usage:                "Cli to deploy mlsql engine in K8s",
		Version:              version.Version(),
		Copyright:            "Apache V2",
		EnableBashCompletion: true,
		Flags:                globalFlags(),
		Commands: []*cli.Command{

		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
