package main

import (
	"github.com/urfave/cli/v2"
)

func run(c *cli.Context) error {
	return nil
}

func engineFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "engine-name",
			Required: true,
			Usage:    "The name of MLSQL Engine",
		},
		&cli.StringFlag{
			Name:     "engine-image",
			Required: true,
			Usage:    "The name of MLSQL Engine",
		},
		&cli.IntFlag{
			Name:  "engine-executor-core-num",
			Value: 1,
			Usage: "the core num of every executor",
		},
		&cli.IntFlag{
			Name:  "engine-executor-num",
			Value: 1,
			Usage: "the size of executor",
		},

		&cli.IntFlag{
			Name:  "engine-executor-memory",
			Value: 1024,
			Usage: "memory size for executor in MB",
		},

		&cli.IntFlag{
			Name:  "engine-driver-memory",
			Value: 1024,
			Usage: "memory size for driver in MB",
		},
		&cli.IntFlag{
			Name:  "engine-driver-core-num",
			Value: 4,
			Usage: "the core num of driver",
		},
		&cli.StringFlag{
			Name:  "engine-access-token",
			Value: "mlsql",
			Usage: "the access token to protect mlsql engine",
		},
		&cli.StringFlag{
			Name:  "engine-jar-path-in-container",
			Value: "",
			Usage: "The path of mlsql engine jar in docker image",
		},
	}
}

func storageFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "storage-name",
			Value: "",
			Usage: "The name of storage",
		},
		&cli.StringFlag{
			Name:  "storage-meta-url",
			Value: "",
			Usage: "the url of meta storage ",
		},
		&cli.StringFlag{
			Name:  "storage-mount-point",
			Value: "",
			Usage: "the mount point for object store/HDFS",
		},
		&cli.StringFlag{
			Name:  "storage-access-key",
			Value: "",
			Usage: "the access key of object store",
		},
		&cli.StringFlag{
			Name:  "storage-secret-key",
			Value: "",
			Usage: "the secret-key of object store",
		},
	}
}

func runFlags() *cli.Command {
	cmd := &cli.Command{
		Name:      "run",
		Usage:     "run MLSQL Engine on specific resource manager framework e.g. K8s, Yarn",
		ArgsUsage: "k8s",
		Action:    run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "kube-config",
				Required: true,
				Usage:    "path of kube config file",
			},
		},
	}
	cmd.Flags = append(cmd.Flags, engineFlags()...)
	cmd.Flags = append(cmd.Flags, storageFlags()...)
	return cmd
}
