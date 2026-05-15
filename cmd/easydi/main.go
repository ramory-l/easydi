// Command easydi generates a compile-time DI container from // di: annotations.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ramory-l/easydi/internal/gen"
	"github.com/ramory-l/easydi/internal/loader"
	"github.com/ramory-l/easydi/internal/resolver"
	"github.com/ramory-l/easydi/internal/scanner"
	"github.com/ramory-l/easydi/internal/topo"
)

// defaultOut is the generated file name when -o is omitted.
const defaultOut = "easydi_gen.go"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "easydi:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] != "gen" {
		return fmt.Errorf("usage: easydi gen [-o file] [-pkg name] <patterns...>")
	}
	fs := flag.NewFlagSet("gen", flag.ContinueOnError)
	outPath := fs.String("o", defaultOut, "output file path")
	pkgName := fs.String("pkg", "", "package name for the generated file")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	patterns := fs.Args()
	if len(patterns) == 0 {
		return fmt.Errorf("no package patterns given")
	}
	if *pkgName == "" {
		return fmt.Errorf("-pkg is required")
	}

	pkgs, err := loader.Load(patterns...)
	if err != nil {
		return err
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		return err
	}
	g, err := resolver.Resolve(res)
	if err != nil {
		return err
	}
	order, err := topo.Sort(g)
	if err != nil {
		return err
	}
	src, err := gen.Generate(g, order, *pkgName)
	if err != nil {
		return err
	}
	if dir := filepath.Dir(*outPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}
	}
	return os.WriteFile(*outPath, src, 0o644)
}
