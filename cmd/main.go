package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
	"mlsql.tech/allwefantasy/deploy/pkg/version"
	"os"
)

var logger = utils.GetLogger("byzer-k8s-deploy")

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

/**
byzer-k8s-deploy run \
--kube-config /tmp/.. \

--engine-name xxxx   \
--engine-image xxxx   \
--engine-executor-core-num xxxx   \
--engine-executor-num xxxx   \
--engine-executor-memory xxxx \

--engine-driver-core-num xxxx   \
--engine-driver-memory xxxx \


--engine-access-token xxxx   \
--engine-jar-path-in-container xxxx   \


--storage-name  xxxx \
--storage-meta-url  xxxxx \
--storage-mount-point  xxxx \
--storage-access-key xxxx     \
--storage-secret-key  xxxx \
k8s
*/
func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name: "version", Aliases: []string{"V"},
		Usage: "print only the version",
	}

	app := cli.App{
		Name:                 "byzer-k8s-deploy",
		Usage:                "CLI to deploy Byzer Engine in K8s",
		Version:              version.Version(),
		Copyright:            "Apache License V2",
		EnableBashCompletion: true,
		Flags:                globalFlags(),
		Commands: []*cli.Command{
			runFlags(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
