package op_utils

import (
	"bytes"
	"encoding/json"
	"mlsql.tech/allwefantasy/deploy/pkg/tpl"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
	"os"
)
var logger = utils.GetLogger("byzer-k8s-deploy")

func ConvertToConfString(extraConfMap map[string]string, converter func(key, value string) string) string {
	if extraConfMap == nil {
		return ""
	}
	var buff bytes.Buffer
	for key, value := range extraConfMap {
		buff.WriteString(converter(key, value))
	}
	return buff.String()
}

func TplEvt(templateStr string, data interface{},verbose bool) (*os.File, error) {
	if verbose == true {
		jsonObj, _ := json.Marshal(data)
		logger.Infof("%s\n", string(jsonObj))
	}
	f, _ := utils.CreateTmpFile(tpl.EvaluateTemplate(templateStr, data))
	return f, nil
}
