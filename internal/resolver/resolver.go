// Package resolver builds the typed easydi dependency graph.
package resolver

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/ramory-l/easydi/internal/parampath"
	"github.com/ramory-l/easydi/internal/scanner"
)

// Root is a di:root in the resolved graph, with the lowercased variable name
// it will take as a parameter of the generated Build function.
type Root struct {
	Name    string // root type name (node name)
	VarName string // generated Build parameter name (lowercased)
	Type    types.Type
}

// Binding is how one provider parameter is satisfied: either by another
// provider node (FromNode) or by a generated expression (Expr) such as a
// root projection or a package-qualified literal.
type Binding struct {
	Param    scanner.Param
	FromNode *Node  // non-nil when satisfied by another provider
	Expr     string // generated expression (root projection / literal); "" when FromNode set
}

// Node is a provider together with the resolved bindings for its parameters.
type Node struct {
	Provider *scanner.Provider
	Bindings []Binding
}

// Name returns the node's graph name (explicit di:provide name= or the
// function name).
func (n *Node) Name() string { return n.Provider.Name }

// Graph is the resolved typed dependency graph: provider nodes plus the
// di:root inputs.
type Graph struct {
	Nodes  []*Node
	Roots  []*Root
	byName map[string]*Node
}

// NodeByName returns the node with the given graph name, or nil.
func (g *Graph) NodeByName(name string) *Node { return g.byName[name] }

// Resolve builds the typed dependency graph from scanned providers and roots:
// it resolves each provider parameter by di:param projection or by Go type
// (strict pointer/value rules; interfaces via Implements), reporting missing
// or ambiguous dependencies.
func Resolve(r *scanner.Result) (*Graph, error) {
	g := &Graph{byName: map[string]*Node{}}

	for _, root := range r.Roots {
		g.Roots = append(g.Roots, &Root{
			Name:    root.Name,
			VarName: strings.ToLower(root.Name),
			Type:    root.Type,
		})
	}
	rootByName := map[string]*Root{}
	for _, rt := range g.Roots {
		rootByName[rt.Name] = rt
	}

	for _, p := range r.Providers {
		if _, dup := g.byName[p.Name]; dup {
			return nil, fmt.Errorf("duplicate provider node name %q", p.Name)
		}
		n := &Node{Provider: p}
		g.Nodes = append(g.Nodes, n)
		g.byName[p.Name] = n
	}

	for _, n := range g.Nodes {
		for _, param := range n.Provider.Params {
			b, err := resolveParam(g, rootByName, n, param)
			if err != nil {
				return nil, fmt.Errorf("provide %s: %w", n.Name(), err)
			}
			n.Bindings = append(n.Bindings, b)
		}
	}
	return g, nil
}

func resolveParam(g *Graph, roots map[string]*Root, n *Node, param scanner.Param) (Binding, error) {
	if param.Use != "" {
		target := g.byName[param.Use]
		if target == nil {
			return Binding{}, fmt.Errorf("di:use %s: no provider node named %s (parameter %s)", param.Use, param.Use, param.Name)
		}
		if !satisfies(target.Provider.Produces, param.Type) {
			return Binding{}, fmt.Errorf("di:use %s: node %s (%s) not assignable to parameter %s (%s)",
				param.Use, param.Use, target.Provider.Produces, param.Name, param.Type)
		}
		return Binding{Param: param, FromNode: target}, nil
	}
	if param.Path != nil {
		return resolvePath(roots, param)
	}
	// Resolve by type.
	var matches []*Node
	for _, cand := range g.Nodes {
		if cand == n {
			continue
		}
		if satisfies(cand.Provider.Produces, param.Type) {
			matches = append(matches, cand)
		}
	}
	switch len(matches) {
	case 1:
		return Binding{Param: param, FromNode: matches[0]}, nil
	case 0:
		return Binding{}, fmt.Errorf("no provider for parameter %s (%s); add // di:provide or // di:param",
			param.Name, param.Type)
	default:
		var names []string
		for _, m := range matches {
			names = append(names, m.Name())
		}
		return Binding{}, fmt.Errorf("parameter %s (%s) is ambiguous between %s; disambiguate with name=",
			param.Name, param.Type, strings.Join(names, ", "))
	}
}

// satisfies enforces strict pointer/value rules: identical concrete types, or
// the parameter is an interface implemented by the provider's actual produced
// type. No implicit address-of/dereference.
func satisfies(produced, want types.Type) bool {
	if types.Identical(produced, want) {
		return true
	}
	if iface, ok := want.Underlying().(*types.Interface); ok {
		return types.Implements(produced, iface)
	}
	return false
}

func resolvePath(roots map[string]*Root, param scanner.Param) (Binding, error) {
	head := param.Path.Segs[0].Name
	root, isRoot := roots[head]
	if !isRoot {
		// Package-qualified literal, e.g. time.Now. Exactly two segments,
		// no calls. Type-checked by the generated file's compilation.
		if len(param.Path.Segs) != 2 || param.Path.Segs[0].Call || param.Path.Segs[1].Call {
			return Binding{}, fmt.Errorf("di:param %q: not a root projection and not a pkg.Ident literal",
				param.Path.String())
		}
		return Binding{Param: param, Expr: param.Path.String()}, nil
	}

	// Root projection: walk the type through the remaining segments.
	cur := root.Type
	for _, seg := range param.Path.Segs[1:] {
		next, err := stepType(cur, seg)
		if err != nil {
			return Binding{}, fmt.Errorf("di:param %s: %w", param.Path.String(), err)
		}
		cur = next
	}
	if !types.AssignableTo(cur, param.Type) {
		return Binding{}, fmt.Errorf("cannot use %s (%s) as %s for parameter %s",
			param.Path.String(), cur, param.Type, param.Name)
	}
	expr := root.VarName
	if rest := segStrings(param.Path.Segs[1:]); len(rest) > 0 {
		expr += "." + strings.Join(rest, ".")
	}
	return Binding{Param: param, Expr: expr}, nil
}

func segStrings(segs []parampath.Seg) []string {
	out := make([]string, len(segs))
	for i, s := range segs {
		out[i] = s.Name
		if s.Call {
			out[i] += "()"
		}
	}
	return out
}

func deref(t types.Type) types.Type {
	if ptr, ok := t.Underlying().(*types.Pointer); ok {
		return ptr.Elem()
	}
	return t
}

func stepType(t types.Type, seg parampath.Seg) (types.Type, error) {
	if seg.Call {
		// Zero-arg method returning exactly one value.
		ms := types.NewMethodSet(t)
		for i := 0; i < ms.Len(); i++ {
			m := ms.At(i)
			if m.Obj().Name() != seg.Name {
				continue
			}
			sig := m.Obj().Type().(*types.Signature)
			if sig.Params().Len() != 0 || sig.Results().Len() != 1 {
				return nil, fmt.Errorf("method %s must take no args and return one value", seg.Name)
			}
			return sig.Results().At(0).Type(), nil
		}
		return nil, fmt.Errorf("no zero-arg method %s on %s", seg.Name, t)
	}
	st, ok := deref(t).Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("%s is not a struct; cannot select field %s", t, seg.Name)
	}
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if f.Name() == seg.Name {
			return f.Type(), nil
		}
	}
	return nil, fmt.Errorf("no field %s on %s", seg.Name, t)
}
