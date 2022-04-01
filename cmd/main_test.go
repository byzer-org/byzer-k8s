package main

import (
	"fmt"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
	"os"
	"testing"
)

// Please start k8s before running this test
func TestConfFile(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	content := `engine.spark.kubernetes.container.image.pullPolicy=IfNotPresent
engine.spark.local.dir=/work/shuffle-data
engine.spark.kubernetes.driver.volumes.hostPath.spark-local-dir-shuffle-disk-drv.mount.path=/work/shuffle-data
engine.spark.kubernetes.driver.volumes.hostPath.spark-local-dir-shuffle-disk-drv.mount.readOnly=false
engine.spark.kubernetes.driver.volumes.hostPath.spark-local-dir-shuffle-disk-drv.options.path=/work/mnt/spark-shuffle-drv
engine.spark.kubernetes.executor.volumes.hostPath.spark-local-dir-shuffle-disk.mount.path=/work/shuffle-data
engine.spark.kubernetes.executor.volumes.hostPath.spark-local-dir-shuffle-disk.mount.readOnly=false
engine.spark.kubernetes.executor.volumes.hostPath.spark-local-dir-shuffle-disk.options.path=/work/mnt/spark-shuffle-exec
engine.streaming.datalake.path=./data/`
	f, e := utils.CreateTmpFile(content)
	if e != nil {
		logger.Error(e)
	}
	defer os.Remove(f.Name())

	args := []string{
		"--verbose",
		"run",
		"--kube-config", fmt.Sprintf("%s/.kube/config", os.Getenv("HOME")),
		"--engine-executor-core-num", "1",
		"--engine-executor-num", "1",
		"--engine-executor-memory", "512",
		"--engine-driver-core-num", "1",
		"--engine-driver-memory", "512",
		"--engine-name", "mlsql-engine",
		"--engine-image", "techmlsql/mlsql-engine:3.0-2.1.0",
		"--engine-jar-path-in-container", "local:///home/deploy/mlsql/libs/byzer-lang-3.1.1-2.12-2.1.0.jar",
		"--engine-config", f.Name(),
	}
	os.Args = append([]string{"mlsql-deploy"}, args...)
	main()
}

// Please start k8s before running this test
func TestNonConfFile(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	args := []string{
		"--verbose",
		"run",
		"--kube-config", fmt.Sprintf("%s/.kube/config", os.Getenv("HOME")),
		"--engine-executor-core-num", "1",
		"--engine-executor-num", "1",
		"--engine-executor-memory", "512",
		"--engine-driver-core-num", "1",
		"--engine-driver-memory", "512",
		"--engine-name", "mlsql-engine",
		"--engine-image", "techmlsql/mlsql-engine:3.0-2.1.0",
		"--engine-jar-path-in-container", "local:///home/deploy/mlsql/libs/byzer-lang-3.1.1-2.12-2.1.0.jar",
	}
	os.Args = append([]string{"mlsql-deploy"}, args...)
	main()
}
