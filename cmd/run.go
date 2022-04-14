package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"mlsql.tech/allwefantasy/deploy/pkg/tpl"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
	"os"
	"strings"
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

	// Read config from file
	var extraConf map[string]string
	var parseConfErr error
	if c.IsSet("engine-config") {
		confFile := c.String("engine-config")
		extraConf, parseConfErr = utils.ReadConfigFile(confFile)
		if parseConfErr != nil {
			logger.Fatalf("Failed to read %s, error %s", confFile, err)
			return err
		}
	}

	tplEvt := func(templateStr string, data interface{}) (*os.File, error) {
		if c.IsSet("verbose") {
			jsonObj, _ := json.Marshal(data)
			logger.Infof("%s\n", string(jsonObj))
		}
		f, _ := utils.CreateTmpFile(tpl.EvaluateTemplate(templateStr, data))
		return f, nil
	}

	// The function converts each key value to a string,
	// conversion logic is defined in converter function
	convertToConfString := func(extraConfMap map[string]string,
		converter func(key, value string) string) string {

		if extraConfMap == nil {
			return ""
		}
		var buff bytes.Buffer
		for key, value := range extraConfMap {
			buff.WriteString(converter(key, value))
		}
		return buff.String()
	}

	// Filters and converts key-value pair to a property of core-site.xml
	storageConfConverter := func(key, value string) string {
		var buf bytes.Buffer
		if strings.HasPrefix(key, "engine.storage") {
			buf.WriteString("<property>\n")
			buf.WriteString(fmt.Sprintf("<name>%s</name>\n", strings.TrimLeft(key, "engine.storage.")))
			buf.WriteString(fmt.Sprintf("<value>%s</value>\n", value))
			buf.WriteString("</property>")
		}
		return buf.String()
	}

	executor := utils.CreateKubeExecutor(&metaConfig.K8sConfig)
	// Step1: configure CM, so we can support JuiceFS FileSystem
	executor.DeleteAny([]string{"configmap", "core-site-xml"})

	coreSiteTmpFile, _ := tplEvt(tpl.TLPCoreSite,
		meta.StorageConfig{
			Name:        metaConfig.StorageConfig.Name,
			MetaUrl:     metaConfig.StorageConfig.MetaUrl,
			MountPoint:  metaConfig.StorageConfig.MountPoint,
			AccessKey:   metaConfig.StorageConfig.AccessKey,
			SecretKey:   metaConfig.StorageConfig.SecretKey,
			ExtraConfig: convertToConfString(extraConf, storageConfConverter),
		},
	)
	defer os.Remove(coreSiteTmpFile.Name())
	_, coreSiteTmpErr := executor.CreateCM([]string{"core-site-xml", "--from-file", "core-site.xml=" + coreSiteTmpFile.Name(), "-o", "json"})
	if coreSiteTmpErr != nil {
		logger.Fatalf("Fail to create core-site-xml in cm \n %s", coreSiteTmpErr.Error())
		return coreSiteTmpErr
	}

	// Step2: Add role and service account in k8s
	createRoleTmpFile, _ := tplEvt(tpl.TLPCreateRole, tpl.Empty{})
	defer os.Remove(createRoleTmpFile.Name())
	_, createRoleErr := executor.CreateDeployment([]string{"-f", createRoleTmpFile.Name(), "-o", "json"})
	if createRoleErr != nil {
		return errors.New(fmt.Sprintf("Fail to apply createRole.yaml \n %s", createRoleErr.Error()))
	}

	bindRoleTmpFile, _ := tplEvt(tpl.TLPRoleBinding, tpl.Empty{})
	defer os.Remove(bindRoleTmpFile.Name())
	_, bindRoleErr := executor.CreateDeployment([]string{"-f", bindRoleTmpFile.Name(), "-o", "json"})
	if bindRoleErr != nil {
		return errors.New(fmt.Sprintf("Fail to apply roleBinding.yaml \n %s", bindRoleErr.Error()))
	}

	// Filters and converts extra spark conf
	sparkConfConverter := func(key, value string) string {
		if strings.HasPrefix(key, "engine.spark") {
			return fmt.Sprintf(" --conf \\\"%s=%s\\\"", strings.TrimPrefix(key, "engine."), value)
		} else {
			return " "
		}
	}
	// Filters and converts extra byzer engine conf.
	mlsqlConfConverter := func(key, value string) string {
		if strings.HasPrefix(key, "engine.streaming") {
			return fmt.Sprintf("\\\" -%s\\\" \\\"%s\\\" ", strings.TrimPrefix(key, "engine."), value)
		} else {
			return " "
		}
	}
	// Step3: Deploy Byzer Engine
	de := meta.DeploymentConfig{
		EngineConfig: &meta.EngineConfig{
			Name:               metaConfig.EngineConfig.Name,
			Image:              metaConfig.EngineConfig.Image,
			ExecutorCoreNum:    metaConfig.EngineConfig.ExecutorCoreNum,
			ExecutorNum:        metaConfig.EngineConfig.ExecutorNum,
			ExecutorMemory:     metaConfig.EngineConfig.ExecutorMemory,
			DriverCoreNum:      metaConfig.EngineConfig.DriverCoreNum,
			DriverMemory:       metaConfig.EngineConfig.DriverMemory,
			AccessToken:        metaConfig.EngineConfig.AccessToken,
			JarPathInContainer: metaConfig.EngineConfig.JarPathInContainer,
			ExtraSparkConfig:   convertToConfString(extraConf, sparkConfConverter),
			ExtraMLSQLConfig:   convertToConfString(extraConf, mlsqlConfConverter),
		},
		K8sAddress:         executor.GetK8sAddress(),
		LimitDriverCoreNum: metaConfig.EngineConfig.DriverCoreNum * 2,
		LimitDriverMemory:  metaConfig.EngineConfig.DriverMemory * 2,
	}

	deployTmpFile, _ := tplEvt(tpl.TLPDeployment, de)
	defer os.Remove(deployTmpFile.Name())

	_, deployErr := executor.CreateDeployment([]string{"-f", deployTmpFile.Name(), "-o", "json"})
	if deployErr != nil {
		return errors.New(fmt.Sprintf("Fail to apply deployment.yaml \n %s", deployErr.Error()))
	}

	// Step4: Expose Byzer Engine service
	_, serviceErr := executor.CreateExpose([]string{"deployment", metaConfig.EngineConfig.Name, "--port", "9003",
		"--target-port", "9003", "--type", "NodePort"})
	if serviceErr != nil {
		return errors.New(fmt.Sprintf("Fail to expose service \n %s", serviceErr.Error()))

	}

	// step5: Create Ingress
	ingressTmpFile, _ := tplEvt(tpl.TLPIngress, de)
	defer os.Remove(ingressTmpFile.Name())
	_, ingressErr := executor.CreateDeployment([]string{"-f", ingressTmpFile.Name(), "-o", "json"})
	if ingressErr != nil {
		return errors.New(fmt.Sprintf("Fail to create ingress for %s: %s", de.Name, ingressErr.Error()))
	}

	// Step6: Wait Byzer Engine proxy service IP ready
	var ip, _ = executor.GetProxyIp()
	var counter int32 = 30
	for ip == "" && counter > 0 {
		time.Sleep(3 * time.Second)
		logger.Infof("Wait load balance ip ready...")
		counter -= 3
		ip, _ = executor.GetProxyIp()
	}

	logger.Infof("Byzer Engine is ready: http://%s:%s", ip, "9003")
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
			Name:     "engine-image",
			Required: true,
			Usage:    "The name of Byzer Engine",
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
			Value: "",
			Usage: "the access token to protect byzer engine",
		},
		&cli.StringFlag{
			Name:     "engine-jar-path-in-container",
			Value:    "",
			Usage:    "The path of byzer engine jar in docker image",
			Required: true,
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
