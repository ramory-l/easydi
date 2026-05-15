package gen

import (
	"os"
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
