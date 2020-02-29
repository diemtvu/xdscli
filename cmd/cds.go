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
	"strings"

	"github.com/spf13/cobra"

	xdsapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/ptypes"

	"istio.io/pkg/log"
)

func cds() *cobra.Command {
	handler := &cdsHandler{}
	localCmd := makeXDSCmd("cds", handler)
	// localCmd.Flags().StringVarP(&handler.matchName, "resource", "r", "", "Show only cluster with this name")
	localCmd.Flags().StringVarP(&handler.fqdn, "fqdn", "", "", "Filter clusters by substring of Service FQDN field")
	localCmd.Flags().StringVarP(&handler.direction, "direction", "d", "", "Filter clusters by Direction field")
	localCmd.Flags().StringVarP(&handler.subset, "subset", "", "", "Filter clusters by substring of Subset field")
	localCmd.Flags().Uint32VarP(&handler.port, "port", "", 0, "Filter clusters by Port field")
	return localCmd
}

type cdsHandler struct {
	// matchName string
	fqdn      string
	direction string
	subset    string
	port      uint32
	showAll   bool
}

func (c *cdsHandler) makeRequest(pod *PodInfo) *xdsapi.DiscoveryRequest {
	return pod.makeRequest("cds")
}

func (c *cdsHandler) Match(cluster *xdsapi.Cluster) bool {
	name := cluster.Name
	if c.fqdn == "" && c.port == 0 && c.subset == "" && c.direction == "" {
		return true
	}
	if c.fqdn != "" && !strings.Contains(name, string(c.fqdn)) {
		return false
	}
	if c.direction != "" && !strings.Contains(name, string(c.direction)) {
		return false
	}
	if c.subset != "" && !strings.Contains(name, c.subset) {
		return false
	}
	if c.port != 0 {
		p := fmt.Sprintf("|%v|", c.port)
		if !strings.Contains(name, p) {
			return false
		}
	}
	return true
}

func (c *cdsHandler) onXDSResponse(resp *xdsapi.DiscoveryResponse) error {
	if len(c.fqdn) == 0 || c.fqdn == "*" || c.fqdn == "all" {
		outputJSON(resp)
		return nil
	}
	seenClusters := make(map[string]*xdsapi.Cluster, len(resp.Resources))
	for _, res := range resp.Resources {
		cluster := &xdsapi.Cluster{}
		if err := ptypes.UnmarshalAny(res, cluster); err != nil {
			log.Errorf("Cannot unmarshal any proto to cluster: %v", err)
			continue
		}
		seenClusters[cluster.Name] = cluster

		if c.Match(cluster) {
			outputJSON(cluster)
			return nil
		}
	}
	msg := fmt.Sprintf("Cannot find any cluster for FQDN %q. Seen:\n", c.fqdn)
	msg += fmt.Sprintln("SERVICE FQDN\tPORT\tSUBSET\tDIRECTION\tTYPE")
	for _, c := range seenClusters {
		msg += fmt.Sprintf("  %s\n", c)
	}
	return fmt.Errorf("%s", msg)
}
