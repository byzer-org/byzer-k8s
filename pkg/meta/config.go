package meta

type K8sConfig struct {
	KubeConfig string
}
type EngineConfig struct {
	Name  string
	Image string

	ExecutorCoreNum int32
	ExecutorNum     int32
	ExecutorMemory  int32

	DriverCoreNum int32
	DriverMemory  int32

	AccessToken        string
	JarPathInContainer string
}

type StorageConfig struct {
	Name       string
	MetaUrl    string
	MountPoint string
	AccessKey  string
	SecretKey  string
}
