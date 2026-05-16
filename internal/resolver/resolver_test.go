package resolver

import (
	"go/types"
	"strings"
	"testing"

	"github.com/ramory-l/easydi/internal/loader"
	"github.com/ramory-l/easydi/internal/scanner"
)

func resolveFixture(t *testing.T) *Graph {
	t.Helper()
	pkgs, err := loader.Load("../testdata/simple")
	if err != nil {
		t.Fatal(err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatal(err)
	}
	g, err := Resolve(res)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestResolveBindings(t *testing.T) {
	g := resolveFixture(t)

	n := g.NodeByName("NewService")
	if n == nil {
		t.Fatal("NewService node missing")
	}
	// r Repo -> provider NewRepo ; sec -> Config root projection
	if got := n.Bindings[0]; got.FromNode == nil || got.FromNode.Name() != "NewRepo" {
		t.Fatalf("binding[0]=%+v", got)
	}
	if got := n.Bindings[1]; got.Expr != "config.Auth.Secret" {
		t.Fatalf("binding[1].Expr=%q want config.Auth.Secret", got.Expr)
	}

	h := g.NodeByName("NewHasher")
	if got := h.Bindings[0].Expr; got != "infra.DB.DSN()" {
		t.Fatalf("NewHasher dsn expr=%q want infra.DB.DSN()", got)
	}

	repo := g.NodeByName("NewRepo")
	if repo.Bindings[0].FromNode == nil || repo.Bindings[0].FromNode.Name() != "NewHasher" {
		t.Fatalf("NewRepo should depend on NewHasher")
	}
}

func TestResolveRoots(t *testing.T) {
	g := resolveFixture(t)
	if len(g.Roots) != 2 {
		t.Fatalf("roots=%d", len(g.Roots))
	}
	// root var names are the lowercased root type names
	names := map[string]bool{}
	for _, r := range g.Roots {
		names[r.VarName] = true
	}
	if !names["config"] || !names["infra"] {
		t.Fatalf("root var names=%v", names)
	}
}

func TestSatisfiesStrict(t *testing.T) {
	intT := types.Typ[types.Int]
	ptrInt := types.NewPointer(intT)
	if satisfies(intT, ptrInt) {
		t.Fatal("int must not satisfy *int")
	}
	if satisfies(ptrInt, intT) {
		t.Fatal("*int must not satisfy int")
	}
	if !satisfies(intT, intT) {
		t.Fatal("int must satisfy int")
	}
}

// TestDiUseHappy verifies that di:use selects the named node when two providers
// both satisfy the parameter interface (which would otherwise be ambiguous).
func TestDiUseHappy(t *testing.T) {
	pkgs, err := loader.Load("../testdata/diuseambig")
	if err != nil {
		t.Fatal(err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatal(err)
	}
	g, err := Resolve(res)
	if err != nil {
		t.Fatalf("Resolve returned error (expected success): %v", err)
	}
	consumer := g.NodeByName("Consumer")
	if consumer == nil {
		t.Fatal("Consumer node missing")
	}
	if len(consumer.Bindings) == 0 {
		t.Fatal("Consumer has no bindings")
	}
	b := consumer.Bindings[0]
	if b.FromNode == nil {
		t.Fatalf("binding[0].FromNode is nil, got Expr=%q", b.Expr)
	}
	if got := b.FromNode.Name(); got != "UserService" {
		t.Fatalf("binding[0].FromNode.Name()=%q, want UserService", got)
	}
}

// TestDiUseUnknownNode verifies that di:use with a non-existent node name
// returns an error containing "no provider node named".
func TestDiUseUnknownNode(t *testing.T) {
	pkgs, err := loader.Load("../testdata/diusenone")
	if err != nil {
		t.Fatal(err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Resolve(res)
	if err == nil {
		t.Fatal("expected error for unknown di:use node, got nil")
	}
	if !strings.Contains(err.Error(), "no provider node named") {
		t.Fatalf("error %q does not contain \"no provider node named\"", err.Error())
	}
}

// TestDiUseNotAssignable verifies that di:use with a node whose produced type
// is not assignable to the parameter type returns an error containing
// "not assignable to parameter".
func TestDiUseNotAssignable(t *testing.T) {
	pkgs, err := loader.Load("../testdata/diusebad")
	if err != nil {
		t.Fatal(err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Resolve(res)
	if err == nil {
		t.Fatal("expected error for not-assignable di:use, got nil")
	}
	if !strings.Contains(err.Error(), "not assignable to parameter") {
		t.Fatalf("error %q does not contain \"not assignable to parameter\"", err.Error())
	}
}

func TestResolveWholeRoot(t *testing.T) {
	pkgs, err := loader.Load("../testdata/wholeroot")
	if err != nil {
		t.Fatal(err)
	}
	res, err := scanner.Scan(pkgs)
	if err != nil {
		t.Fatal(err)
	}
	g, err := Resolve(res)
	if err != nil {
		t.Fatal(err)
	}
	n := g.NodeByName("NewThing")
	if n == nil {
		t.Fatal("NewThing node missing")
	}
	if got := n.Bindings[0].Expr; got != "config" {
		t.Fatalf("whole-root expr=%q want %q", got, "config")
	}
}
