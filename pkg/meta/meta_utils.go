package meta

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/create"
	k8sdelete "k8s.io/kubectl/pkg/cmd/delete"
	"k8s.io/kubectl/pkg/cmd/expose"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/completion"
	"k8s.io/kubectl/pkg/util/i18n"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
	"os"
	"regexp"
	"strings"
)

var logger = utils.GetLogger("byzer-k8s-deploy/meta")

type KubeExecutor struct {
	kubeConfig  *K8sConfig
	KubeFactory cmdutil.Factory
}

func CreateKubeExecutor(kubeConfig *K8sConfig) *KubeExecutor {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	tmpfile, _ := utils.CreateTmpFile(kubeConfig.KubeConfig)
	tmpfileName := tmpfile.Name()
	kubeConfigFlags.KubeConfig = &tmpfileName
	kubeConfigFlags.Namespace = &kubeConfig.Namespace
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
	o := create.NewConfigMapOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "configmap NAME [--from-file=[key=]source] [--from-literal=key1=value1] [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"cm"},
		Short:                 i18n.T("Create a config map from a local file, directory or literal value"),
		Long:                  "",
		Example:               "",
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(executor.KubeFactory, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}
	o.PrintFlags.AddFlags(cmd)

	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)

	cmd.Flags().StringSliceVar(&o.FileSources, "from-file", o.FileSources, "Key file can be specified using its file path, in which case file basename will be used as configmap key, or optionally with a key and file path, in which case the given key will be used.  Specifying a directory will iterate each named file in the directory whose basename is a valid configmap key.")
	cmd.Flags().StringArrayVar(&o.LiteralSources, "from-literal", o.LiteralSources, "Specify a key and literal value to insert in configmap (i.e. mykey=somevalue)")
	cmd.Flags().StringSliceVar(&o.EnvFileSources, "from-env-file", o.EnvFileSources, "Specify the path to a file to read lines of key=val pairs to create a configmap.")
	cmd.Flags().BoolVar(&o.AppendHash, "append-hash", o.AppendHash, "Append a hash of the configmap to its name.")

	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl-create")

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

func parseK8sConfig(config string) (*api.Config, error) {
	if len(config) == 0 {
		return nil, errors.New("k8s config should not be empty")
	}
	configObj, error := clientcmd.Load([]byte(config))
	if error != nil {
		panic(error.Error())
	}
	return configObj, nil
}

// GetK8sAddress returns apiServer address,
// if multiple clusters exist, it takes one by random.
func (executor *KubeExecutor) GetK8sAddress() string {
	config, error := parseK8sConfig(executor.kubeConfig.KubeConfig)
	if error != nil {
		panic(error)
	}

	var apiServer string
	for _, v := range config.Clusters {
		apiServer = v.Server
		break
	}
	logger.Infof("apiServer %s\n", apiServer)
	if len(apiServer) == 0 {
		panic("Failed to read apiServer ")
	}
	return apiServer
}

func (executor *KubeExecutor) GetProxyIp() (string, error) {
	command := []string{"svc", "-o", "json"}
	infoJson, getError := executor.GetInfo(command)

	if getError != nil {
		error := errors.New(fmt.Sprintf("fail to get service \n %s", getError.Error()))
		return "", error
	}
	query := utils.BuildJsonQueryFromStr(infoJson)
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
	flags := apply.NewApplyFlags(executor.KubeFactory, ioStreams)

	cmd := &cobra.Command{
		Use:                   "apply (-f FILENAME | -k DIRECTORY)",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Apply a configuration to a resource by file name or stdin"),
		Long:                  "",
		Example:               "",
		Run: func(cmd *cobra.Command, args []string) {
			o, err := flags.ToOptions(cmd, "", args)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run())
		},
	}

	flags.AddFlags(cmd)

	// apply subcommands
	cmd.AddCommand(apply.NewCmdApplyViewLastApplied(flags.Factory, flags.IOStreams))
	cmd.AddCommand(apply.NewCmdApplySetLastApplied(flags.Factory, flags.IOStreams))
	cmd.AddCommand(apply.NewCmdApplyEditLastApplied(flags.Factory, flags.IOStreams))

	return cmd
}

func (executor *KubeExecutor) newExposeService(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := expose.NewExposeServiceOptions(ioStreams)
	exposeResources := i18n.T(`pod (po), service (svc), replicationcontroller (rc), deployment (deploy), replicaset (rs)`)
	validArgs := []string{}
	resources := regexp.MustCompile(`\s*,`).Split(exposeResources, -1)
	for _, r := range resources {
		validArgs = append(validArgs, strings.Fields(r)[0])
	}

	cmd := &cobra.Command{
		Use:                   "expose (-f FILENAME | TYPE NAME) [--port=port] [--protocol=TCP|UDP|SCTP] [--target-port=number-or-name] [--name=name] [--external-ip=external-ip-of-service] [--type=type]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Take a replication controller, service, deployment or pod and expose it as a new Kubernetes service"),
		Long:                  "",
		Example:               "",
		ValidArgsFunction:     completion.SpecifiedResourceTypeAndNameCompletionFunc(executor.KubeFactory, validArgs),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(executor.KubeFactory, cmd))
			cmdutil.CheckErr(o.RunExpose(cmd, args))
		},
	}

	o.RecordFlags.AddFlags(cmd)
	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Protocol, "protocol", o.Protocol, i18n.T("The network protocol for the service to be created. Default is 'TCP'."))
	cmd.Flags().StringVar(&o.Port, "port", o.Port, i18n.T("The port that the service should serve on. Copied from the resource being exposed, if unspecified"))
	cmd.Flags().StringVar(&o.Type, "type", o.Type, i18n.T("Type for this service: ClusterIP, NodePort, LoadBalancer, or ExternalName. Default is 'ClusterIP'."))
	cmd.Flags().StringVar(&o.LoadBalancerIP, "load-balancer-ip", o.LoadBalancerIP, i18n.T("IP to assign to the LoadBalancer. If empty, an ephemeral IP will be created and used (cloud-provider specific)."))
	cmd.Flags().StringVar(&o.Selector, "selector", o.Selector, i18n.T("A label selector to use for this service. Only equality-based selector requirements are supported. If empty (the default) infer the selector from the replication controller or replica set.)"))
	cmd.Flags().StringVarP(&o.Labels, "labels", "l", o.Labels, "Labels to apply to the service created by this call.")
	cmd.Flags().StringVar(&o.TargetPort, "target-port", o.TargetPort, i18n.T("Name or number for the port on the container that the service should direct traffic to. Optional."))
	cmd.Flags().StringVar(&o.ExternalIP, "external-ip", o.ExternalIP, i18n.T("Additional external IP address (not managed by Kubernetes) to accept for the service. If this IP is routed to a node, the service can be accessed by this IP in addition to its generated service IP."))
	cmd.Flags().StringVar(&o.Name, "name", o.Name, i18n.T("The name for the newly created object."))
	cmd.Flags().StringVar(&o.SessionAffinity, "session-affinity", o.SessionAffinity, i18n.T("If non-empty, set the session affinity for the service to this; legal values: 'None', 'ClientIP'"))
	cmd.Flags().StringVar(&o.ClusterIP, "cluster-ip", o.ClusterIP, i18n.T("ClusterIP to be assigned to the service. Leave empty to auto-allocate, or set to 'None' to create a headless service."))
	//cmdutil.AddFieldManagerFlagVar(cmd,v, "kubectl-expose")
	o.AddOverrideFlags(cmd)

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
		Short:                 i18n.T("Delete resources by file names, stdin, resources and names, or by resources and label selector"),
		Long:                  "",
		Example:               "",
		ValidArgsFunction:     completion.ResourceTypeAndNameCompletionFunc(executor.KubeFactory),
		Run: func(cmd *cobra.Command, args []string) {
			o, err := deleteFlags.ToOptions(nil, streams)
			cmdutil.CheckErr(err)
			cmdutil.CheckErr(o.Complete(executor.KubeFactory, args, cmd))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunDelete(executor.KubeFactory))
		},
		SuggestFor: []string{"rm"},
	}

	deleteFlags.AddFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)

	return cmd
}

func (executor *KubeExecutor) newGetCmd(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := get.NewGetOptions("", ioStreams)

	cmd := &cobra.Command{
		Use:                   fmt.Sprintf("get [(-o|--output=)%s] (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags]", strings.Join(o.PrintFlags.AllowedFormats(), "|")),
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Display one or many resources"),
		Long:                  "",
		Example:               "",
		// ValidArgsFunction is set when this function is called so that we have access to the util package
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(executor.KubeFactory, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd))
			cmdutil.CheckErr(o.Run(executor.KubeFactory, cmd, args))
		},
		SuggestFor: []string{"list", "ps"},
	}

	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Raw, "raw", o.Raw, "Raw URI to request from the server.  Uses the transport specified by the kubeconfig file.")
	cmd.Flags().BoolVarP(&o.Watch, "watch", "w", o.Watch, "After listing/getting the requested object, watch for changes.")
	cmd.Flags().BoolVar(&o.WatchOnly, "watch-only", o.WatchOnly, "Watch for changes to the requested object(s), without listing/getting first.")
	cmd.Flags().BoolVar(&o.OutputWatchEvents, "output-watch-events", o.OutputWatchEvents, "Output watch event objects when --watch or --watch-only is used. Existing objects are output as initial ADDED events.")
	cmd.Flags().BoolVar(&o.IgnoreNotFound, "ignore-not-found", o.IgnoreNotFound, "If the requested object does not exist the command will return exit code 0.")
	cmd.Flags().StringVar(&o.FieldSelector, "field-selector", o.FieldSelector, "Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.")
	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	//get.addOpenAPIPrintColumnFlags(cmd, o)
	//get.addServerPrintColumnFlags(cmd, o)
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, "identifying the resource to get from a server.")
	cmdutil.AddChunkSizeFlag(cmd, &o.ChunkSize)
	cmdutil.AddLabelSelectorFlagVar(cmd, &o.LabelSelector)
	var supportedSubresources = []string{"status", "scale"}
	cmdutil.AddSubresourceFlags(cmd, &o.Subresource, "If specified, gets the subresource of the requested object.", supportedSubresources...)
	return cmd
}
