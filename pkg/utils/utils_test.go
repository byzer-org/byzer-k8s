package utils

import (
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"os"
	"testing"
)

func TestReadConfigFile(t *testing.T) {
	content := `engine.spark.confkey=confvalue
engine.storage.confkey=confvalue`
	f, e := CreateTmpFile(content)
	if e != nil {
		t.Error(e)
	}
	defer os.Remove(f.Name())

	conf, err := ReadConfigFile(f.Name())

	if err != nil {
		t.Error(e)
	}
	if conf["engine.spark.confkey"] != "confvalue" {
		t.Error("Config file should contain engine.spark.confkey=confvalue")
	}

	if conf["engine.storage.confkey"] != "confvalue" {
		t.Error("Config file should contain engine.storage.confkey=confvalue")
	}

}

func TestReadK8sApiServerAddress(t *testing.T) {
	content := `
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/hadoop/.minikube/ca.crt
    extensions:
    - extension:
        last-update: Wed, 22 Sep 2021 10:09:26 CST
        provider: minikube.sigs.k8s.io
        version: v1.23.0
      name: cluster_info
    server: https://192.168.49.2:8443
  name: minikube
contexts:
- context:
    cluster: minikube
    extensions:
    - extension:
        last-update: Wed, 22 Sep 2021 10:09:26 CST
        provider: minikube.sigs.k8s.io
        version: v1.23.0
      name: context_info
    namespace: default
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
- name: minikube
  user:
    client-certificate: /home/hadoop/.minikube/profiles/minikube/client.crt
    client-key: /home/hadoop/.minikube/profiles/minikube/client.key
`

	kubeConfig := meta.K8sConfig{
		KubeConfig: content,
	}
	executor := CreateKubeExecutor(&kubeConfig)
	apiServer := executor.GetK8sAddress()
	if "https://192.168.49.2:8443" != apiServer {
		t.Error("Failed to parse config")
	}
}
