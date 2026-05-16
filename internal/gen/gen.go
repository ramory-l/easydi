// Package gen emits the easydi container source file.
package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"sort"
	"strings"

	"github.com/ramory-l/easydi/internal/resolver"
)

// importTracker records the packages referenced while building type strings.
// Aliases are assigned in a second pass once the full set is known.
type importTracker struct {
	// path -> package identifier (types.Package.Name()).
	names map[string]string
	// path -> final alias (filled by finalize()).
	alias map[string]string
	final bool
}

func newImportTracker() *importTracker {
	return &importTracker{names: map[string]string{}, alias: map[string]string{}}
}

func (it *importTracker) qualifier(p *types.Package) string {
	if p == nil {
		return ""
	}
	it.names[p.Path()] = p.Name()
	if it.final {
		return it.alias[p.Path()]
	}
	// Pre-finalize (collecting pass): a stable placeholder; the real type
	// strings are produced only after finalize().
	return p.Name()
}

// add records a package path that must be imported even if no type string
// referenced it (e.g. the provider call qualifier, std-lib helpers).
func (it *importTracker) add(path, name string) {
	it.names[path] = name
}

func (it *importTracker) finalize() {
	paths := make([]string, 0, len(it.names))
	for p := range it.names {
		paths = append(paths, p)
	}
	it.alias = computeAliases(paths, it.names)
	it.final = true
}

func (it *importTracker) block() string {
	if len(it.names) == 0 {
		return ""
	}
	paths := make([]string, 0, len(it.names))
	for p := range it.names {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var b strings.Builder
	b.WriteString("import (\n")
	for _, p := range paths {
		if a := it.alias[p]; a != "" && a != it.names[p] {
			fmt.Fprintf(&b, "\t%s %q\n", a, p)
		} else {
			fmt.Fprintf(&b, "\t%q\n", p)
		}
	}
	b.WriteString(")\n")
	return b.String()
}

const lifecycleSrc = `
// Start starts every exposed lifecycle.Starter in dependency order. If a
// Starter fails, Start closes every exposed lifecycle.Closer constructed
// before it (in reverse order) and returns the error. After Start returns a
// non-nil error the container is fully unwound: callers must NOT call Close.
func (c *Container) Start(ctx context.Context) error {
	nodes := c.Exposed()
	for i, n := range nodes {
		s, ok := n.(lifecycle.Starter)
		if !ok {
			continue
		}
		if err := s.Start(ctx); err != nil {
			for j := i - 1; j >= 0; j-- {
				if cl, ok := nodes[j].(lifecycle.Closer); ok {
					_ = cl.Close(ctx)
				}
			}
			return err
		}
	}
	return nil
}

// Close closes every exposed lifecycle.Closer in reverse dependency order and
// returns the joined error of all failures. Do not call Close after Start has
// returned a non-nil error (Start already unwound).
func (c *Container) Close(ctx context.Context) error {
	nodes := c.Exposed()
	var errs []error
	for i := len(nodes) - 1; i >= 0; i-- {
		if cl, ok := nodes[i].(lifecycle.Closer); ok {
			if err := cl.Close(ctx); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}
`

// Generate emits the gofmt-formatted source of the easydi container for the
// given resolved graph: a Container struct, a Build(<roots>) (*Container,
// error) constructor with providers in topological order, an Exposed()
// accessor for di:expose nodes, and Start/Close lifecycle methods over the
// exposed nodes. pkgName sets the package clause.
func Generate(g *resolver.Graph, order []*resolver.Node, pkgName string) ([]byte, error) {
	it := newImportTracker()

	// emitBody renders the full body using the current qualifier state.
	emitBody := func() string {
		q := it.qualifier
		ts := func(t types.Type) string { return types.TypeString(unaliasDeep(t), q) }

		var body bytes.Buffer

		body.WriteString("type Container struct {\n")
		for _, n := range order {
			fmt.Fprintf(&body, "\t%s %s\n", n.Name(), ts(n.Provider.Produces))
		}
		body.WriteString("}\n\n")

		var params []string
		for _, r := range g.Roots {
			params = append(params, fmt.Sprintf("%s %s", r.VarName, ts(r.Type)))
		}
		fmt.Fprintf(&body, "func Build(%s) (*Container, error) {\n", strings.Join(params, ", "))
		body.WriteString("\tc := &Container{}\n")

		for _, n := range order {
			args := make([]string, 0, len(n.Bindings))
			for _, b := range n.Bindings {
				if b.FromNode != nil {
					args = append(args, "c."+b.FromNode.Name())
				} else {
					args = append(args, b.Expr)
				}
			}
			call := fmt.Sprintf("%s.%s(%s)",
				it.qualifier(n.Provider.Pkg.Types),
				n.Provider.FuncName,
				strings.Join(args, ", "))
			if n.Provider.ReturnsError {
				fmt.Fprintf(&body, "\tv%s, err := %s\n", n.Name(), call)
				fmt.Fprintf(&body, "\tif err != nil {\n\t\treturn nil, fmt.Errorf(\"provide %s: %%w\", err)\n\t}\n", n.Name())
				fmt.Fprintf(&body, "\tc.%s = v%s\n", n.Name(), n.Name())
			} else {
				fmt.Fprintf(&body, "\tc.%s = %s\n", n.Name(), call)
			}
		}
		body.WriteString("\treturn c, nil\n}\n\n")

		var exposed []string
		for _, n := range order {
			if n.Provider.Expose {
				exposed = append(exposed, "c."+n.Name())
			}
		}
		body.WriteString("func (c *Container) Exposed() []any {\n")
		if len(exposed) == 0 {
			body.WriteString("\treturn nil\n")
		} else {
			fmt.Fprintf(&body, "\treturn []any{%s}\n", strings.Join(exposed, ", "))
		}
		body.WriteString("}\n")

		body.WriteString(lifecycleSrc)
		return body.String()
	}

	// Pass 1: collect every referenced package.
	_ = emitBody()
	for _, n := range order {
		if n.Provider.ReturnsError {
			it.add("fmt", "fmt")
			break
		}
	}
	it.add("context", "context")
	it.add("errors", "errors")
	it.add("github.com/ramory-l/easydi/lifecycle", "lifecycle")

	// Pass 2: assign unique aliases, then render with final qualifiers.
	it.finalize()
	bodyStr := emitBody()

	var file bytes.Buffer
	fmt.Fprintf(&file, "// Code generated by easydi. DO NOT EDIT.\n\npackage %s\n\n", pkgName)
	file.WriteString(it.block())
	file.WriteString("\n")
	file.WriteString(bodyStr)

	formatted, err := format.Source(file.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w\n%s", err, file.String())
	}
	return formatted, nil
}

// unaliasDeep is a temporary identity shim; Task 3 will replace this with a
// real implementation in unalias.go that unwraps named types defined in
// aliased packages so that types.TypeString produces correct qualified names.
func unaliasDeep(t types.Type) types.Type { return t }
