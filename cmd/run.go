package main

import (
	"github.com/urfave/cli/v2"
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"mlsql.tech/allwefantasy/deploy/pkg/operator"
)

func run(c *cli.Context) error {

	engineConfig := meta.BuildEngineConfig(c)
	k8sConfig := meta.BuildKubeConfig(c)
	var storageConfig = meta.BuildStorageConfig(c)

	metaConfig := meta.MetaConfig{
		K8sConfig:     k8sConfig,
		EngineConfig:  engineConfig,
		StorageConfig: storageConfig,
	}

	// Read config from file
	extraConf, err := meta.BuildExtraEngineConfig(c)
	if err != nil {
		return err
	}

	verbose := c.IsSet("verbose") && c.Bool("verbose")

	executor := meta.CreateKubeExecutor(&metaConfig.K8sConfig)

	// Step1: configure CM, so we can support JuiceFS FileSystem
	cmOp := operator.NewConfigMapOp(executor, metaConfig, extraConf)

	cmOpErr := cmOp.Execute(verbose)

	if cmOpErr != nil {
		return cmOpErr
	}

	// Step2: Add role and service account in k8s
	roleOp := operator.NewRole(executor, metaConfig)
	bindRoleErr := roleOp.Execute(verbose)
	if bindRoleErr != nil {
		return bindRoleErr
	}

	// Step3: Deploy Byzer Engine
	deployOp := operator.NewDeploy(executor, metaConfig, extraConf)
	deployErr := deployOp.Execute(verbose)
	if deployErr != nil {
		return deployErr
	}
	return nil
}

func engineFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "engine-name",
			Required: true,
			Usage:    "The name of Byzer Engine",
		},
		&cli.StringFlag{
			Name:     "engine-version",
			Required: false,
			Usage:    "The version of Byzer Engine",
		},
		&cli.StringFlag{
			Name:     "engine-image",
			Required: false,
			Usage:    "The name of Byzer Engine",
		},
		&cli.StringFlag{
			Name:     "engine-service-account-name",
			Required: false,
			Usage:    "the service account name of byzer engine in K8s",
			Value:    "default",
		},
		&cli.StringFlag{
			Name:     "engine-role-name",
			Required: false,
			Usage:    "the  role name of byzer engine in K8s",
			Value:    "byzer-role",
		},
		&cli.StringFlag{
			Name:     "engine-role-binding-name",
			Required: false,
			Usage:    "the role binding name of byzer engine in K8s",
			Value:    "byzer-role-binding",
		},
		&cli.StringFlag{
			Name:     "engine-namespace",
			Required: false,
			Usage:    "The namespace Byzer Engine deployed. ",
			Value:    "default",
		},
		&cli.Float64Flag{
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
		&cli.Float64Flag{
			Name:  "engine-driver-core-num",
			Value: 4,
			Usage: "the core num of driver",
		},
		&cli.StringFlag{
			Name:  "engine-access-token",
			Value: "",
			Usage: "the access token to protect byzer engine",
		},
		&cli.StringFlag{
			Name:     "engine-jar-path-in-container",
			Value:    "",
			Usage:    "The path of byzer engine jar in docker image",
			Required: false,
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
		Usage:     "run Byzer Engine on specific resource manager framework e.g. K8s, Yarn",
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

	cmd.Flags = append(cmd.Flags, &cli.StringFlag{
		Name:     "engine-config",
		Required: false,
		Usage:    "path extra engine config file",
	})
	return cmd
}
