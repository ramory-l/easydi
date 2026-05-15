// Package topo topologically sorts the easydi graph.
package topo

import (
	"fmt"
	"strings"

	"github.com/ramory-l/easydi/internal/resolver"
)

const (
	white = 0
	gray  = 1
	black = 2
)

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
