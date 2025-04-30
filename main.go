package main

import (
	"fmt"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-kube-client/v3/pkg/kubeclient"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spf13/cobra"

	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/input/survey"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"k8s.io/client-go/tools/clientcmd"
)

type ContextOptions struct {
	options.BaseOptions

	Args   []string
	Filter string
	Shell  bool
}

var (
	version      string
	context_long = templates.LongDesc(`
		Displays or changes the current Kubernetes context (cluster).`)
	context_example = templates.Examples(`
		# to select the context to switch to
		jx context

		# view the current context
		jx context -b`)
)

func Main() *cobra.Command {
	o := ContextOptions{}
	cmd := &cobra.Command{
		Use:     "jx-context",
		Short:   "View or change the current Kubernetes context (Kubernetes cluster)",
		Long:    context_long,
		Example: context_example,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, _, filteredContextNames, err := o.filteredContextNames()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			var contextNames []string
			for _, v := range filteredContextNames {
				if strings.HasPrefix(v, toComplete) {
					contextNames = append(contextNames, v)
				}
			}
			return contextNames, cobra.ShellCompDirectiveNoFileComp
		},
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			o.Args = args
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.AddBaseFlags(cmd)
	cmd.Flags().StringVarP(&o.Filter, "filter", "f", "", "Filter the list of contexts to switch between using the given text")
	cmd.Flags().BoolVarP(&o.Shell, "shell", "s", false, "Start shell with chosen context")
	return cmd
}

func main() {
	if err := Main().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func (o *ContextOptions) Run() error {
	tmpKubeConfig := ""
	if o.Shell {
		// Copy kube config to temporary file
		if o.BatchMode {
			return fmt.Errorf("--batch mode is incompatible with --shell")
		}
		defaultPaths := clientcmd.NewDefaultPathOptions().GetLoadingPrecedence()
		data, err := os.ReadFile(defaultPaths[0])
		if err != nil {
			return err
		}
		tmpKubeConfig = filepath.Join(os.TempDir(), fmt.Sprintf("kube-config-%d", os.Getpid()))
		err = os.WriteFile(tmpKubeConfig, data, 0600)
		if err != nil {
			return err
		}
		err = os.Setenv("KUBECONFIG", tmpKubeConfig)
		if err != nil {
			return err
		}
	}

	config, po, contextNames, err := o.filteredContextNames()
	if err != nil {
		return err
	}

	ctxName := ""
	args := o.Args
	if len(args) > 0 {
		ctxName = args[0]
		if stringhelpers.StringArrayIndex(contextNames, ctxName) < 0 {
			return options.InvalidArg(ctxName, contextNames)
		}
	}

	if ctxName == "" && !o.BatchMode {
		defaultCtxName := config.CurrentContext
		pick, err := o.PickContext(contextNames, defaultCtxName)
		if err != nil {
			return err
		}
		ctxName = pick
	}
	info := termcolor.ColorInfo
	if ctxName != "" && ctxName != config.CurrentContext {
		ctx := config.Contexts[ctxName]
		if ctx == nil {
			return fmt.Errorf("could not find Kubernetes context %s", ctxName)
		}
		newConfig := *config
		newConfig.CurrentContext = ctxName
		err = clientcmd.ModifyConfig(po, newConfig, false)
		if err != nil {
			return fmt.Errorf("failed to update the kube config %s", err)
		}
		log.Logger().Infof("Now using namespace '%s' from context named '%s' on server '%s'.\n",
			info(ctx.Namespace), info(newConfig.CurrentContext), info(kube.Server(config, ctx)))
	} else {
		ns := kube.CurrentNamespace(config)
		server := kube.CurrentServer(config)
		log.Logger().Infof("Using namespace '%s' from context named '%s' on server '%s'.\n",
			info(ns), info(config.CurrentContext), info(server))
	}
	if o.Shell {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "bash"
		}

		cmd := &cmdrunner.Command{
			Name: shell,
			In:   os.Stdin,
			Out:  os.Stdout,
			Err:  os.Stderr,
		}
		_, err = cmdrunner.QuietCommandRunner(cmd)
		if err != nil {
			return err
		}
		return os.Remove(tmpKubeConfig)
	}
	return nil
}

func (o *ContextOptions) filteredContextNames() (*api.Config, *clientcmd.PathOptions, []string, error) {
	config, po, err := kubeclient.LoadConfig()
	if err != nil {
		return nil, nil, nil, err
	}

	if config == nil || config.Contexts == nil || len(config.Contexts) == 0 {
		return nil, nil, nil, fmt.Errorf("no Kubernetes contexts available! Try create or connect to cluster")
	}

	var contextNames []string
	for k, v := range config.Contexts {
		if k != "" && v != nil {
			if o.Filter == "" || strings.Contains(k, o.Filter) {
				contextNames = append(contextNames, k)
			}
		}
	}
	sort.Strings(contextNames)
	return config, po, contextNames, nil
}

func (o *ContextOptions) PickContext(names []string, defaultValue string) (string, error) {
	if len(names) == 0 {
		return "", nil
	}
	if len(names) == 1 {
		return names[0], nil
	}

	return survey.NewInput().PickNameWithDefault(names, "Change Kubernetes context:", defaultValue, "")
}
