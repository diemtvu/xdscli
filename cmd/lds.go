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

func lds() *cobra.Command {
	handler := &ldsHandler{}
	localCmd := makeXDSCmd("lds", handler)
	localCmd.Flags().StringVarP(&handler.matchName, "name", "n", "", "Show only listener with this name")
	localCmd.Flags().BoolVarP(&handler.showAll, "all", "a", false, "If set, output the whole LDS response.")
	return localCmd
}

type ldsHandler struct {
	matchName string
	showAll   bool
}

func (c *ldsHandler) makeRequest(pod *PodInfo) *xdsapi.DiscoveryRequest {
	return pod.makeRequest("lds")
}

func (c *ldsHandler) onXDSResponse(resp *xdsapi.DiscoveryResponse) error {
	if c.showAll {
		outputJSON(resp)
		return nil
	}
	seenListener := make([]string, 0, len(resp.Resources))
	for _, res := range resp.Resources {
		listener := &xdsapi.Listener{}
		if err := ptypes.UnmarshalAny(res, listener); err != nil {
			log.Errorf("Cannot unmarshal any proto to listener: %v", err)
			continue
		}

		seenListener = append(seenListener, listener.Name)
		if c.matchName == listener.Name {
			outputJSON(listener)
			return nil
		}
	}
	msg := fmt.Sprintf("Cannot find any listener with name %q. Seen:\n", c.matchName)
	for _, c := range seenListener {
		msg += fmt.Sprintf("  %s\n", c)
	}
	return fmt.Errorf("%s", msg)
}
