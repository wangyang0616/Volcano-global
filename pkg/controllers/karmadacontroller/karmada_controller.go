/*
Copyright 2023 The Volcano Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package karmadacontroller

import (
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/karmada-io/karmada/cmd/controller-manager/app"
	"github.com/karmada-io/karmada/cmd/controller-manager/app/options"
	"github.com/karmada-io/karmada/pkg/sharedcli"
	"github.com/karmada-io/karmada/pkg/sharedcli/klogflag"
	"github.com/karmada-io/karmada/pkg/version/sharedcommand"
	"github.com/spf13/cobra"

	"k8s.io/component-base/cli"
	"k8s.io/component-base/term"
	"k8s.io/klog"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"volcano.sh/volcano/pkg/controllers/framework"
)

var kc *karmadacontroller

func init() {
	kc = &karmadacontroller{}
	if err := framework.RegisterController(kc); err != nil {
		klog.Errorf("failed to initial karmada controller, error: %s", err.Error())
	}
}

type karmadacontroller struct {
	karmadaOptions options.Options
	karmadaCmd     *cobra.Command
}

func (kc *karmadacontroller) Initialize(opt *framework.ControllerOption) error {
	opts := options.NewOptions()

	ctx := controllerruntime.SetupSignalHandler()
	cmd := &cobra.Command{
		Use: "karmada-controller-manager",
		Long: `The karmada-controller-manager runs various controllers.
The controllers watch Karmada objects and then talk to the underlying clusters' API servers 
to create regular Kubernetes resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// validate options
			if errs := opts.Validate(); len(errs) != 0 {
				return errs.ToAggregate()
			}

			return app.Run(ctx, opts)
		},
	}

	fss := cliflag.NamedFlagSets{}

	genericFlagSet := fss.FlagSet("generic")
	// Add the flag(--kubeconfig) that is added by controller-runtime
	// (https://github.com/kubernetes-sigs/controller-runtime/blob/v0.11.1/pkg/client/config/config.go#L39),
	// and update the flag usage.
	// genericFlagSet.AddGoFlagSet(flag.CommandLine)
	// genericFlagSet.Lookup("kubeconfig").Usage = "Path to karmada control plane kubeconfig file."
	// opts.AddFlags(genericFlagSet, controllers.ControllerNames(), controllersDisabledByDefault.List())

	// Set klog flags
	logsFlagSet := fss.FlagSet("logs")
	klogflag.Add(logsFlagSet)

	cmd.AddCommand(sharedcommand.NewCmdVersion("karmada-controller-manager"))
	cmd.Flags().AddFlagSet(genericFlagSet)
	cmd.Flags().AddFlagSet(logsFlagSet)

	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	sharedcli.SetUsageAndHelpFunc(cmd, fss, cols)
	kc.karmadaCmd = cmd
	return nil
}

func (kc *karmadacontroller) Name() string {
	return "karmada-controller"
}

func (kc *karmadacontroller) Run(stopCh <-chan struct{}) {
	cli.Run(kc.karmadaCmd)
}
