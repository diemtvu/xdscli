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

	// If set, is router proxy (ingress/egress), otherwise, is sidecar
	proxyType string

	// Path to output file. Leave blank to output to stdout.
	outputFile string

	// short (default) or json
	outputFormat string
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", "~/.kube/config", "path to the kubeconfig file. Default is ~/.kube/config")
	RootCmd.PersistentFlags().StringVarP(&pilotURL, "pilot-url", "u", "", "pilot address. Will try port forward if not provided.")
	RootCmd.PersistentFlags().BoolVarP(&streaming, "watch", "w", false, "After listing/getting the requested object, watch for changes.")
	RootCmd.PersistentFlags().StringVarP(&proxyTag, "proxytag", "t", "", "Pod name or app label or istio label to identify the proxy.")
	RootCmd.PersistentFlags().StringVarP(&proxyType, "proxytype", "", "sidecar", "router or sidecar. Default sidecar")
	RootCmd.PersistentFlags().StringVarP(&outputFile, "file", "f", "", "output file. Leave blank to go to stdout")
	RootCmd.PersistentFlags().StringVarP(&outputFormat, "out", "o", "json", "output format. Accepted values: short, json (default)")
	
	RootCmd.AddCommand(lds())
	RootCmd.AddCommand(cds())
	RootCmd.AddCommand(eds())
	RootCmd.AddCommand(rds())
}

// RootCmd is the root command line.
var RootCmd = &cobra.Command{}
