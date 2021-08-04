package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/jsonq"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/create"
	k8sdelete "k8s.io/kubectl/pkg/cmd/delete"
	"k8s.io/kubectl/pkg/cmd/expose"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/generate/versioned"
	"k8s.io/kubectl/pkg/util/i18n"
	"mlsql.tech/allwefantasy/deploy/pkg/meta"
	"os"
	"strings"
)

func CreateTmpFile(content string) (*os.File, error) {
	//fsys := os.DirFS(".")
	tmpfile, _ := os.CreateTemp(".", "*")
	tmpfileName := tmpfile.Name()
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return nil, fmt.Errorf("Fail to create tmp file [%s] ", tmpfileName)
	}
	return tmpfile, nil
}

func BuildJsonQueryFromStr(jsonStr string) *jsonq.JsonQuery {
	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(jsonStr))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)
	return jq

}

type KubeExecutor struct {
	kubeConfig  *meta.K8sConfig
	KubeFactory cmdutil.Factory
}

func CreateKubeExecutor(kubeConfig *meta.K8sConfig) *KubeExecutor {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	tmpfile, _ := CreateTmpFile(kubeConfig.KubeConfig)
	tmpfileName := tmpfile.Name()
	kubeConfigFlags.KubeConfig = &tmpfileName
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	kubeExecutor := &KubeExecutor{
		kubeConfig:  kubeConfig,
		KubeFactory: f,
	}
	return kubeExecutor
}

func (executor *KubeExecutor) setupCommand(createCommand func(...interface{}) *cobra.Command) (string, error) {
	config := executor.KubeFactory.ToRawKubeConfigLoader()
	defer os.Remove(config.ConfigAccess().GetExplicitFile())

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	var b2 bytes.Buffer
	errorWriter := bufio.NewWriter(&b2)

	in, out, err := strings.NewReader(""), writer, errorWriter
	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}
	c := createCommand("kubectl", executor.KubeFactory, ioStreams)
	c.Execute()
	out.Flush()

	errorWriter.Flush()
	errorStr := string(b2.Bytes())
	if len(errorStr) > 0 {
		return "", fmt.Errorf(errorStr)
	}
	return string(b.Bytes()), nil
}

func (executor *KubeExecutor) GetInfo(command []string) (string, error) {

	create := func(objs ...interface{}) *cobra.Command {
		c := executor.newGetCmd(objs[2].(genericclioptions.IOStreams))
		c.SetArgs(command)
		return c
	}
	return executor.setupCommand(create)
}

func (executor *KubeExecutor) newCreateCM(ioStreams genericclioptions.IOStreams) *cobra.Command {
	options := &create.ConfigMapOpts{
		CreateSubcommandOptions: create.NewCreateSubcommandOptions(ioStreams),
	}

	cmd := &cobra.Command{
		Use:                   "configmap NAME [--from-file=[key=]source] [--from-literal=key1=value1] [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"cm"},
		Short:                 i18n.T("Create a configmap from a local file, directory or literal value"),
		Long:                  "configMapLong",
		Example:               "configMapExample",
		Run: func(cmd *cobra.Command, args []string) {
			options.Complete(executor.KubeFactory, cmd, args)
			err2 := options.Run()

			if err2 != nil {
				cmd.PrintErr(err2.Error())
			}
		},
	}

	options.CreateSubcommandOptions.PrintFlags.AddFlags(cmd)
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, versioned.ConfigMapV1GeneratorName)
	cmd.Flags().StringSlice("from-file", []string{}, "Key file can be specified using its file path, in which case file basename will be used as configmap key, or optionally with a key and file path, in which case the given key will be used.  Specifying a directory will iterate each named file in the directory whose basename is a valid configmap key.")
	cmd.Flags().StringArray("from-literal", []string{}, "Specify a key and literal value to insert in configmap (i.e. mykey=somevalue)")
	cmd.Flags().String("from-env-file", "", "Specify the path to a file to read lines of key=val pairs to create a configmap (i.e. a Docker .env file).")
	cmd.Flags().Bool("append-hash", false, "Append a hash of the configmap to its name.")
	cmdutil.AddFieldManagerFlagVar(cmd, &options.CreateSubcommandOptions.FieldManager, "kubectl-create")
	return cmd
}

func (executor *KubeExecutor) CreateDeployment(command []string) (string, error) {

	create := func(objs ...interface{}) *cobra.Command {
		c := executor.newCmdApply(objs[2].(genericclioptions.IOStreams))
		c.SetArgs(command)
		return c
	}
	return executor.setupCommand(create)
}

func (executor *KubeExecutor) GetProxyIp() (string, error) {
	command := []string{"svc", "-o", "json"}
	infoJson, getError := executor.GetInfo(command)

	if getError != nil {
		error := errors.New(fmt.Sprintf("fail to get service \n %s", getError.Error()))
		return "", error
	}
	query := BuildJsonQueryFromStr(infoJson)
	ip, ipError := query.String("items", "1", "status", "loadBalancer", "ingress", "0", "ip")
	if ipError != nil {
		error := errors.New(fmt.Sprintf("fail to get ip from service \n %s", ipError.Error()))
		return "", error
	}
	return ip, nil
}

func (executor *KubeExecutor) CreateExpose(command []string) (string, error) {

	create := func(objs ...interface{}) *cobra.Command {
		c := executor.newExposeService(objs[2].(genericclioptions.IOStreams))
		c.SetArgs(command)
		return c
	}
	return executor.setupCommand(create)
}

func (executor *KubeExecutor) DeleteAny(command []string) (string, error) {
	var recover = func() { recover() }
	defer recover()

	create := func(objs ...interface{}) *cobra.Command {
		c := executor.newCmdDelete(objs[2].(genericclioptions.IOStreams))
		c.SetArgs(command)
		return c
	}
	return executor.setupCommand(create)
}

func (executor *KubeExecutor) CreateCM(command []string) (string, error) {
	create := func(objs ...interface{}) *cobra.Command {
		c := executor.newCreateCM(objs[2].(genericclioptions.IOStreams))
		c.SetArgs(command)
		return c
	}
	return executor.setupCommand(create)
}

func (executor *KubeExecutor) newCmdApply(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := apply.NewApplyOptions(ioStreams)

	// Store baseName for use in printing warnings / messages involving the base command name.
	// This is useful for downstream command that wrap this one.
	// o.cmdBaseName = baseName

	cmd := &cobra.Command{
		Use:                   "apply (-f FILENAME | -k DIRECTORY)",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Apply a configuration to a resource by filename or stdin"),
		Long:                  "",
		Example:               "",
		Run: func(cmd *cobra.Command, args []string) {
			o.Complete(executor.KubeFactory, cmd)
			// cmdutil.CheckErr(apply.validateArgs(cmd, args))
			// cmdutil.CheckErr(apply.validatePruneAll(o.Prune, o.All, o.Selector))
			err2 := o.Run()

			if err2 != nil {
				cmd.PrintErr(err2.Error())
			}
		},
	}
	cmd.SetIn(ioStreams.In)
	cmd.SetOut(ioStreams.Out)
	cmd.SetErr(ioStreams.ErrOut)

	// bind flag structs
	o.DeleteFlags.AddFlags(cmd)
	o.RecordFlags.AddFlags(cmd)
	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().BoolVar(&o.Overwrite, "overwrite", o.Overwrite, "Automatically resolve conflicts between the modified and live configuration by using values from the modified configuration")
	cmd.Flags().BoolVar(&o.Prune, "prune", o.Prune, "Automatically delete resource objects, including the uninitialized ones, that do not appear in the configs and are created by either apply or create --save-config. Should be used with either -l or --all.")
	cmdutil.AddValidateFlags(cmd)
	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&o.All, "all", o.All, "Select all resources in the namespace of the specified resource types.")
	cmd.Flags().StringArrayVar(&o.PruneWhitelist, "prune-whitelist", o.PruneWhitelist, "Overwrite the default whitelist with <group/version/kind> for --prune")
	cmd.Flags().BoolVar(&o.OpenAPIPatch, "openapi-patch", o.OpenAPIPatch, "If true, use openapi to calculate diff when the openapi presents and the resource can be found in the openapi spec. Otherwise, fall back to use baked-in types.")
	cmdutil.AddDryRunFlag(cmd)
	cmdutil.AddServerSideApplyFlags(cmd)
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, apply.FieldManagerClientSideApply)

	return cmd
}

func (executor *KubeExecutor) newExposeService(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := expose.NewExposeServiceOptions(ioStreams)

	validArgs := []string{}

	cmd := &cobra.Command{
		Use:                   "expose (-f FILENAME | TYPE NAME) [--port=port] [--protocol=TCP|UDP|SCTP] [--target-port=number-or-name] [--name=name] [--external-ip=external-ip-of-service] [--type=type]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Take a replication controller, service, deployment or pod and expose it as a new Kubernetes Service"),
		Long:                  "",
		Example:               "",
		Run: func(cmd *cobra.Command, args []string) {
			o.Complete(executor.KubeFactory, cmd)
			err2 := o.RunExpose(cmd, args)
			if err2 != nil {
				cmd.PrintErr(err2.Error())
			}
		},
		ValidArgs: validArgs,
	}

	o.RecordFlags.AddFlags(cmd)
	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().String("generator", "service/v2", i18n.T("The name of the API generator to use. There are 2 generators: 'service/v1' and 'service/v2'. The only difference between them is that service port in v1 is named 'default', while it is left unnamed in v2. Default is 'service/v2'."))
	cmd.Flags().String("protocol", "", i18n.T("The network protocol for the service to be created. Default is 'TCP'."))
	cmd.Flags().String("port", "", i18n.T("The port that the service should serve on. Copied from the resource being exposed, if unspecified"))
	cmd.Flags().String("type", "", i18n.T("Type for this service: ClusterIP, NodePort, LoadBalancer, or ExternalName. Default is 'ClusterIP'."))
	cmd.Flags().String("load-balancer-ip", "", i18n.T("IP to assign to the LoadBalancer. If empty, an ephemeral IP will be created and used (cloud-provider specific)."))
	cmd.Flags().String("selector", "", i18n.T("A label selector to use for this service. Only equality-based selector requirements are supported. If empty (the default) infer the selector from the replication controller or replica set.)"))
	cmd.Flags().StringP("labels", "l", "", "Labels to apply to the service created by this call.")
	cmd.Flags().String("container-port", "", i18n.T("Synonym for --target-port"))
	cmd.Flags().MarkDeprecated("container-port", "--container-port will be removed in the future, please use --target-port instead")
	cmd.Flags().String("target-port", "", i18n.T("Name or number for the port on the container that the service should direct traffic to. Optional."))
	cmd.Flags().String("external-ip", "", i18n.T("Additional external IP address (not managed by Kubernetes) to accept for the service. If this IP is routed to a node, the service can be accessed by this IP in addition to its generated service IP."))
	cmd.Flags().String("overrides", "", i18n.T("An inline JSON override for the generated object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field."))
	cmd.Flags().String("name", "", i18n.T("The name for the newly created object."))
	cmd.Flags().String("session-affinity", "", i18n.T("If non-empty, set the session affinity for the service to this; legal values: 'None', 'ClientIP'"))
	cmd.Flags().String("cluster-ip", "", i18n.T("ClusterIP to be assigned to the service. Leave empty to auto-allocate, or set to 'None' to create a headless service."))
	//cmdutil.AddFieldManagerFlagVar(cmd, &o.fieldManager, "kubectl-expose")
	usage := "identifying the resource to expose a service"
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	cmdutil.AddDryRunFlag(cmd)
	cmdutil.AddApplyAnnotationFlags(cmd)
	return cmd
}

func (executor *KubeExecutor) newCmdDelete(streams genericclioptions.IOStreams) *cobra.Command {
	deleteFlags := k8sdelete.NewDeleteCommandFlags("containing the resource to delete.")

	cmd := &cobra.Command{
		Use:                   "delete ([-f FILENAME] | [-k DIRECTORY] | TYPE [(NAME | -l label | --all)])",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Delete resources by filenames, stdin, resources and names, or by resources and label selector"),
		Long:                  "",
		Example:               "",
		Run: func(cmd *cobra.Command, args []string) {
			o := deleteFlags.ToOptions(nil, streams)
			o.Complete(executor.KubeFactory, args, cmd)
			err2 := o.RunDelete(executor.KubeFactory)
			if err2 != nil {
				cmd.PrintErr(err2.Error())
			}
		},
		SuggestFor: []string{"rm"},
	}

	deleteFlags.AddFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)

	return cmd
}

func (executor *KubeExecutor) newGetCmd(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := get.NewGetOptions("kubectl", ioStreams)

	cmd := &cobra.Command{
		Use:                   "get [(-o|--output=)json|yaml|wide|custom-columns=...|custom-columns-file=...|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=...] (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Display one or many resources"),
		Long:                  "",
		Example:               "",
		Run: func(cmd *cobra.Command, args []string) {
			o.Complete(executor.KubeFactory, cmd, args)
			err2 := o.Run(executor.KubeFactory, cmd, args)
			if err2 != nil {
				cmd.PrintErr(err2.Error())
			}
		},
		SuggestFor: []string{"list", "ps"},
	}

	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Raw, "raw", o.Raw, "Raw URI to request from the server.  Uses the transport specified by the kubeconfig file.")
	cmd.Flags().BoolVarP(&o.Watch, "watch", "w", o.Watch, "After listing/getting the requested object, watch for changes. Uninitialized objects are excluded if no object name is provided.")
	cmd.Flags().BoolVar(&o.WatchOnly, "watch-only", o.WatchOnly, "Watch for changes to the requested object(s), without listing/getting first.")
	cmd.Flags().BoolVar(&o.OutputWatchEvents, "output-watch-events", o.OutputWatchEvents, "Output watch event objects when --watch or --watch-only is used. Existing objects are output as initial ADDED events.")
	cmd.Flags().Int64Var(&o.ChunkSize, "chunk-size", o.ChunkSize, "Return large lists in chunks rather than all at once. Pass 0 to disable. This flag is beta and may change in the future.")
	cmd.Flags().BoolVar(&o.IgnoreNotFound, "ignore-not-found", o.IgnoreNotFound, "If the requested object does not exist the command will return exit code 0.")
	cmd.Flags().StringVarP(&o.LabelSelector, "selector", "l", o.LabelSelector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().StringVar(&o.FieldSelector, "field-selector", o.FieldSelector, "Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.")
	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, "identifying the resource to get from a server.")
	return cmd
}
