package operator

import (
	"bytes"
	"fmt"
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"mlsql.tech/allwefantasy/deploy/pkg/op_utils"
	"mlsql.tech/allwefantasy/deploy/pkg/tpl"
	"os"
	"strings"
)

type ConfigMapOp struct {
	executor   *meta.KubeExecutor
	metaConfig meta.MetaConfig
	extraConf  map[string]string
}

func NewConfigMapOp(executor *meta.KubeExecutor, config meta.MetaConfig, extraConf map[string]string) *ConfigMapOp {
	v := ConfigMapOp{executor: executor, metaConfig: config, extraConf: extraConf}
	return &v
}

// Filters and converts key-value pair to a property of core-site.xml
func storageConfConverter(key, value string) string {
	var buf bytes.Buffer
	if strings.HasPrefix(key, "engine.storage") {
		buf.WriteString("<property>\n")
		buf.WriteString(fmt.Sprintf("<name>%s</name>\n", strings.TrimLeft(key, "engine.storage.")))
		buf.WriteString(fmt.Sprintf("<value>%s</value>\n", value))
		buf.WriteString("</property>")
	}
	return buf.String()
}

func (v *ConfigMapOp) Execute(verbose bool) error {
	v.executor.DeleteAny([]string{"configmap", "core-site-xml"})

	coreSiteTmpFile, _ := op_utils.TplEvt(tpl.TLPCoreSite,
		meta.StorageConfig{
			Name:        v.metaConfig.StorageConfig.Name,
			MetaUrl:     v.metaConfig.StorageConfig.MetaUrl,
			MountPoint:  v.metaConfig.StorageConfig.MountPoint,
			AccessKey:   v.metaConfig.StorageConfig.AccessKey,
			SecretKey:   v.metaConfig.StorageConfig.SecretKey,
			ExtraConfig: op_utils.ConvertToConfString(v.extraConf, storageConfConverter),
		},
		verbose)
	defer os.Remove(coreSiteTmpFile.Name())
	_, coreSiteTmpErr := v.executor.CreateCM([]string{"core-site-xml", "--from-file", "core-site.xml=" + coreSiteTmpFile.Name(), "-o", "json"})
	if coreSiteTmpErr != nil {
		logger.Fatalf("Fail to create core-site-xml in cm \n %s", coreSiteTmpErr.Error())
		return coreSiteTmpErr
	}
	return nil
}
