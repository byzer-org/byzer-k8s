package tpl
import _ "embed"
var (
	//go:embed templates/core-site.xml
	TLPCoreSite string

	//go:embed templates/createRole.yaml
	TLPCreateRole string

	//go:embed templates/roleBinding.yaml
	TLPRoleBinding string

	//go:embed templates/deployment.yaml
	TLPDeployment string
)
