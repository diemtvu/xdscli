// Tool to get xDS configs from pilot. This tool simulate envoy proxy gRPC call. This can be used
// with either real pilot deployment (via port forwarding) or with a local pilot.
//
// Examples:
//
//
// To get LDS or CDS, use -type lds or -type cds, and provide the pod id or app label. For example:
// ```bash
// go run pilot_cli.go lds --proxytag httpbin-5766dd474b-2hlnx
// go run pilot_cli.go lds --proxytag httpbin
// ```
// Note If more than one pod match with the app label, one will be picked arbitrarily.
//
// To show only LDS with name 10.0.0.1_80
// ```bash
// go run pilot_cli.go lds --proxytag httpbin --name "10.0.0.1_80"
// ```
//
// For EDS/RDS, provide list of corresponding clusters or routes name. For example:
// ```bash
// go run pilot_cli.go eds --proxytag httpbin \
// --resources "inbound|http||sleep.default.svc.cluster.local" \
// --resources outbound|http||httpbin.default.svc.cluster.local"
// ```
//
// Script requires kube config in order to connect to k8s registry to get pod information (for LDS and CDS type). The default
// value for kubeconfig path is .kube/config in home folder (works for Linux only). It can be changed via -kubeconfig flag.
// ```bash
// go run pilot_cli.go --type lds --proxytag httpbin --kubeconfig path/to/kube/config
// ```
package main

import (
	"istio.io/pkg/log"
	cmd "xdscli/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Errorf("%v", err)
	}
}
