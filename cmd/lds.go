package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	xdsapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"

	"github.com/golang/protobuf/ptypes"
	any "github.com/golang/protobuf/ptypes/any"

	"istio.io/istio/pilot/pkg/networking/util"
	"istio.io/pkg/log"
)

func lds() *cobra.Command {
	handler := &ldsHandler{}
	localCmd := makeXDSCmd("lds", handler)
	localCmd.Flags().StringVarP(&handler.matchName, "resource", "r", "virtualInbound", "Filter listeners by name")
	localCmd.Flags().StringVarP(&handler.matchAddress, "address", "a", "", "Filter listeners by address field")
	localCmd.Flags().StringVarP(&handler.matchType, "type", "", "", "Filter listeners by type field")
	localCmd.Flags().Uint32VarP(&handler.matchPort, "port", "p", 0, "Filter listeners by Port field")
	localCmd.Flags().StringVarP(&handler.matchChainAddress, "chain-address", "", "", "Filter listeners filter-chain by address field")
	localCmd.Flags().Uint32VarP(&handler.matchChainPort, "chain-port", "", 0, "Filter listeners by destination port field")
	return localCmd
}

type ldsHandler struct {
	matchName         string
	matchAddress      string
	matchPort         uint32
	matchType         string
	matchChainAddress string
	matchChainPort    uint32
	showAll           bool
}

func (c *ldsHandler) makeRequest(pod *PodInfo) *xdsapi.DiscoveryRequest {
	return pod.makeRequest("lds")
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

func (c *ldsHandler) filter(l *xdsapi.Listener) *xdsapi.Listener {
	if !c.matchFilter(l) {
		return nil
	}
	if len(c.matchAddress) == 0 && c.matchChainPort == 0 {
		return l
	}
	newL := &xdsapi.Listener{
		Name:         l.Name,
		Address:      l.Address,
		FilterChains: []*listener.FilterChain{},
	}
	for _, ch := range l.FilterChains {
		if c.matchFilterChain(ch) {
			newL.FilterChains = append(newL.FilterChains, ch)
		}
	}
	return newL
}

func (c *ldsHandler) matchFilter(l *xdsapi.Listener) bool {
	if c.matchName != l.Name {
		return false
	}
	if len(c.matchAddress) != 0 && c.matchAddress != retrieveListenerAddress(l) {
		return false
	}
	if c.matchPort != 0 && c.matchPort != retrieveListenerPort(l) {
		return false
	}
	return true
}

func (c *ldsHandler) matchFilterChain(chain *listener.FilterChain) bool {
	if len(c.matchChainAddress) != 0 {
		matched := false
		for _, prefix := range chain.FilterChainMatch.PrefixRanges {
			if c.matchChainAddress == prefix.AddressPrefix {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if c.matchChainPort != 0 && c.matchChainPort != chain.FilterChainMatch.DestinationPort.Value {
		return false
	}
	return true
}

func (c *ldsHandler) onXDSResponse(resp *xdsapi.DiscoveryResponse) error {
	if len(c.matchName) == 0 || c.matchName == "*" || c.matchName == "all" {
		c.output(resp)
		return nil
	}
	filterResp := &xdsapi.DiscoveryResponse{
		Resources: []*any.Any{},
	}
	for _, res := range resp.Resources {
		listener := &xdsapi.Listener{}
		if err := ptypes.UnmarshalAny(res, listener); err != nil {
			log.Errorf("Cannot unmarshal any proto to listener: %v", err)
			continue
		}
		if filterListener := c.filter(listener); filterListener != nil {
			if r, err := ptypes.MarshalAny(filterListener); err != nil {
				log.Errorf("Cannot marshal listener to any proto: %v", err)
			} else {
				filterResp.Resources = append(filterResp.Resources, r)
			}
		}
	}
	if len(filterResp.Resources) != 0 {
		c.output(filterResp)
		return nil
	}
	return fmt.Errorf("Cannot find listener matching conditions. Seen listeners:\n%s", c.outputShort(resp))
}

func retrieveTransportProtocol(ch *listener.FilterChain) string {
	if ch.TransportSocket == nil {
		return "TCP"
	}
	tlsContext := &auth.DownstreamTlsContext{}
	if err := ptypes.UnmarshalAny(ch.TransportSocket.GetTypedConfig(), tlsContext); err != nil {
		log.Errorf("Cannot unmarshal any proto to TLSContext: %v", err)
		return "UNKNOWN"
	}
	if tlsContext.RequireClientCertificate.Value {
		return "MTLS"
	}
	return "TLS"
}

func (c *ldsHandler) output(resp *xdsapi.DiscoveryResponse) {
	if outputFormat == "json" {
		outputJSON(resp)
		return
	}

	fmt.Println(c.outputShort(resp))
}

func retrieveAllTransportProtocol(vs []*listener.FilterChain) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = retrieveTransportProtocol(v)
	}
	return vsm
}

func (c *ldsHandler) outputShort(resp *xdsapi.DiscoveryResponse) string {
	var buf bytes.Buffer
	w := new(tabwriter.Writer).Init(&buf, 0, 8, 5, ' ', 0)
	fmt.Fprintln(w, "NAME\tADDRESS\tPORT\tTYPE\tPROTOCOL")
	for _, res := range resp.Resources {
		listener := &xdsapi.Listener{}
		if err := ptypes.UnmarshalAny(res, listener); err != nil {
			log.Errorf("Cannot unmarshal any proto to listener: %v", err)
			continue
		}
		address := retrieveListenerAddress(listener)
		port := retrieveListenerPort(listener)
		listenerType := retrieveListenerType(listener)
		fmt.Fprintf(w, "%s\t%s\t%v\t%v\t%s\n", listener.Name, address, port, listenerType, retrieveAllTransportProtocol(listener.FilterChains))
	}
	w.Flush()
	return buf.String()
}
