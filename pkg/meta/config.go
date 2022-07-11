package meta

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
)

type MetaConfig struct {
	K8sConfig     K8sConfig
	EngineConfig  EngineConfig
	StorageConfig StorageConfig
}

type K8sConfig struct {
	KubeConfig string
	Namespace  string
}
type EngineConfig struct {
	Name          string
	Image         string
	EngineVersion string

	ExecutorCoreNum float64
	ExecutorNum     int64
	ExecutorMemory  int64

	DriverCoreNum float64
	DriverMemory  int64

	Namespace          string
	ServiceAccountName string
	RoleName           string
	RoleBindingName    string

	AccessToken        string
	JarPathInContainer string
	ExtraSparkConfig   string
	ExtraMLSQLConfig   string
}

func BuildKubeConfig(c *cli.Context) K8sConfig {
	kubeConfigPath := c.String("kube-config")
	b, err := ioutil.ReadFile(kubeConfigPath)
	if err != nil {
		panic(fmt.Sprintf("load kube config file from %s: %s", kubeConfigPath, err))
	}
	k8sConfig := K8sConfig{KubeConfig: string(b), Namespace: c.String("engine-namespace")}
	return k8sConfig
}

func BuildExtraEngineConfig(c *cli.Context) (map[string]string, error) {
	var extraConf map[string]string
	var parseConfErr error
	if c.IsSet("engine-config") {
		confFile := c.String("engine-config")
		extraConf, parseConfErr = utils.ReadConfigFile(confFile)
		if parseConfErr != nil {
			return extraConf, parseConfErr
		}
	}
	return extraConf, nil
}

func BuildEngineConfig(c *cli.Context) EngineConfig {
	engineVersion := c.String("engine-version")
	var engineImage = c.String("engine-image")
	var jarPathInContainer = c.String("engine-jar-path-in-container")

	if engineVersion != "" {
		if engineImage == "" {
			engineImage = fmt.Sprintf("byzer/byzer-lang-k8s-base:3.1.1-%s", engineVersion)
		}
		if jarPathInContainer == "" {
			jarPathInContainer = fmt.Sprintf("local:///home/deploy/byzer-lang/main/byzer-lang-3.1.1-2.12-%s.jar", engineVersion)
		}
	}

	if engineImage == "" || jarPathInContainer == "" {
		panic("egnine-version or [engine-image,engine-jar-path-in-container] should be specified")
	}

	engineConfig := EngineConfig{
		Name:          c.String("engine-name"),
		EngineVersion: engineVersion,
		Image:         engineImage,

		ServiceAccountName: c.String("engine-service-account-name"),
		Namespace:          c.String("engine-namespace"),
		RoleName:           c.String("engine-role-name"),
		RoleBindingName:    c.String("engine-role-binding-name"),

		ExecutorCoreNum:    c.Float64("engine-executor-core-num"),
		ExecutorNum:        c.Int64("engine-executor-num"),
		ExecutorMemory:     c.Int64("engine-executor-memory"),
		DriverCoreNum:      c.Float64("engine-driver-core-num"),
		DriverMemory:       c.Int64("engine-driver-memory"),
		AccessToken:        c.String("engine-access-token"),
		JarPathInContainer: jarPathInContainer,
	}
	return engineConfig
}

type StorageConfig struct {
	Name        string
	MetaUrl     string
	MountPoint  string
	AccessKey   string
	SecretKey   string
	ExtraConfig string
}

func BuildStorageConfig(c *cli.Context) StorageConfig {
	var storageConfig StorageConfig
	if c.IsSet("storage-name") {
		storageConfig = StorageConfig{
			Name:       c.String("storage-name"),
			MetaUrl:    c.String("storage-meta-url"),
			MountPoint: c.String("storage-mount-point"),
			AccessKey:  c.String("storage-access-key"),
			SecretKey:  c.String("storage-secret-key"),
		}
	}
	return storageConfig
}

type DeploymentConfig struct {
	*EngineConfig
	K8sAddress         string
	LimitDriverCoreNum float64
	LimitDriverMemory  int64
}
