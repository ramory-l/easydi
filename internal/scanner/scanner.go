// Package scanner walks loaded packages and extracts easydi providers/roots.
package scanner

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/ramory-l/easydi/internal/annotation"
	"github.com/ramory-l/easydi/internal/parampath"
)

// Param is a single constructor parameter of a di:provide function.
type Param struct {
	Name string
	Type types.Type
	Path *parampath.Path // nil = resolve by type
	Use  string          // di:use node name; "" = none
}

// Provider is a di:provide constructor function and the metadata needed to
// turn it into a dependency-graph node.
type Provider struct {
	Pkg          *packages.Package
	FuncName     string
	Name         string // node name: explicit name= or FuncName
	Produces     types.Type
	ReturnsError bool
	Expose       bool
	Params       []Param
}

// Root is a di:root type: an external value supplied to the generated Build
// function and projected from by di:param paths.
type Root struct {
	Name string
	Type types.Type
}

// Result is the set of providers and roots discovered across all scanned
// packages.
type Result struct {
	Providers []*Provider
	Roots     []*Root
}

// Scan walks the syntax of the loaded packages and collects every di:root
// type and di:provide function, mapping each di:param comment to its
// parameter by source line.
func Scan(pkgs []*packages.Package) (*Result, error) {
	res := &Result{}
	for _, p := range pkgs {
		for _, file := range p.Syntax {
			lineDirectives := indexComments(p, file)
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.GenDecl:
					if err := scanRoots(p, d, res); err != nil {
						return nil, err
					}
				case *ast.FuncDecl:
					if err := scanProvider(p, d, lineDirectives, res); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return res, nil
}

// indexComments maps a source line number to the di directive that ends on
// that line, for every comment in the file.
func indexComments(p *packages.Package, file *ast.File) map[int]annotation.Directive {
	out := map[int]annotation.Directive{}
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			d, ok, err := annotation.Parse(text)
			if err != nil {
				// surfaced later if it actually annotates something; here
				// we still index nothing for malformed lines.
				continue
			}
			if !ok {
				continue
			}
			line := p.Fset.Position(c.End()).Line
			out[line] = d
		}
	}
	return out
}

// docDirectives parses the di directives present in a declaration's doc
// comment group, reporting whether it carries di:provide (and its name=),
// di:root, or di:expose.
func docDirectives(doc *ast.CommentGroup) (provide bool, name string, root, expose bool, err error) {
	if doc == nil {
		return
	}
	for _, c := range doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		d, ok, perr := annotation.Parse(text)
		if perr != nil {
			return false, "", false, false, perr
		}
		if !ok {
			continue
		}
		switch d.Kind {
		case annotation.Provide:
			provide = true
			name = d.Name
		case annotation.Root:
			root = true
		case annotation.Expose:
			expose = true
		}
	}
	return
}

func scanRoots(p *packages.Package, d *ast.GenDecl, res *Result) error {
	for _, spec := range d.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		// di:root may sit on the GenDecl doc (single spec) or the TypeSpec doc.
		_, _, isRoot, _, err := docDirectives(d.Doc)
		if err != nil {
			return err
		}
		if !isRoot {
			_, _, isRoot, _, err = docDirectives(ts.Doc)
			if err != nil {
				return err
			}
		}
		if !isRoot {
			continue
		}
		obj := p.TypesInfo.Defs[ts.Name]
		if obj == nil {
			return fmt.Errorf("di:root %s: no type info", ts.Name.Name)
		}
		res.Roots = append(res.Roots, &Root{Name: ts.Name.Name, Type: obj.Type()})
	}
	return nil
}

func scanProvider(p *packages.Package, fn *ast.FuncDecl, lineDir map[int]annotation.Directive, res *Result) error {
	provide, name, _, expose, err := docDirectives(fn.Doc)
	if err != nil {
		return fmt.Errorf("%s: %w", fn.Name.Name, err)
	}
	if !provide {
		return nil
	}
	obj := p.TypesInfo.Defs[fn.Name]
	if obj == nil {
		return fmt.Errorf("provider %s: no type info", fn.Name.Name)
	}
	sig := obj.Type().(*types.Signature)

	produces, returnsErr, err := producedType(fn.Name.Name, sig)
	if err != nil {
		return err
	}

	prov := &Provider{
		Pkg:          p,
		FuncName:     fn.Name.Name,
		Name:         cmp.Or(name, fn.Name.Name),
		Produces:     produces,
		ReturnsError: returnsErr,
		Expose:       expose,
	}

	for _, field := range fn.Type.Params.List {
		var pth *parampath.Path
		var use string
		line := p.Fset.Position(field.Pos()).Line

		d1, ok1 := lineDir[line-1]
		d2, ok2 := lineDir[line-2]

		// Detect mutual exclusion: di:param and di:use stacked above the param.
		kinds := map[annotation.Kind]bool{}
		if ok1 {
			kinds[d1.Kind] = true
		}
		if ok2 {
			kinds[d2.Kind] = true
		}
		if kinds[annotation.Param] && kinds[annotation.Use] {
			return fmt.Errorf("%s: parameter on line %d: di:param and di:use are mutually exclusive", fn.Name.Name, line)
		}

		// Apply the directive on the line immediately above the param.
		if ok1 {
			switch d1.Kind {
			case annotation.Param:
				parsed, perr := parampath.Parse(d1.Path)
				if perr != nil {
					return fmt.Errorf("%s: %w", fn.Name.Name, perr)
				}
				pth = &parsed
			case annotation.Use:
				use = d1.Node
			}
		}

		// One field may declare multiple names sharing a type.
		names := field.Names
		if len(names) == 0 {
			names = []*ast.Ident{ast.NewIdent("_")}
		}
		for _, nm := range names {
			ft := p.TypesInfo.TypeOf(field.Type)
			if ft == nil {
				return fmt.Errorf("%s: param %s has no type", fn.Name.Name, nm.Name)
			}
			prov.Params = append(prov.Params, Param{Name: nm.Name, Type: ft, Path: pth, Use: use})
		}
	}

	res.Providers = append(res.Providers, prov)
	return nil
}

func producedType(fnName string, sig *types.Signature) (types.Type, bool, error) {
	r := sig.Results()
	errType := types.Universe.Lookup("error").Type()
	switch r.Len() {
	case 1:
		if types.Identical(r.At(0).Type(), errType) {
			return nil, false, fmt.Errorf("provider %s returns only error", fnName)
		}
		return r.At(0).Type(), false, nil
	case 2:
		if !types.Identical(r.At(1).Type(), errType) {
			return nil, false, fmt.Errorf("provider %s: second result must be error", fnName)
		}
		return r.At(0).Type(), true, nil
	default:
		return nil, false, fmt.Errorf("provider %s must return (T) or (T, error)", fnName)
	}
}
