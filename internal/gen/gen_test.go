package gen

import (
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
