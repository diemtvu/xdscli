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
