// Copyright 2020 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// kubeConfig is the path to the kubeconfig file
	kubeConfig string

	// pilotURL is the pilot/istiod URL.
	pilotURL string

	streaming bool

	// Pod name or app label or istio label to identify the proxy.
	proxyTag string

	// Either sidecar, ingress or router
	proxyType string

	// Path to output file. Leave blank to output to stdout.
	outputFile string
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", "~/.kube/config", "path to the kubeconfig file. Default is ~/.kube/config")
	RootCmd.PersistentFlags().StringVarP(&pilotURL, "pilot", "p", "", "pilot address. Will try port forward if not provided.")
	RootCmd.PersistentFlags().BoolVarP(&streaming, "streaming", "s", false, "If set, waiting on streaming gRPC until terminated.")
	RootCmd.PersistentFlags().StringVarP(&proxyTag, "proxytag", "t", "", "Pod name or app label or istio label to identify the proxy.")
	RootCmd.PersistentFlags().StringVarP(&proxyType, "proxytype", "", "sidecar", "sidecar, ingress, router. Default 'sidecar'.")
	RootCmd.PersistentFlags().StringVarP(&outputFile, "out", "o", "", "output file. Leave blank to go to stdout")

	RootCmd.AddCommand(lds())
	RootCmd.AddCommand(cds())
	RootCmd.AddCommand(eds())
	RootCmd.AddCommand(rds())
}

// RootCmd is the root command line.
var RootCmd = &cobra.Command{}
