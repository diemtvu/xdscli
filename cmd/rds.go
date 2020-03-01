package cmd

import (
	"github.com/spf13/cobra"

	xdsapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func rds() *cobra.Command {
	handler := &rdsHandler{}
	localCmd := makeXDSCmd("rds", handler)
	localCmd.Flags().StringArrayVarP(&handler.resources, "resources", "r", nil, "Resources to show")
	return localCmd
}

type rdsHandler struct {
	resources []string
}

func (c *rdsHandler) makeRequest(pod *PodInfo) *xdsapi.DiscoveryRequest {
	return pod.appendResources(pod.makeRequest("rds"), c.resources)
}

func (c *rdsHandler) onXDSResponse(resp *xdsapi.DiscoveryResponse) error {
	outputJSON(resp)
	return nil
}
