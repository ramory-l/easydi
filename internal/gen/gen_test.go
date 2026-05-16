package gen

import (
	"go/format"
	"os"
	"strings"
	"testing"

	"github.com/ramory-l/easydi/internal/loader"
	"github.com/ramory-l/easydi/internal/resolver"
	"github.com/ramory-l/easydi/internal/scanner"
	"github.com/ramory-l/easydi/internal/topo"
)

func TestGenerateGolden(t *testing.T) {
	pkgs, _ := loader.Load("../testdata/simple")
	res, _ := scanner.Scan(pkgs)
	g, err := resolver.Resolve(res)
	if err != nil {
		t.Fatal(err)
	}
	order, err := topo.Sort(g)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Generate(g, order, "diout")
	if err != nil {
		t.Fatal(err)
	}

	const goldenPath = "testdata/simple.golden"
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(goldenPath, out, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(want) {
		t.Fatalf("generated output mismatch.\n--- got ---\n%s", out)
	}
}

func TestGenerateEmitsLifecycle(t *testing.T) {
	pkgs, err := loader.Load("../testdata/simple")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	g, err := resolver.Resolve(res)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	order, err := topo.Sort(g)
	if err != nil {
		t.Fatalf("sort: %v", err)
	}
	out, err := Generate(g, order, "diout")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	s := string(out)
	for _, want := range []string{
		`"context"`,
		`"errors"`,
		`"github.com/ramory-l/easydi/lifecycle"`,
		"func (c *Container) Start(ctx context.Context) error {",
		"func (c *Container) Close(ctx context.Context) error {",
		"nodes := c.Exposed()",
		"n.(lifecycle.Starter)",
		"nodes[i].(lifecycle.Closer)",
		"return errors.Join(errs...)",
		"for j := i - 1; j >= 0; j--",
		"_ = cl.Close(ctx)",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("generated output missing %q\n---\n%s", want, s)
		}
	}
}

func TestGenerateAliasRootEmitsRHSNoSelfImport(t *testing.T) {
	pkgs, err := loader.Load("../testdata/aliasroot/...")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	g, err := resolver.Resolve(res)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	order, err := topo.Sort(g)
	if err != nil {
		t.Fatalf("sort: %v", err)
	}
	out, err := Generate(g, order, "di")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	s := string(out)
	// Must NOT import its own output package (the aliasroot/di package path).
	if strings.Contains(s, `testdata/aliasroot/di"`) {
		t.Fatalf("generated file imports its own package (self-import):\n%s", s)
	}
	// Build must take the alias RHS *cfg.Config, not the alias name di.Config.
	if !strings.Contains(s, "func Build(config *cfg.Config)") {
		t.Fatalf("Build must take the alias RHS *cfg.Config, got:\n%s", s)
	}
	if _, ferr := format.Source(out); ferr != nil {
		t.Fatalf("generated source invalid: %v\n%s", ferr, s)
	}
}

func TestGenerateDedupesCollidingImports(t *testing.T) {
	pkgs, err := loader.Load("../testdata/collide/...")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	g, err := resolver.Resolve(res)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	order, err := topo.Sort(g)
	if err != nil {
		t.Fatalf("sort: %v", err)
	}
	out, err := Generate(g, order, "diout")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	s := string(out)
	// Two packages named "svc" must NOT both be imported as bare "svc".
	// computeAliases produces lowerCamelCase suffix-based aliases (e.g. aSvc,
	// bSvc) so we check that at least one import line carries an explicit alias
	// for one of the collide/*/svc paths.
	if strings.Count(s, `"github.com/ramory-l/easydi/internal/testdata/collide/`) < 2 {
		t.Fatalf("expected both collide svc imports present, got:\n%s", s)
	}
	if strings.Contains(s, "\t\"github.com/ramory-l/easydi/internal/testdata/collide/a/svc\"") &&
		strings.Contains(s, "\t\"github.com/ramory-l/easydi/internal/testdata/collide/b/svc\"") {
		// Both imported as bare "svc" — collision not resolved.
		t.Fatalf("both svc packages imported without alias (collision not resolved):\n%s", s)
	}
	// The generated file must be valid Go (parses + gofmt-stable).
	if _, ferr := format.Source(out); ferr != nil {
		t.Fatalf("generated source not valid Go: %v\n%s", ferr, s)
	}
	// The body must use the aliased qualifiers, not a bare ambiguous "svc.".
	if !strings.Contains(s, "aSvc.") || !strings.Contains(s, "bSvc.") {
		t.Fatalf("expected aliased qualifiers (aSvc./bSvc.) in body, got:\n%s", s)
	}
	// Determinism: regenerate, expect identical bytes.
	out2, err := Generate(g, order, "diout")
	if err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if string(out2) != s {
		t.Fatalf("generation not deterministic")
	}
}

func TestGenerateV3EndToEndCompiles(t *testing.T) {
	pkgs, err := loader.Load("../testdata/v3e2e/...")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	g, err := resolver.Resolve(res)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	order, err := topo.Sort(g)
	if err != nil {
		t.Fatalf("sort: %v", err)
	}
	out, err := Generate(g, order, "di")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	s := string(out)

	// Feature B: no self-import of the output di package.
	if strings.Contains(s, `testdata/v3e2e/di"`) {
		t.Fatalf("self-import present:\n%s", s)
	}
	// Feature B: Build must take the alias RHS *cfg.Config, not the alias name di.Config.
	if !strings.Contains(s, "func Build(config *cfg.Config)") {
		t.Fatalf("alias root RHS missing from Build signature:\n%s", s)
	}
	// Feature C: di:use Svc must wire c.Svc into NewWidget, not c.Repo.
	// We check the Widget call expression directly rather than fragile suffix
	// matching that could fire on NewSvc's argument list.
	if !strings.Contains(s, "NewWidget(c.Svc)") {
		t.Fatalf("di:use must bind Svc: expected NewWidget(c.Svc) in output:\n%s", s)
	}
	if strings.Contains(s, "NewWidget(c.Repo)") {
		t.Fatalf("di:use must NOT bind Repo: found NewWidget(c.Repo) in output:\n%s", s)
	}
	// Feature A: two packages named "store" must be aliased (no bare collision).
	if !strings.Contains(s, "aStore.") || !strings.Contains(s, "bStore.") {
		t.Fatalf("expected aliased store qualifiers (aStore./bStore.) in output:\n%s", s)
	}
	// Output must be valid gofmt'd Go.
	if _, ferr := format.Source(out); ferr != nil {
		t.Fatalf("generated invalid Go: %v\n%s", ferr, s)
	}
}
