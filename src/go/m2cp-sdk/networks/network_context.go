package networks

/*
* A collection of information about the network connection, which is shared between all elements being part of the
* network connection.
 */

import (
	"fmt"
	"m2cp"
)

type NetworkContext struct {
	Domain *domain
	//Stats  *TrafficStatsCombined
	ctp   m2cp.ContextPlus
	nodes map[string]m2cp.Node
}

func NewNetworkContext(domain *domain, ctpParent m2cp.ContextPlus) *NetworkContext {
	return &NetworkContext{
		Domain: domain,
		//Stats:  NewTrafficStatsCombined(),
		ctp:   ctpParent.BranchWithName("networks.NetworkContext"),
		nodes: make(map[string]m2cp.Node, 0),
	}
}

func (nc *NetworkContext) Stop() {
	nc.ctp.Done()
}

func (nc *NetworkContext) AddNode(node m2cp.Node) {
	nc.nodes[node.GetName()] = node
}

func (nc *NetworkContext) GetNodeByAddress(address string) m2cp.Node {
	for _, n := range nc.nodes {
		if n.GetAddress() == address {
			return n
		}
	}
	return nil
}

func (nc *NetworkContext) GetNodeByName(name string) (m2cp.Node, error) {
	if n, ok := nc.nodes[name]; ok {
		return n, nil
	}
	return nil, fmt.Errorf("node '%s' not found", name)
}

func (nc *NetworkContext) GetNodes() []m2cp.Node {
	nodes := make([]m2cp.Node, 0)
	for _, n := range nc.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}
