package main

import (
	"fmt"
	"os"
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
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			o.Args = args
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.AddBaseFlags(cmd)
	cmd.Flags().StringVarP(&o.Filter, "filter", "f", "", "Filter the list of contexts to switch between using the given text")
	return cmd
}

func main() {
	if err := Main().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func (o *ContextOptions) Run() error {
	config, po, err := kubeclient.LoadConfig()
	if err != nil {
		return err
	}

	if config == nil || config.Contexts == nil || len(config.Contexts) == 0 {
		return fmt.Errorf("No Kubernetes contexts available! Try create or connect to cluster?")
	}

	contextNames := []string{}
	for k, v := range config.Contexts {
		if k != "" && v != nil {
			if o.Filter == "" || strings.Index(k, o.Filter) >= 0 {
				contextNames = append(contextNames, k)
			}
		}
	}
	sort.Strings(contextNames)

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
			return fmt.Errorf("Could not find Kubernetes context %s", ctxName)
		}
		newConfig := *config
		newConfig.CurrentContext = ctxName
		err = clientcmd.ModifyConfig(po, newConfig, false)
		if err != nil {
			return fmt.Errorf("Failed to update the kube config %s", err)
		}
		log.Logger().Infof("Now using namespace '%s' from context named '%s' on server '%s'.\n",
			info(ctx.Namespace), info(newConfig.CurrentContext), info(kube.Server(config, ctx)))
	} else {
		ns := kube.CurrentNamespace(config)
		server := kube.CurrentServer(config)
		log.Logger().Infof("Using namespace '%s' from context named '%s' on server '%s'.\n",
			info(ns), info(config.CurrentContext), info(server))
	}
	return nil
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
