package topo

import (
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
