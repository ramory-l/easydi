// Package loader wraps go/packages with the modes easydi needs.
package loader

import (
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

func Load(patterns ...string) ([]*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{Mode: mode}, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	var errs int
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		errs += len(p.Errors)
		for _, e := range p.Errors {
			err = fmt.Errorf("%s: %s", p.PkgPath, e)
		}
	})
	if errs > 0 {
		return nil, fmt.Errorf("packages had %d errors: %w", errs, err)
	}
	return pkgs, nil
}
