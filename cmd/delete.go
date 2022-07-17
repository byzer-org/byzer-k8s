package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
	"strings"
)

func delete(c *cli.Context) error {

	k8sConfig := meta.BuildKubeConfig(c)
	engineName := c.String("engine-name")
	namespace := c.String("engine-namespace")

	metaConfig := meta.MetaConfig{
		K8sConfig:    k8sConfig,
		EngineConfig: meta.EngineConfig{Name: engineName, Namespace: namespace},
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

	// clean executor pods
	podsJson, _ := executor.GetInfo([]string{"pods", "-o", "json"})
	query := utils.BuildJsonQueryFromStr(podsJson)
	items, _ := query.Array("items")
	var podNames = make([]string, 0)
	for _, item := range items {
		v := item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		podNames = append(podNames, v)
	}

	for _, podName := range podNames {
		if strings.HasPrefix(podName, metaConfig.EngineConfig.Name) && strings.Contains(podName, "exec") {
			logger.Info(fmt.Sprintf("delete pod:%s", podName))
			executor.DeleteAny([]string{"pod", podName})
		}

	}

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
			&cli.StringFlag{
				Name:     "engine-namespace",
				Required: false,
				Usage:    "",
				Value:    "default",
			},
		},
	}

	return cmd
}
