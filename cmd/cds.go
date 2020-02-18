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
	"fmt"

	"github.com/spf13/cobra"

	xdsapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/ptypes"

	"istio.io/pkg/log"
)

func cds() *cobra.Command {
	handler := &cdsHandler{}
	localCmd := makeXDSCmd("cds", handler)
	localCmd.Flags().StringVarP(&handler.matchName, "name", "n", "", "Show only cluster with this name")
	localCmd.Flags().BoolVarP(&handler.showAll, "all", "a", false, "If set, output the whole CDS response.")
	return localCmd
}

type cdsHandler struct {
	matchName string
	showAll   bool
}

func (c *cdsHandler) makeRequest(pod *PodInfo) *xdsapi.DiscoveryRequest {
	return pod.makeRequest("cds")
}

func (c *cdsHandler) onXDSResponse(resp *xdsapi.DiscoveryResponse) error {
	if c.showAll {
		outputJSON(resp)
		return nil
	}
	seenClusters := make([]string, 0, len(resp.Resources))
	for _, res := range resp.Resources {
		cluster := &xdsapi.Cluster{}
		if err := ptypes.UnmarshalAny(res, cluster); err != nil {
			log.Errorf("Cannot unmarshal any proto to cluster: %v", err)
			continue
		}
		seenClusters = append(seenClusters, cluster.Name)

		if c.matchName == cluster.Name {
			outputJSON(cluster)
			return nil
		}
	}
	msg := fmt.Sprintf("Cannot find any listener with name %q. Seen:\n", c.matchName)
	for _, c := range seenClusters {
		msg += fmt.Sprintf("  %s\n", c)
	}
	return fmt.Errorf("%s", msg)
}
