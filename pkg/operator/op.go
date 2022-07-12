package operator

import "mlsql.tech/allwefantasy/deploy/pkg/utils"

var logger = utils.GetLogger("byzer-k8s-deploy/op")

type BaseOp interface {
	Execute(verbose bool) error
}
