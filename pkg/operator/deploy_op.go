package operator

import (
	"errors"
	"fmt"
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"mlsql.tech/allwefantasy/deploy/pkg/op_utils"
	"mlsql.tech/allwefantasy/deploy/pkg/tpl"
	"os"
	"strings"
)

type DeployOp struct {
	executor   *meta.KubeExecutor
	metaConfig meta.MetaConfig
	extraConf  map[string]string
}

func NewDeploy(executor *meta.KubeExecutor, config meta.MetaConfig, extraConf map[string]string) *DeployOp {
	v := DeployOp{executor: executor, metaConfig: config, extraConf: extraConf}
	return &v
}

func (v *DeployOp) Execute(verbose bool) error {
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

	v.metaConfig.EngineConfig.ExtraMLSQLConfig = op_utils.ConvertToConfString(v.extraConf, mlsqlConfConverter)
	v.metaConfig.EngineConfig.ExtraSparkConfig = op_utils.ConvertToConfString(v.extraConf, sparkConfConverter)

	de := meta.DeploymentConfig{
		EngineConfig:       &v.metaConfig.EngineConfig,
		K8sAddress:         v.executor.GetK8sAddress(),
		LimitDriverCoreNum: v.metaConfig.EngineConfig.DriverCoreNum * 2,
		LimitDriverMemory:  v.metaConfig.EngineConfig.DriverMemory * 2,
	}

	// Step1: Deploy byzer engine to K8s
	logger.Info("Deploy Byzer in K8s")
	deployTmpFile, _ := op_utils.TplEvt(tpl.TLPDeployment, de, verbose)
	defer os.Remove(deployTmpFile.Name())

	_, deployErr := v.executor.CreateDeployment([]string{"-f", deployTmpFile.Name(), "-o", "json"})

	// logger.Info(deployInfo)

	if deployErr != nil {
		return errors.New(fmt.Sprintf("Fail to apply deployment.yaml \n %s", deployErr.Error()))
	}

	// Step2: Expose Byzer Engine as Service
	logger.Info("Expose Byzer as service")
	serviceTmpFile, _ := op_utils.TplEvt(tpl.TLPService, de, verbose)
	defer os.Remove(serviceTmpFile.Name())
	_, serviceErr := v.executor.CreateDeployment([]string{"-f", serviceTmpFile.Name(), "-o", "json"})

	//_, serviceErr := v.executor.CreateExpose([]string{"deployment", v.metaConfig.EngineConfig.Name, "--port", "9003",
	//	"--target-port", "9003", "--type", "ClusterIP", "--namespace", v.metaConfig.EngineConfig.Namespace})

	// logger.Info(serviceInfo)

	if serviceErr != nil {
		return errors.New(fmt.Sprintf("Fail to expose service \n %s", serviceErr.Error()))

	}

	// step2: Create Ingress
	logger.Info("Expose Byzer as Ingress")
	ingressTmpFile, _ := op_utils.TplEvt(tpl.TLPIngress, de, verbose)
	defer os.Remove(ingressTmpFile.Name())
	_, ingressErr := v.executor.CreateDeployment([]string{"-f", ingressTmpFile.Name(), "-o", "json"})

	//logger.Info(ingressInfo)

	if ingressErr != nil {
		return errors.New(fmt.Sprintf("Fail to create ingress for %s: %s", de.Name, ingressErr.Error()))
	}

	//// Step4: Wait Byzer Engine proxy service IP ready
	//var ip, _ = v.executor.GetProxyIp()
	//var counter int32 = 30
	//for ip == "" && counter > 0 {
	//	time.Sleep(3 * time.Second)
	//	logger.Infof("Wait load balance ip ready...")
	//	counter -= 3
	//	ip, _ = v.executor.GetProxyIp()
	//}
	//
	//logger.Infof("Byzer Engine is ready: http://%s:%s", ip, "9003")
	return nil
}
