package operator

import (
	"errors"
	"fmt"
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"mlsql.tech/allwefantasy/deploy/pkg/op_utils"
	"mlsql.tech/allwefantasy/deploy/pkg/tpl"
	"os"
)

type RoleOp struct {
	executor   *meta.KubeExecutor
	metaConfig meta.MetaConfig
}

func NewRole(executor *meta.KubeExecutor, config meta.MetaConfig) *RoleOp {
	v := RoleOp{executor: executor, metaConfig: config}
	return &v
}

func (v *RoleOp) Execute(verbose bool) error {

	// create serviceaccount
	logger.Info(fmt.Sprintf("Create service account :%s", v.metaConfig.EngineConfig.ServiceAccountName))
	createAccountTmpFile, _ := op_utils.TplEvt(tpl.TLPServiceAccount, v.metaConfig, verbose)
	defer os.Remove(createAccountTmpFile.Name())
	_, createAccountErr := v.executor.CreateDeployment([]string{"-f", createAccountTmpFile.Name(), "-o", "json"})
	if createAccountErr != nil {
		return errors.New(fmt.Sprintf("Fail to apply createRole.yaml \n %s", createAccountErr.Error()))
	}

	// create role
	logger.Info(fmt.Sprintf("Create role :%s with binding %s", v.metaConfig.EngineConfig.RoleName, v.metaConfig.EngineConfig.RoleBindingName))
	createRoleTmpFile, _ := op_utils.TplEvt(tpl.TLPCreateRole, v.metaConfig, verbose)
	defer os.Remove(createRoleTmpFile.Name())
	_, createRoleErr := v.executor.CreateDeployment([]string{"-f", createRoleTmpFile.Name(), "-o", "json"})
	if createRoleErr != nil {
		return errors.New(fmt.Sprintf("Fail to apply createRole.yaml \n %s", createRoleErr.Error()))
	}

	// create role binding
	bindRoleTmpFile, _ := op_utils.TplEvt(tpl.TLPRoleBinding, v.metaConfig, verbose)
	defer os.Remove(bindRoleTmpFile.Name())
	_, bindRoleErr := v.executor.CreateDeployment([]string{"-f", bindRoleTmpFile.Name(), "-o", "json"})
	if bindRoleErr != nil {
		return errors.New(fmt.Sprintf("Fail to apply roleBinding.yaml \n %s", bindRoleErr.Error()))
	}
	return nil
}
