package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
)

func delete(c *cli.Context) error {

	k8sConfig := meta.BuildKubeConfig(c)
	engineName := c.String("engine-name")

	metaConfig := meta.MetaConfig{
		K8sConfig:    k8sConfig,
		EngineConfig: meta.EngineConfig{Name: engineName},
	}

	executor := meta.CreateKubeExecutor(&metaConfig.K8sConfig)

	logger.Info(fmt.Sprintf("delete ingress:%s", metaConfig.EngineConfig.Name))
	executor.DeleteAny([]string{"ingress", metaConfig.EngineConfig.Name})

	logger.Info(fmt.Sprintf("delete service:%s", metaConfig.EngineConfig.Name))
	executor.DeleteAny([]string{"service", metaConfig.EngineConfig.Name})

	logger.Info(fmt.Sprintf("delete deployment:%s", metaConfig.EngineConfig.Name))
	executor.DeleteAny([]string{"deployment", metaConfig.EngineConfig.Name})

	logger.Info(fmt.Sprintf("delete configmap:%s", fmt.Sprintf("%s-core-site-xml", metaConfig.EngineConfig.Name)))
	executor.DeleteAny([]string{"configmap", fmt.Sprintf("%s-core-site-xml", metaConfig.EngineConfig.Name)})

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
