package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	xdsapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/ptypes"
	any "github.com/golang/protobuf/ptypes/any"
	"istio.io/istio/pilot/pkg/model"

	"istio.io/pkg/log"
)

func cds() *cobra.Command {
	handler := &cdsHandler{}
	localCmd := makeXDSCmd("cds", handler)
	localCmd.Flags().StringVarP(&handler.matchName, "resource", "r", "", "Show only cluster with this name")
	localCmd.Flags().StringVarP(&handler.fqdn, "fqdn", "", "", "Filter clusters by substring of Service FQDN field")
	localCmd.Flags().StringVarP(&handler.direction, "direction", "d", "", "Filter clusters by Direction field")
	localCmd.Flags().StringVarP(&handler.subset, "subset", "", "", "Filter clusters by substring of Subset field")
	localCmd.Flags().Uint32VarP(&handler.port, "port", "p", 0, "Filter clusters by Port field")
	return localCmd
}

type cdsHandler struct {
	matchName string
	fqdn      string
	direction string
	subset    string
	port      uint32
	showAll   bool
}

func (c *cdsHandler) makeRequest(pod *PodInfo) *xdsapi.DiscoveryRequest {
	return pod.makeRequest("cds")
}

func (c *cdsHandler) match(cluster *xdsapi.Cluster) bool {
	name := cluster.Name
	if c.matchName != "" && c.matchName != name {
		return false
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
	if len(c.matchName) == 0 || c.matchName == "*" || c.matchName == "all" {
		c.output(resp)
		return nil
	}
	filterResp := &xdsapi.DiscoveryResponse{
		Resources: []*any.Any{},
	}
	for _, res := range resp.Resources {
		cluster := &xdsapi.Cluster{}
		if err := ptypes.UnmarshalAny(res, cluster); err != nil {
			log.Errorf("Cannot unmarshal any proto to cluster: %v", err)
			continue
		}
		
		if c.match(cluster) {
			filterResp.Resources = append(filterResp.Resources, res)
		}
	}

	if len(filterResp.Resources) == 0{
		return fmt.Errorf("Cannot find cluster matched conditions. Found:\n%s", c.outputShort(resp))
	}
	c.output(filterResp)
	return nil
}

func (c *cdsHandler) output(resp *xdsapi.DiscoveryResponse) {
	if outputFormat == "json" {
		outputJSON(resp)
		return
	}

	fmt.Println(c.outputShort(resp))
}

func retrieveSocketMatch(cluster *xdsapi.Cluster) []string {
	ret := make([]string, 0, len(cluster.TransportSocketMatches))
	for _, m := range cluster.TransportSocketMatches {
		ret = append(ret, m.Name)
	}
	return ret
}

func (c *cdsHandler) outputShort(resp *xdsapi.DiscoveryResponse) string {
	var buf bytes.Buffer
	w := new(tabwriter.Writer).Init(&buf, 0, 8, 5, ' ', 0)
	fmt.Fprintln(w,  "SERVICE FQDN\tPORT\tSUBSET\tDIRECTION\tTYPE\tSOCKET_MATCH")
	for _, res := range resp.Resources {
		cluster := &xdsapi.Cluster{}
		if err := ptypes.UnmarshalAny(res, cluster); err != nil {
			log.Errorf("Cannot unmarshal any proto to cluster: %v", err)
			continue
		}
		if len(strings.Split(cluster.Name, "|")) > 3 {
			direction, subset, fqdn, port := model.ParseSubsetKey(cluster.Name)
			if subset == "" {
				subset = "-"
			}
			_, _ = fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%s\t%v\n", fqdn, port, subset, direction, cluster.GetType(), retrieveSocketMatch(cluster))
		} else {
			_, _ = fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%s\t%v\n", cluster.Name, "-", "-", "-", cluster.GetType(), retrieveSocketMatch(cluster))
		}
	}
	w.Flush()
	return buf.String()
}