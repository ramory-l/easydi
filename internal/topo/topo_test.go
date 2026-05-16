package topo

import (
	"strings"
	"testing"

	"github.com/ramory-l/easydi/internal/loader"
	"github.com/ramory-l/easydi/internal/resolver"
	"github.com/ramory-l/easydi/internal/scanner"
)

func TestSortOrder(t *testing.T) {
	pkgs, _ := loader.Load("../testdata/simple")
	res, _ := scanner.Scan(pkgs)
	g, err := resolver.Resolve(res)
	if err != nil {
		t.Fatal(err)
	}
	order, err := Sort(g)
	if err != nil {
		t.Fatal(err)
	}
	pos := map[string]int{}
	for i, n := range order {
		pos[n.Name()] = i
	}
	if !(pos["NewHasher"] < pos["NewRepo"] && pos["NewRepo"] < pos["NewService"]) {
		t.Fatalf("bad order: %v", pos)
	}
}

func TestSortCycle(t *testing.T) {
	pkgs, _ := loader.Load("../testdata/cycle")
	res, _ := scanner.Scan(pkgs)
	g, err := resolver.Resolve(res)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Sort(g); err == nil {
		t.Fatal("expected cycle error")
	}
}

// TestDiUseCycle verifies that a dependency cycle introduced via di:use is
// accepted by resolver.Resolve (the edges bind fine) but detected by Sort.
func TestDiUseCycle(t *testing.T) {
	pkgs, err := loader.Load("../testdata/diusecycle")
	if err != nil {
		t.Fatal(err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatal(err)
	}
	g, err := resolver.Resolve(res)
	if err != nil {
		t.Fatalf("resolver.Resolve must succeed for a cycle graph: %v", err)
	}
	_, err = Sort(g)
	if err == nil {
		t.Fatal("Sort: expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("Sort error %q does not contain \"dependency cycle\"", err.Error())
	}
}
