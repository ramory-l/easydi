// Package topo topologically sorts the easydi graph.
package topo

import (
	"fmt"
	"strings"

	"github.com/ramory-l/easydi/internal/resolver"
)

// DFS visitation colors: white = unvisited, gray = on the current stack
// (a back-edge to gray is a cycle), black = fully processed.
const (
	white = 0
	gray  = 1
	black = 2
)

// Sort returns the graph's nodes in dependency order (dependencies before
// dependents) using a tri-color DFS, reporting an error if a cycle exists.
func Sort(g *resolver.Graph) ([]*resolver.Node, error) {
	color := map[*resolver.Node]int{}
	var order []*resolver.Node

	var visit func(n *resolver.Node, stack []string) error
	visit = func(n *resolver.Node, stack []string) error {
		switch color[n] {
		case gray:
			return fmt.Errorf("dependency cycle: %s -> %s",
				strings.Join(stack, " -> "), n.Name())
		case black:
			return nil
		}
		color[n] = gray
		for _, b := range n.Bindings {
			if b.FromNode == nil {
				continue
			}
			if err := visit(b.FromNode, append(stack, n.Name())); err != nil {
				return err
			}
		}
		color[n] = black
		order = append(order, n)
		return nil
	}

	for _, n := range g.Nodes {
		if err := visit(n, nil); err != nil {
			return nil, err
		}
	}
	return order, nil
}
