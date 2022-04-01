package tpl

import (
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"strings"
	"testing"
)

func TestEvaluateDeploymentTemplate(t *testing.T) {
	de := meta.DeploymentConfig{
		EngineConfig: &meta.EngineConfig{
			Name:               "TestEvaluateDeploymentTemplate",
			Image:              "techmlsql/mlsql-engine:3.0-2.1.0",
			ExecutorCoreNum:    1,
			ExecutorNum:        1,
			ExecutorMemory:     512,
			DriverCoreNum:      1,
			DriverMemory:       512,
			AccessToken:        "mlsql",
			JarPathInContainer: "local:///home/deploy/mlsql/libs/byzer-lang-3.1.1-2.12-2.1.0.jar",
		},
		K8sAddress:         "https://localhost:8443",
		LimitDriverCoreNum: 1,
		LimitDriverMemory:  1,
	}
	tplStr := EvaluateTemplate(TLPDeployment, de)

	if !strings.Contains(tplStr, "local:///home/deploy/mlsql/libs/byzer-lang-3.1.1-2.12-2.1.0.jar") {
		t.Errorf("Evaluated container args is invalid\n%s", tplStr)
	}
}
