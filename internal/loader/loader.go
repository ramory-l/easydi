// Package loader wraps go/packages with the modes easydi needs.
package loader

import (
	"errors"
	"fmt"

	"golang.org/x/tools/go/packages"
)

const mode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo

// Load loads the packages matching patterns with full type and syntax
// information, returning a joined error if any loaded package has errors.
func Load(patterns ...string) ([]*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{Mode: mode}, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	var errs []error
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		for _, e := range p.Errors {
			errs = append(errs, fmt.Errorf("%s: %s", p.PkgPath, e))
		}
	})
	if len(errs) > 0 {
		return nil, fmt.Errorf("packages had %d errors: %w", len(errs), errors.Join(errs...))
	}
	return pkgs, nil
}
