package tpl

import (
	"bytes"
	_ "embed"
	"text/template"
)

var (
	//go:embed templates/core-site.xml
	TLPCoreSite string

	//go:embed templates/createRole.yaml
	TLPCreateRole string

	//go:embed templates/roleBinding.yaml
	TLPRoleBinding string

	//go:embed templates/deployment.yaml
	TLPDeployment string

	//go:embed templates/ingress.yaml
	TLPIngress string

	//go:embed templates/core-site.yaml
	TLPCoreSiteDeployment string

	//go:embed templates/service.yaml
	TLPService string

	//go:embed templates/service-account.yaml
	TLPServiceAccount string
)

func EvaluateTemplate(templateStr string, data interface{}) string {
	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		panic(err)
	}
	var tpl bytes.Buffer
	tmpl.Execute(&tpl, data)
	return tpl.String()
}

type Empty struct {
}
