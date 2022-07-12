package main

import (
	"github.com/urfave/cli/v2"
)

func delete(c *cli.Context) error {

	//k8sConfig := meta.BuildKubeConfig(c)
	//engineName := c.String("engine-name")



	return nil
}

func deleteFlags() *cli.Command {
	cmd := &cli.Command{
		Name:      "delete",
		Usage:     "undeploy byzer engine from ",
		ArgsUsage: "k8s",
		Action:    delete,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "kube-config",
				Required: true,
				Usage:    "path of kube config file",
			},
			&cli.StringFlag{
				Name:     "engine-name",
				Required: true,
				Usage:    "the engine-name we want to undeploy",
			},
		},
	}

	return cmd
}
