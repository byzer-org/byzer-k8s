package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"mlsql.tech/allwefantasy/deploy/pkg/tpl"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
	"os"
	"time"
)

func run(c *cli.Context) error {
	engineConfig := meta.EngineConfig{
		Name:               c.String("engine-name"),
		Image:              c.String("engine-image"),
		ExecutorCoreNum:    c.Int64("engine-executor-core-num"),
		ExecutorNum:        c.Int64("engine-executor-num"),
		ExecutorMemory:     c.Int64("engine-executor-memory"),
		DriverCoreNum:      c.Int64("engine-driver-core-num"),
		DriverMemory:       c.Int64("engine-driver-memory"),
		AccessToken:        c.String("engine-access-token"),
		JarPathInContainer: c.String("engine-jar-path-in-container"),
	}

	kubeConfigPath := c.String("kube-config")
	b, err := ioutil.ReadFile(kubeConfigPath)
	if err != nil {
		logger.Fatalf("load kube config file from %s: %s", kubeConfigPath, err)
		return nil
	}
	k8sConfig := meta.K8sConfig{KubeConfig: string(b)}

	var storageConfig meta.StorageConfig
	if c.IsSet("storage-name") {
		storageConfig = meta.StorageConfig{
			Name:       c.String("storage-name"),
			MetaUrl:    c.String("storage-meta-url"),
			MountPoint: c.String("storage-mount-point"),
			AccessKey:  c.String("storage-access-key"),
			SecretKey:  c.String("storage-secret-key"),
		}
	}

	metaConfig := meta.MetaConfig{
		K8sConfig:     k8sConfig,
		EngineConfig:  engineConfig,
		StorageConfig: storageConfig,
	}

	executor := utils.CreateKubeExecutor(&metaConfig.K8sConfig)

	tplEvt := func(templateStr string, data interface{}) (*os.File, error) {
		f, _ := utils.CreateTmpFile(tpl.EvaluateTemplate(templateStr, data))
		return f, nil
	}

	// Step1: configure CM, so we can support JuiceFS FileSystem
	executor.DeleteAny([]string{"configmap", "core-site-xml"})
	coreSiteTmpFile, _ := tplEvt(tpl.TLPCoreSite, metaConfig.StorageConfig)
	defer os.Remove(coreSiteTmpFile.Name())
	_, coreSiteTmpErr := executor.CreateCM([]string{"core-site-xml", "--from-file", "core-site.xml=" + coreSiteTmpFile.Name(), "-o", "json"})
	if coreSiteTmpErr != nil {
		logger.Fatalf("Fail to create core-site-xml in cm \n %s", coreSiteTmpErr.Error())
		return coreSiteTmpErr
	}

	// Step2: Add role and service count in k8s
	createRoleTmpFile, _ := tplEvt(tpl.TLPCreateRole, tpl.Empty{})
	defer os.Remove(createRoleTmpFile.Name())
	_, createRoleErr := executor.CreateDeployment([]string{"-f", createRoleTmpFile.Name(), "-o", "json"})
	if createRoleErr != nil {
		error := errors.New(fmt.Sprintf("Fail to apply createRole.yaml \n %s", createRoleErr.Error()))
		return error
	}

	bindRoleTmpFile, _ := tplEvt(tpl.TLPRoleBinding, tpl.Empty{})
	defer os.Remove(bindRoleTmpFile.Name())
	_, bindRoleErr := executor.CreateDeployment([]string{"-f", bindRoleTmpFile.Name(), "-o", "json"})
	if bindRoleErr != nil {
		error := errors.New(fmt.Sprintf("Fail to apply roleBinding.yaml \n %s", bindRoleErr.Error()))
		return error
	}

	// Step3: Deploy MLSQL Engine
	de := struct {
		*meta.EngineConfig
		K8sAddress         string
		LimitDriverCoreNum int64
		LimitDriverMemory  int64
	}{
		EngineConfig:       &metaConfig.EngineConfig,
		K8sAddress:         executor.GetK8sAddress(),
		LimitDriverCoreNum: metaConfig.EngineConfig.DriverCoreNum * 2,
		LimitDriverMemory:  metaConfig.EngineConfig.DriverMemory * 2,
	}

	deployTmpFile, _ := tplEvt(tpl.TLPDeployment, de)
	// defer os.Remove(deployTmpFile.Name())
	_, deployErr := executor.CreateDeployment([]string{"-f", deployTmpFile.Name(), "-o", "json"})
	if deployErr != nil {
		error := errors.New(fmt.Sprintf("Fail to apply deployment.yaml \n %s", deployErr.Error()))
		return error
	}

	// Step4: Expose MLSQL Engine service
	_, serviceErr := executor.CreateExpose([]string{"deployment", metaConfig.EngineConfig.Name, "--port", "9003",
		"--target-port", "9003", "--type", "LoadBalancer"})
	if serviceErr != nil {
		error := errors.New(fmt.Sprintf("Fail to expose service \n %s", serviceErr.Error()))
		return error
	}

	// Step5: Wait MLSQL Engine proxy service IP ready
	var ip, _ = executor.GetProxyIp()
	var counter int32 = 30
	for ip == "" && counter > 0 {
		time.Sleep(3 * time.Second)
		logger.Infof("Wait load balance ip ready...")
		counter -= 3
		ip, _ = executor.GetProxyIp()
	}

	logger.Infof("MLSQL Engine is ready: http://%s:%s", ip, "9003")
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
