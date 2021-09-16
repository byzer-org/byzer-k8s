package meta

type MetaConfig struct {
	K8sConfig     K8sConfig
	EngineConfig  EngineConfig
	StorageConfig StorageConfig
}

type K8sConfig struct {
	KubeConfig string
}
type EngineConfig struct {
	Name  string
	Image string

	ExecutorCoreNum int64
	ExecutorNum     int64
	ExecutorMemory  int64

	DriverCoreNum int64
	DriverMemory  int64

	AccessToken        string
	JarPathInContainer string
	ExtraSparkConfig   string
	ExtraMLSQLConfig   string
}

type StorageConfig struct {
	Name        string
	MetaUrl     string
	MountPoint  string
	AccessKey   string
	SecretKey   string
	ExtraConfig string
}
