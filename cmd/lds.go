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
	listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"

	// core1 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/ptypes"

	"istio.io/istio/pilot/pkg/networking/util"
	"istio.io/pkg/log"
)

func lds() *cobra.Command {
	handler := &ldsHandler{}
	localCmd := makeXDSCmd("lds", handler)
	localCmd.Flags().StringVarP(&handler.matchName, "resource", "r", "virtualInbound", "Show only listener with this name")
	localCmd.Flags().StringVarP(&handler.matchAddress, "address", "a", "", "Filter listeners by address field")
	localCmd.Flags().StringVarP(&handler.matchType, "type", "y", "", "Filter listeners by type field")
	localCmd.Flags().Uint32VarP(&handler.matchPort, "port", "", 0, "Filter listeners by Port field")
	return localCmd
}

type ldsHandler struct {
	matchName    string
	matchAddress string
	matchType    string
	matchPort    uint32
	showAll      bool
}

func (c *ldsHandler) makeRequest(pod *PodInfo) *xdsapi.DiscoveryRequest {
	return pod.makeRequest("lds")
}

func matchAddress(address string, port uint32, filterChainMatch *listener.FilterChainMatch) bool {
	log.Debugf("Matching %s:%d to %v : %v", address, port, filterChainMatch.PrefixRanges, filterChainMatch.DestinationPort)
	if port != 0 && filterChainMatch.DestinationPort != nil {
		if port != filterChainMatch.DestinationPort.Value {
			return false
		}
	}

	if len(address) != 0 {
		for _, prefix := range filterChainMatch.PrefixRanges {
			if address == prefix.AddressPrefix {
				return true
			}
		}
		return false
	}

	return true
}

func retrieveListenerAddress(l *xdsapi.Listener) string {
	return l.Address.GetSocketAddress().Address
}

func retrieveListenerPort(l *xdsapi.Listener) uint32 {
	return l.Address.GetSocketAddress().GetPortValue()
}

const (
	// HTTPListener identifies a listener as being of HTTP type by the presence of an HTTP connection manager filter
	HTTPListener = "envoy.http_connection_manager"

	// TCPListener identifies a listener as being of TCP type by the presence of TCP proxy filter
	TCPListener = "envoy.tcp_proxy"
)

// retrieveListenerType classifies a Listener as HTTP|TCP|HTTP+TCP|UNKNOWN
func retrieveListenerType(l *xdsapi.Listener) string {
	nHTTP := 0
	nTCP := 0
	for _, filterChain := range l.GetFilterChains() {
		for _, filter := range filterChain.GetFilters() {
			if filter.Name == HTTPListener {
				nHTTP++
			} else if filter.Name == TCPListener {
				if !strings.Contains(string(filter.GetTypedConfig().GetValue()), util.BlackHoleCluster) {
					nTCP++
				}
			}
		}
	}

	if nHTTP > 0 {
		if nTCP == 0 {
			return "HTTP"
		}
		return "HTTP+TCP"
	} else if nTCP > 0 {
		return "TCP"
	}

	return "UNKNOWN"
}

func (c *ldsHandler) onXDSResponse(resp *xdsapi.DiscoveryResponse) error {
	if len(c.matchName) == 0 || c.matchName == "*" || c.matchName == "all" {
		outputJSON(resp)
		return nil
	}
	seenListener := make(map[string]*xdsapi.Listener, len(resp.Resources))
	for _, res := range resp.Resources {
		listener := &xdsapi.Listener{}
		if err := ptypes.UnmarshalAny(res, listener); err != nil {
			log.Errorf("Cannot unmarshal any proto to listener: %v", err)
			continue
		}

		seenListener[listener.Name] = listener
		if c.matchName == listener.Name {
			if len(c.matchAddress) == 0 && c.matchPort == 0 {
				outputJSON(listener)
			}
			for _, ch := range listener.FilterChains {
				if matchAddress(c.matchAddress, c.matchPort, ch.FilterChainMatch) {
					outputJSON(ch)
				}
			}
			return nil
		}
	}
	msg := fmt.Sprintf("Cannot find any listener with name %q. Seen:\n", c.matchName)
	msg += fmt.Sprintln("NAME\tADDRESS\tPORT\tTYPE")
	for name, listener := range seenListener {
		address := retrieveListenerAddress(listener)
		port := retrieveListenerPort(listener)
		listenerType := retrieveListenerType(listener)
		msg += fmt.Sprintf("%s\t%s\t%v\t%v\n", name, address, port, listenerType)
	}
	return fmt.Errorf("%s", msg)
}
